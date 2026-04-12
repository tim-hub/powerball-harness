package lifecycle_test

import (
	"testing"

	"github.com/Chachamaru127/claude-code-harness/go/internal/lifecycle"
)

// ============================================================
// IsTerminal
// ============================================================

func TestIsTerminal_CommittedIsTerminal(t *testing.T) {
	if !lifecycle.IsTerminal(lifecycle.StateCommitted) {
		t.Error("COMMITTED は終端状態であるべき")
	}
}

func TestIsTerminal_AbortedIsTerminal(t *testing.T) {
	if !lifecycle.IsTerminal(lifecycle.StateAborted) {
		t.Error("ABORTED は終端状態であるべき")
	}
}

func TestIsTerminal_CancelledIsTerminal(t *testing.T) {
	if !lifecycle.IsTerminal(lifecycle.StateCancelled) {
		t.Error("CANCELLED は終端状態であるべき")
	}
}

func TestIsTerminal_NonTerminalStates(t *testing.T) {
	nonTerminal := []lifecycle.AgentState{
		lifecycle.StateSpawning,
		lifecycle.StateRunning,
		lifecycle.StateReviewing,
		lifecycle.StateApproved,
		lifecycle.StateFailed,
		lifecycle.StateStale,
		lifecycle.StateRecovering,
	}
	for _, s := range nonTerminal {
		if lifecycle.IsTerminal(s) {
			t.Errorf("状態 %s は終端状態であってはならない", s)
		}
	}
}

// ============================================================
// NewStateMachine / Current
// ============================================================

func TestNewStateMachine_StartsAtSpawning(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	if sm.Current() != lifecycle.StateSpawning {
		t.Errorf("初期状態は SPAWNING であるべきだが %s だった", sm.Current())
	}
}

func TestNewStateMachine_HistoryIsEmpty(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	if len(sm.History()) != 0 {
		t.Errorf("初期状態では履歴が空であるべきだが %d 件あった", len(sm.History()))
	}
}

// ============================================================
// 正常系全遷移パス
// ============================================================

func TestHappyPath_FullTransition(t *testing.T) {
	sm := lifecycle.NewStateMachine()

	steps := []struct {
		to      lifecycle.AgentState
		trigger string
	}{
		{lifecycle.StateRunning, "agent started"},
		{lifecycle.StateReviewing, "task completed"},
		{lifecycle.StateApproved, "review passed"},
		{lifecycle.StateCommitted, "commit succeeded"},
	}

	for _, step := range steps {
		if err := sm.Transition(step.to, step.trigger); err != nil {
			t.Fatalf("遷移 → %s が失敗: %v", step.to, err)
		}
		if sm.Current() != step.to {
			t.Errorf("遷移後の状態が %s であるべきだが %s だった", step.to, sm.Current())
		}
	}
}

// ============================================================
// 異常系遷移
// ============================================================

func TestAbnormal_SpawningToFailed(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	if err := sm.Transition(lifecycle.StateFailed, "spawn error"); err != nil {
		t.Errorf("SPAWNING → FAILED は許可されるべきだが: %v", err)
	}
}

func TestAbnormal_RunningToFailed(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateRunning, "started")
	if err := sm.Transition(lifecycle.StateFailed, "retry exceeded"); err != nil {
		t.Errorf("RUNNING → FAILED は許可されるべきだが: %v", err)
	}
}

func TestAbnormal_RunningToCancelled(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateRunning, "started")
	if err := sm.Transition(lifecycle.StateCancelled, "user interrupt"); err != nil {
		t.Errorf("RUNNING → CANCELLED は許可されるべきだが: %v", err)
	}
}

func TestAbnormal_ReviewingToFailed(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateRunning, "started")
	_ = sm.Transition(lifecycle.StateReviewing, "task done")
	if err := sm.Transition(lifecycle.StateFailed, "review error"); err != nil {
		t.Errorf("REVIEWING → FAILED は許可されるべきだが: %v", err)
	}
}

func TestAbnormal_ReviewingToCancelled(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateRunning, "started")
	_ = sm.Transition(lifecycle.StateReviewing, "task done")
	if err := sm.Transition(lifecycle.StateCancelled, "user interrupt"); err != nil {
		t.Errorf("REVIEWING → CANCELLED は許可されるべきだが: %v", err)
	}
}

func TestAbnormal_RunningToStale(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateRunning, "started")
	if err := sm.Transition(lifecycle.StateStale, "24h exceeded"); err != nil {
		t.Errorf("RUNNING → STALE は許可されるべきだが: %v", err)
	}
}

func TestAbnormal_ReviewingToStale(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateRunning, "started")
	_ = sm.Transition(lifecycle.StateReviewing, "task done")
	if err := sm.Transition(lifecycle.StateStale, "24h exceeded"); err != nil {
		t.Errorf("REVIEWING → STALE は許可されるべきだが: %v", err)
	}
}

func TestAbnormal_FailedToRecovering(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateFailed, "error")
	if err := sm.Transition(lifecycle.StateRecovering, "recovery started"); err != nil {
		t.Errorf("FAILED → RECOVERING は許可されるべきだが: %v", err)
	}
}

func TestAbnormal_RecoveringToRunning(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateFailed, "error")
	_ = sm.Transition(lifecycle.StateRecovering, "recovery started")
	if err := sm.Transition(lifecycle.StateRunning, "recovery succeeded"); err != nil {
		t.Errorf("RECOVERING → RUNNING は許可されるべきだが: %v", err)
	}
}

func TestAbnormal_RecoveringToAborted(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateFailed, "error")
	_ = sm.Transition(lifecycle.StateRecovering, "recovery started")
	if err := sm.Transition(lifecycle.StateAborted, "recovery failed"); err != nil {
		t.Errorf("RECOVERING → ABORTED は許可されるべきだが: %v", err)
	}
}

// ============================================================
// 不正遷移
// ============================================================

func TestInvalidTransition_SpawningToCommitted(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	err := sm.Transition(lifecycle.StateCommitted, "skip all")
	if err == nil {
		t.Error("SPAWNING → COMMITTED は拒否されるべきだが許可された")
	}
}

func TestInvalidTransition_SpawningToReviewing(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	err := sm.Transition(lifecycle.StateReviewing, "skip running")
	if err == nil {
		t.Error("SPAWNING → REVIEWING は拒否されるべきだが許可された")
	}
}

func TestInvalidTransition_RunningToApproved(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateRunning, "started")
	err := sm.Transition(lifecycle.StateApproved, "skip review")
	if err == nil {
		t.Error("RUNNING → APPROVED は拒否されるべきだが許可された")
	}
}

func TestInvalidTransition_ApprovedToRunning(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateRunning, "started")
	_ = sm.Transition(lifecycle.StateReviewing, "task done")
	_ = sm.Transition(lifecycle.StateApproved, "approved")
	err := sm.Transition(lifecycle.StateRunning, "go back")
	if err == nil {
		t.Error("APPROVED → RUNNING は拒否されるべきだが許可された")
	}
}

// ============================================================
// 終端状態からの遷移は全て失敗
// ============================================================

func TestTerminalStates_AllTransitionsBlocked(t *testing.T) {
	allStates := []lifecycle.AgentState{
		lifecycle.StateSpawning,
		lifecycle.StateRunning,
		lifecycle.StateReviewing,
		lifecycle.StateApproved,
		lifecycle.StateCommitted,
		lifecycle.StateFailed,
		lifecycle.StateCancelled,
		lifecycle.StateStale,
		lifecycle.StateRecovering,
		lifecycle.StateAborted,
	}

	terminalList := []lifecycle.AgentState{
		lifecycle.StateCommitted,
		lifecycle.StateAborted,
		lifecycle.StateCancelled,
	}

	for _, terminal := range terminalList {
		for _, target := range allStates {
			sm := newSMAt(t, terminal)
			err := sm.Transition(target, "from terminal")
			if err == nil {
				t.Errorf("終端状態 %s → %s は拒否されるべきだが許可された", terminal, target)
			}
		}
	}
}

// newSMAt は指定した終端状態にある StateMachine を返すヘルパー。
// COMMITTED / ABORTED / CANCELLED のそれぞれを最短経路で到達する。
func newSMAt(t *testing.T, state lifecycle.AgentState) *lifecycle.StateMachine {
	t.Helper()
	sm := lifecycle.NewStateMachine()

	var steps []struct {
		to      lifecycle.AgentState
		trigger string
	}

	switch state {
	case lifecycle.StateCommitted:
		steps = []struct {
			to      lifecycle.AgentState
			trigger string
		}{
			{lifecycle.StateRunning, "started"},
			{lifecycle.StateReviewing, "done"},
			{lifecycle.StateApproved, "approved"},
			{lifecycle.StateCommitted, "committed"},
		}
	case lifecycle.StateAborted:
		steps = []struct {
			to      lifecycle.AgentState
			trigger string
		}{
			{lifecycle.StateFailed, "error"},
			{lifecycle.StateRecovering, "recovering"},
			{lifecycle.StateAborted, "recovery failed"},
		}
	case lifecycle.StateCancelled:
		steps = []struct {
			to      lifecycle.AgentState
			trigger string
		}{
			{lifecycle.StateRunning, "started"},
			{lifecycle.StateCancelled, "user interrupt"},
		}
	default:
		t.Fatalf("newSMAt: %s は終端状態ではありません", state)
	}

	for _, step := range steps {
		if err := sm.Transition(step.to, step.trigger); err != nil {
			t.Fatalf("セットアップ遷移 → %s が失敗: %v", step.to, err)
		}
	}
	return sm
}

// ============================================================
// 遷移履歴
// ============================================================

func TestHistory_RecordsAllTransitions(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateRunning, "trigger-1")
	_ = sm.Transition(lifecycle.StateReviewing, "trigger-2")

	history := sm.History()
	if len(history) != 2 {
		t.Fatalf("履歴は 2 件あるべきだが %d 件だった", len(history))
	}

	if history[0].From != lifecycle.StateSpawning || history[0].To != lifecycle.StateRunning {
		t.Errorf("履歴[0] が期待と異なる: %+v", history[0])
	}
	if history[0].Trigger != "trigger-1" {
		t.Errorf("履歴[0].Trigger が %q であるべきだが %q だった", "trigger-1", history[0].Trigger)
	}
	if history[1].From != lifecycle.StateRunning || history[1].To != lifecycle.StateReviewing {
		t.Errorf("履歴[1] が期待と異なる: %+v", history[1])
	}
	if history[1].Trigger != "trigger-2" {
		t.Errorf("履歴[1].Trigger が %q であるべきだが %q だった", "trigger-2", history[1].Trigger)
	}
}

func TestHistory_FailedTransitionNotRecorded(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateRunning, "ok")
	// 不正遷移（失敗するはず）
	_ = sm.Transition(lifecycle.StateCommitted, "invalid")

	history := sm.History()
	if len(history) != 1 {
		t.Errorf("不正遷移は履歴に記録されないべきだが %d 件あった", len(history))
	}
}

func TestHistory_ReturnsIndependentCopy(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	_ = sm.Transition(lifecycle.StateRunning, "started")

	h1 := sm.History()
	h1[0].Trigger = "TAMPERED"

	h2 := sm.History()
	if h2[0].Trigger == "TAMPERED" {
		t.Error("History() が内部スライスの参照を返している（コピーであるべき）")
	}
}

// ============================================================
// CanTransition
// ============================================================

func TestCanTransition_ValidTransitionReturnsTrue(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	if !sm.CanTransition(lifecycle.StateRunning) {
		t.Error("SPAWNING から RUNNING への遷移は可能であるべき")
	}
}

func TestCanTransition_InvalidTransitionReturnsFalse(t *testing.T) {
	sm := lifecycle.NewStateMachine()
	if sm.CanTransition(lifecycle.StateCommitted) {
		t.Error("SPAWNING から COMMITTED への遷移は不可であるべき")
	}
}
