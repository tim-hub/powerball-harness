package event

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// StalenessThreshold は連続失敗カウンターのリセット閾値（秒）。
// 前回失敗から StalenessThreshold 秒以上経過した場合はリセットする。
const StalenessThreshold = 60

// PostToolFailureHandler は PostToolUseFailure フックハンドラ。
// 連続ツール失敗をカウントし、3 回連続で escalation メッセージを返す。
//
// shell 版: scripts/hook-handlers/post-tool-failure.sh
type PostToolFailureHandler struct {
	// StateDir はカウンターファイルの保存先。
	// 空の場合は ResolveStateDir(projectRoot) を使う。
	StateDir string
	// nowFunc はテスト用の時刻注入関数。nil の場合は time.Now() を使う。
	nowFunc func() time.Time
}

// postToolFailureInput は PostToolUseFailure フックの stdin JSON。
type postToolFailureInput struct {
	ToolName string `json:"tool_name"`
	// toolName の別名も許容
	ToolNameAlt string `json:"toolName,omitempty"`
	Error       string `json:"error,omitempty"`
	Message     string `json:"message,omitempty"`
}

// counterRecord はカウンターファイルのレコード。
type counterRecord struct {
	Count     int
	Timestamp int64
}

// Handle は stdin から PostToolUseFailure ペイロードを読み取り、
// 連続失敗カウントに応じた systemMessage を stdout に書き出す。
func (h *PostToolFailureHandler) Handle(r io.Reader, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil || len(data) == 0 {
		return WriteJSON(w, SystemMessageResponse{})
	}

	var inp postToolFailureInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return WriteJSON(w, SystemMessageResponse{})
	}

	toolName := inp.ToolName
	if toolName == "" {
		toolName = inp.ToolNameAlt
	}
	if toolName == "" {
		toolName = "unknown"
	}

	errorMsg := inp.Error
	if errorMsg == "" {
		errorMsg = inp.Message
	}
	if len(errorMsg) > 200 {
		errorMsg = errorMsg[:200]
	}

	// ステートディレクトリの確保
	stateDir := h.StateDir
	if stateDir == "" {
		projectRoot := resolveProjectRoot(data)
		stateDir = ResolveStateDir(projectRoot)
	}

	if err := EnsureStateDir(stateDir); err != nil {
		return WriteJSON(w, SystemMessageResponse{})
	}

	counterFile := filepath.Join(stateDir, "tool-failure-counter.txt")
	if isSymlink(counterFile) {
		return WriteJSON(w, SystemMessageResponse{})
	}

	now := h.now().Unix()
	rec := h.readCounter(counterFile, now)
	rec.Count++
	rec.Timestamp = now

	if err := h.writeCounter(counterFile, rec); err != nil {
		return WriteJSON(w, SystemMessageResponse{})
	}

	if rec.Count >= 3 {
		// 3 回連続失敗: escalation
		h.resetCounter(counterFile)
		msg := fmt.Sprintf(
			"WARNING: %d consecutive tool failures detected (tool: %s). "+
				"Stop retrying the same approach. Diagnose the root cause or try an alternative approach. "+
				"Last error: %s",
			rec.Count, toolName, errorMsg,
		)
		return WriteJSON(w, SystemMessageResponse{SystemMessage: msg})
	}

	// 失敗 1-2 回: 警告のみ
	msg := fmt.Sprintf(
		"Tool failure #%d/3 (tool: %s). Will escalate after 3 consecutive failures.",
		rec.Count, toolName,
	)
	return WriteJSON(w, SystemMessageResponse{SystemMessage: msg})
}

// now は現在時刻を返す（テスト用に注入可能）。
func (h *PostToolFailureHandler) now() time.Time {
	if h.nowFunc != nil {
		return h.nowFunc()
	}
	return time.Now()
}

// readCounter はカウンターファイルを読み取る。
// ファイルがない、または形式が不正な場合は count=0 を返す。
// 前回失敗から StalenessThreshold 秒以上経過していた場合もリセットする。
func (h *PostToolFailureHandler) readCounter(path string, now int64) counterRecord {
	data, err := os.ReadFile(path)
	if err != nil {
		return counterRecord{}
	}

	parts := strings.Fields(strings.TrimSpace(string(data)))
	if len(parts) < 2 {
		return counterRecord{}
	}

	count, err1 := strconv.Atoi(parts[0])
	ts, err2 := strconv.ParseInt(parts[1], 10, 64)
	if err1 != nil || err2 != nil {
		return counterRecord{}
	}

	// 古い場合はリセット
	if now-ts > StalenessThreshold {
		return counterRecord{}
	}

	return counterRecord{Count: count, Timestamp: ts}
}

// writeCounter はカウンターファイルに書き出す。
func (h *PostToolFailureHandler) writeCounter(path string, rec counterRecord) error {
	if isSymlink(path) {
		return fmt.Errorf("security: symlinked counter file: %s", path)
	}
	content := fmt.Sprintf("%d %d\n", rec.Count, rec.Timestamp)
	return os.WriteFile(path, []byte(content), 0600)
}

// resetCounter はカウンターを 0 にリセットする。
func (h *PostToolFailureHandler) resetCounter(path string) {
	_ = h.writeCounter(path, counterRecord{Count: 0, Timestamp: 0})
}

// resolveProjectRoot は入力 JSON や環境変数からプロジェクトルートを推測する。
func resolveProjectRoot(data []byte) string {
	// CWD フィールドを試みる
	var v struct {
		CWD string `json:"cwd"`
	}
	if err := json.Unmarshal(data, &v); err == nil && v.CWD != "" {
		return v.CWD
	}

	// 環境変数フォールバック
	if r := os.Getenv("HARNESS_PROJECT_ROOT"); r != "" {
		return r
	}
	if r := os.Getenv("PROJECT_ROOT"); r != "" {
		return r
	}

	cwd, _ := os.Getwd()
	return cwd
}
