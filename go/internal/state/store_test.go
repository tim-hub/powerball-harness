package state_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Chachamaru127/claude-code-harness/go/internal/state"
)

// ============================================================
// テストヘルパー
// ============================================================

// newTestStore はテスト用の一時 DB ファイルを作成し、HarnessStore を返す。
// t.Cleanup で自動的にクローズ・削除される。
func newTestStore(t *testing.T) *state.HarnessStore {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	store, err := state.NewHarnessStore(dbPath)
	if err != nil {
		t.Fatalf("NewHarnessStore: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := store.Close(); closeErr != nil {
			t.Logf("store.Close: %v", closeErr)
		}
	})
	return store
}

// nowISO は現在時刻を ISO 8601 UTC 文字列で返す。
func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ============================================================
// スキーマ初期化テスト
// ============================================================

func TestSchemaInit(t *testing.T) {
	store := newTestStore(t)

	// schema_meta にバージョンが記録されていること
	version, err := store.GetMeta("version")
	if err != nil {
		t.Fatalf("GetMeta: %v", err)
	}
	if version == "" {
		t.Fatal("schema version not recorded")
	}
	if version != fmt.Sprintf("%d", state.SchemaVersion) {
		t.Errorf("version = %q, want %q", version, fmt.Sprintf("%d", state.SchemaVersion))
	}
}

func TestSchemaInit_Idempotent(t *testing.T) {
	// 同じパスで 2 回 NewHarnessStore を呼んでも問題ないこと
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "idempotent.db")

	store1, err := state.NewHarnessStore(dbPath)
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	store1.Close()

	store2, err := state.NewHarnessStore(dbPath)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	defer store2.Close()

	version, err := store2.GetMeta("version")
	if err != nil || version == "" {
		t.Fatalf("version after re-open: err=%v, version=%q", err, version)
	}
}

// ============================================================
// セッション管理テスト
// ============================================================

func TestSession_UpsertAndGet(t *testing.T) {
	store := newTestStore(t)

	session := state.SessionState{
		SessionID:   "sess-001",
		Mode:        state.SessionModeWork,
		ProjectRoot: "/tmp/project",
		StartedAt:   nowISO(),
		Context:     map[string]interface{}{"key": "value"},
	}

	if err := store.UpsertSession(session); err != nil {
		t.Fatalf("UpsertSession: %v", err)
	}

	got, err := store.GetSession("sess-001")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got == nil {
		t.Fatal("expected session, got nil")
	}

	if got.SessionID != "sess-001" {
		t.Errorf("SessionID = %q, want %q", got.SessionID, "sess-001")
	}
	if got.Mode != state.SessionModeWork {
		t.Errorf("Mode = %q, want %q", got.Mode, state.SessionModeWork)
	}
	if got.ProjectRoot != "/tmp/project" {
		t.Errorf("ProjectRoot = %q, want %q", got.ProjectRoot, "/tmp/project")
	}
	if got.Context["key"] != "value" {
		t.Errorf("Context[key] = %v, want %q", got.Context["key"], "value")
	}
}

func TestSession_UpsertUpdatesExisting(t *testing.T) {
	store := newTestStore(t)

	session := state.SessionState{
		SessionID:   "sess-update",
		Mode:        state.SessionModeNormal,
		ProjectRoot: "/tmp/project",
		StartedAt:   nowISO(),
	}
	if err := store.UpsertSession(session); err != nil {
		t.Fatalf("UpsertSession (insert): %v", err)
	}

	// モードを更新
	session.Mode = state.SessionModeBreezing
	if err := store.UpsertSession(session); err != nil {
		t.Fatalf("UpsertSession (update): %v", err)
	}

	got, err := store.GetSession("sess-update")
	if err != nil || got == nil {
		t.Fatalf("GetSession: err=%v, got=%v", err, got)
	}
	if got.Mode != state.SessionModeBreezing {
		t.Errorf("Mode = %q, want breezing", got.Mode)
	}
}

func TestSession_GetNotFound(t *testing.T) {
	store := newTestStore(t)

	got, err := store.GetSession("nonexistent")
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestSession_EndSession(t *testing.T) {
	store := newTestStore(t)

	session := state.SessionState{
		SessionID:   "sess-end",
		Mode:        state.SessionModeWork,
		ProjectRoot: "/tmp/project",
		StartedAt:   nowISO(),
	}
	if err := store.UpsertSession(session); err != nil {
		t.Fatalf("UpsertSession: %v", err)
	}

	if err := store.EndSession("sess-end"); err != nil {
		t.Fatalf("EndSession: %v", err)
	}

	got, err := store.GetSession("sess-end")
	if err != nil || got == nil {
		t.Fatalf("GetSession after end: err=%v, got=%v", err, got)
	}
	if got.EndedAt == nil {
		t.Error("expected EndedAt to be set")
	}
}

// ============================================================
// シグナル管理テスト
// ============================================================

func TestSignal_SendAndReceive(t *testing.T) {
	store := newTestStore(t)

	// 前提: セッションを先に登録する（FK 制約のため sessions への INSERT は不要だが
	// work_states には必要。signals は直接 from_session_id を持つため sessions 不要）

	toSession := "sess-receiver"
	sig := state.Signal{
		Type:          "test.event",
		FromSessionID: "sess-sender",
		ToSessionID:   &toSession,
		Payload:       map[string]interface{}{"data": 42},
	}

	id, err := store.SendSignal(sig)
	if err != nil {
		t.Fatalf("SendSignal: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}

	// receiver が受信する
	received, err := store.ReceiveSignals("sess-receiver")
	if err != nil {
		t.Fatalf("ReceiveSignals: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(received))
	}
	if received[0].Type != "test.event" {
		t.Errorf("Type = %q, want %q", received[0].Type, "test.event")
	}
	if received[0].Payload["data"] != float64(42) {
		t.Errorf("Payload[data] = %v, want 42", received[0].Payload["data"])
	}

	// 2 回目は消費済みなので空になること
	received2, err := store.ReceiveSignals("sess-receiver")
	if err != nil {
		t.Fatalf("ReceiveSignals (2nd): %v", err)
	}
	if len(received2) != 0 {
		t.Errorf("expected 0 signals after consume, got %d", len(received2))
	}
}

func TestSignal_Broadcast(t *testing.T) {
	store := newTestStore(t)

	// ブロードキャスト（to_session_id = nil）
	sig := state.Signal{
		Type:          "broadcast.event",
		FromSessionID: "sess-broadcaster",
		ToSessionID:   nil,
		Payload:       map[string]interface{}{"msg": "hello"},
	}

	if _, err := store.SendSignal(sig); err != nil {
		t.Fatalf("SendSignal broadcast: %v", err)
	}

	// どのセッションでも受信できること（送信者以外）
	received, err := store.ReceiveSignals("sess-anyone")
	if err != nil {
		t.Fatalf("ReceiveSignals: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 broadcast signal, got %d", len(received))
	}
}

func TestSignal_SenderDoesNotReceiveOwn(t *testing.T) {
	store := newTestStore(t)

	sig := state.Signal{
		Type:          "own.event",
		FromSessionID: "sess-self",
		ToSessionID:   nil,
		Payload:       map[string]interface{}{},
	}
	if _, err := store.SendSignal(sig); err != nil {
		t.Fatalf("SendSignal: %v", err)
	}

	// 送信者自身は受信しないこと
	received, err := store.ReceiveSignals("sess-self")
	if err != nil {
		t.Fatalf("ReceiveSignals: %v", err)
	}
	if len(received) != 0 {
		t.Errorf("sender should not receive own signal, got %d", len(received))
	}
}

// ============================================================
// タスク失敗管理テスト
// ============================================================

func TestTaskFailure_RecordAndGet(t *testing.T) {
	store := newTestStore(t)

	detail := "stack trace here"
	failure := state.TaskFailure{
		TaskID:   "task-001",
		Severity: "error",
		Message:  "something went wrong",
		Detail:   &detail,
		Attempt:  1,
	}

	id, err := store.RecordFailure(failure, "sess-worker")
	if err != nil {
		t.Fatalf("RecordFailure: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}

	failures, err := store.GetFailures("task-001")
	if err != nil {
		t.Fatalf("GetFailures: %v", err)
	}
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure, got %d", len(failures))
	}

	f := failures[0]
	if f.TaskID != "task-001" {
		t.Errorf("TaskID = %q, want %q", f.TaskID, "task-001")
	}
	if f.Severity != "error" {
		t.Errorf("Severity = %q, want error", f.Severity)
	}
	if f.Message != "something went wrong" {
		t.Errorf("Message = %q, want ...", f.Message)
	}
	if f.Detail == nil || *f.Detail != "stack trace here" {
		t.Errorf("Detail = %v, want %q", f.Detail, "stack trace here")
	}
	if f.Attempt != 1 {
		t.Errorf("Attempt = %d, want 1", f.Attempt)
	}
}

func TestTaskFailure_MultipleAttempts(t *testing.T) {
	store := newTestStore(t)

	for i := 1; i <= 3; i++ {
		f := state.TaskFailure{
			TaskID:   "task-multi",
			Severity: "warning",
			Message:  fmt.Sprintf("attempt %d failed", i),
			Attempt:  i,
		}
		if _, err := store.RecordFailure(f, "sess-x"); err != nil {
			t.Fatalf("RecordFailure attempt %d: %v", i, err)
		}
	}

	failures, err := store.GetFailures("task-multi")
	if err != nil {
		t.Fatalf("GetFailures: %v", err)
	}
	if len(failures) != 3 {
		t.Fatalf("expected 3 failures, got %d", len(failures))
	}
	// 時刻順に並んでいること
	if failures[0].Attempt != 1 || failures[2].Attempt != 3 {
		t.Errorf("failures not ordered by attempt: %v", failures)
	}
}

func TestTaskFailure_GetNotFound(t *testing.T) {
	store := newTestStore(t)

	failures, err := store.GetFailures("nonexistent-task")
	if err != nil {
		t.Fatalf("GetFailures: %v", err)
	}
	if len(failures) != 0 {
		t.Errorf("expected empty slice, got %d", len(failures))
	}
}

// ============================================================
// work_states 管理テスト
// ============================================================

func TestWorkState_SetAndGet(t *testing.T) {
	store := newTestStore(t)

	// sessions テーブルへ先に INSERT（FK 制約があるため）
	if err := store.UpsertSession(state.SessionState{
		SessionID:   "sess-work",
		Mode:        state.SessionModeWork,
		ProjectRoot: "/tmp",
		StartedAt:   nowISO(),
	}); err != nil {
		t.Fatalf("UpsertSession: %v", err)
	}

	opts := state.WorkStateOptions{
		CodexMode:     true,
		BypassRmRf:    false,
		BypassGitPush: true,
		WorkMode:      true,
	}
	if err := store.SetWorkState("sess-work", opts); err != nil {
		t.Fatalf("SetWorkState: %v", err)
	}

	ws, err := store.GetWorkState("sess-work")
	if err != nil {
		t.Fatalf("GetWorkState: %v", err)
	}
	if ws == nil {
		t.Fatal("expected WorkState, got nil")
	}

	if !ws.CodexMode {
		t.Error("CodexMode should be true")
	}
	if ws.BypassRmRf {
		t.Error("BypassRmRf should be false")
	}
	if !ws.BypassGitPush {
		t.Error("BypassGitPush should be true")
	}
	if !ws.WorkMode {
		t.Error("WorkMode should be true")
	}
}

func TestWorkState_GetNotFound(t *testing.T) {
	store := newTestStore(t)

	ws, err := store.GetWorkState("nonexistent-session")
	if err != nil {
		t.Fatalf("GetWorkState: %v", err)
	}
	if ws != nil {
		t.Errorf("expected nil for nonexistent session, got %+v", ws)
	}
}

func TestWorkState_Upsert(t *testing.T) {
	store := newTestStore(t)

	if err := store.UpsertSession(state.SessionState{
		SessionID:   "sess-upsert",
		Mode:        state.SessionModeWork,
		ProjectRoot: "/tmp",
		StartedAt:   nowISO(),
	}); err != nil {
		t.Fatalf("UpsertSession: %v", err)
	}

	// 初回設定
	if err := store.SetWorkState("sess-upsert", state.WorkStateOptions{CodexMode: false}); err != nil {
		t.Fatalf("SetWorkState (1st): %v", err)
	}

	// 更新
	if err := store.SetWorkState("sess-upsert", state.WorkStateOptions{CodexMode: true}); err != nil {
		t.Fatalf("SetWorkState (2nd): %v", err)
	}

	ws, err := store.GetWorkState("sess-upsert")
	if err != nil || ws == nil {
		t.Fatalf("GetWorkState: err=%v, ws=%v", err, ws)
	}
	if !ws.CodexMode {
		t.Error("CodexMode should be true after upsert")
	}
}

func TestWorkState_CleanExpired(t *testing.T) {
	store := newTestStore(t)

	// CleanExpiredWorkStates はタイムスタンプを直接操作できないため、
	// 空の DB で呼び出して 0 件を確認する（ロジックの動作検証）
	n, err := store.CleanExpiredWorkStates()
	if err != nil {
		t.Fatalf("CleanExpiredWorkStates: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 deletions in empty db, got %d", n)
	}
}

// ============================================================
// schema_meta キー/バリュー管理テスト
// ============================================================

func TestMeta_SetAndGet(t *testing.T) {
	store := newTestStore(t)

	if err := store.SetMeta("test_key", "test_value"); err != nil {
		t.Fatalf("SetMeta: %v", err)
	}

	value, err := store.GetMeta("test_key")
	if err != nil {
		t.Fatalf("GetMeta: %v", err)
	}
	if value != "test_value" {
		t.Errorf("GetMeta = %q, want %q", value, "test_value")
	}
}

func TestMeta_Upsert(t *testing.T) {
	store := newTestStore(t)

	if err := store.SetMeta("key", "v1"); err != nil {
		t.Fatalf("SetMeta v1: %v", err)
	}
	if err := store.SetMeta("key", "v2"); err != nil {
		t.Fatalf("SetMeta v2: %v", err)
	}

	value, err := store.GetMeta("key")
	if err != nil || value != "v2" {
		t.Errorf("GetMeta = %q, err=%v; want v2", value, err)
	}
}

func TestMeta_GetNotFound(t *testing.T) {
	store := newTestStore(t)

	value, err := store.GetMeta("nonexistent_key")
	if err != nil {
		t.Fatalf("GetMeta: %v", err)
	}
	if value != "" {
		t.Errorf("expected empty string, got %q", value)
	}
}

// ============================================================
// assumptions テスト
// ============================================================

func TestAssumption_RecordAndGet(t *testing.T) {
	store := newTestStore(t)

	taskID := "task-abc"
	a := state.Assumption{
		SessionID:  "sess-agent",
		TaskID:     &taskID,
		Assumption: "The user wants a REST API",
		Confidence: 0.9,
	}

	id, err := store.RecordAssumption(a)
	if err != nil {
		t.Fatalf("RecordAssumption: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}

	assumptions, err := store.GetAssumptions("sess-agent")
	if err != nil {
		t.Fatalf("GetAssumptions: %v", err)
	}
	if len(assumptions) != 1 {
		t.Fatalf("expected 1 assumption, got %d", len(assumptions))
	}

	got := assumptions[0]
	if got.Assumption != "The user wants a REST API" {
		t.Errorf("Assumption = %q", got.Assumption)
	}
	if got.Confidence != 0.9 {
		t.Errorf("Confidence = %f, want 0.9", got.Confidence)
	}
	if got.TaskID == nil || *got.TaskID != "task-abc" {
		t.Errorf("TaskID = %v, want %q", got.TaskID, "task-abc")
	}
}

// ============================================================
// 並列アクセステスト
// ============================================================

// TestConcurrentAccess は 3 goroutine が並列で INSERT/SELECT しても
// デッドロックが発生しないことを確認する。
func TestConcurrentAccess(t *testing.T) {
	store := newTestStore(t)

	const goroutines = 3
	const opsPerGoroutine = 10

	var wg sync.WaitGroup
	errCh := make(chan error, goroutines*opsPerGoroutine)

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(gID int) {
			defer wg.Done()

			for i := 0; i < opsPerGoroutine; i++ {
				// INSERT
				sig := state.Signal{
					Type:          fmt.Sprintf("concurrent.event.%d.%d", gID, i),
					FromSessionID: fmt.Sprintf("sess-g%d", gID),
					ToSessionID:   nil,
					Payload:       map[string]interface{}{"gid": gID, "i": i},
				}
				if _, err := store.SendSignal(sig); err != nil {
					errCh <- fmt.Errorf("g%d SendSignal %d: %w", gID, i, err)
					return
				}

				// SELECT（自分以外が送信したもの）
				receiver := fmt.Sprintf("sess-receiver-%d-%d", gID, i)
				if _, err := store.ReceiveSignals(receiver); err != nil {
					errCh <- fmt.Errorf("g%d ReceiveSignals %d: %w", gID, i, err)
					return
				}
			}
		}(g)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Error(err)
		}
	}
}

// ============================================================
// パスと DB ファイル作成テスト
// ============================================================

func TestNewHarnessStore_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	// 存在しないサブディレクトリの下に DB を作成する
	dbPath := filepath.Join(dir, "sub", "deep", "state.db")

	store, err := state.NewHarnessStore(dbPath)
	if err != nil {
		t.Fatalf("NewHarnessStore: %v", err)
	}
	defer store.Close()

	if _, statErr := os.Stat(dbPath); os.IsNotExist(statErr) {
		t.Errorf("expected DB file to be created at %s", dbPath)
	}
}

func TestResolveStatePath(t *testing.T) {
	// 環境変数を設定して優先順位を確認する
	t.Run("CLAUDE_PLUGIN_DATA takes priority", func(t *testing.T) {
		t.Setenv("CLAUDE_PLUGIN_DATA", "/plugin/data")

		path := state.ResolveStatePath("/project")
		if path != "/plugin/data/state.db" {
			t.Errorf("path = %q, want /plugin/data/state.db", path)
		}
	})

	t.Run("ProjectRoot fallback", func(t *testing.T) {
		t.Setenv("CLAUDE_PLUGIN_DATA", "")

		path := state.ResolveStatePath("/my/project")
		if path != "/my/project/.harness/state.db" {
			t.Errorf("path = %q, want /my/project/.harness/state.db", path)
		}
	})
}
