package hookhandler

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// WorktreeRemoveHandler は WorktreeRemove フックハンドラ。
// Breezing サブエージェント終了時に worktree 固有の一時ファイルをクリーンアップする。
//
// shell 版: scripts/hook-handlers/worktree-remove.sh
type WorktreeRemoveHandler struct{}

// worktreeRemoveInput は WorktreeRemove フックの stdin JSON。
type worktreeRemoveInput struct {
	SessionID string `json:"session_id,omitempty"`
	CWD       string `json:"cwd,omitempty"`
}

// worktreeRemoveResponse は WorktreeRemove フックのレスポンス。
type worktreeRemoveResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// Handle は stdin から WorktreeRemove ペイロードを読み取り、
// worktree 固有の一時ファイルを削除して結果を stdout に書き出す。
func (h *WorktreeRemoveHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	// ペイロードが空の場合はスキップ
	if len(data) == 0 || string(data) == "\n" || string(data) == "\r\n" {
		return writeJSON(w, worktreeRemoveResponse{
			Decision: "approve",
			Reason:   "WorktreeRemove: no payload",
		})
	}

	var inp worktreeRemoveInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return writeJSON(w, worktreeRemoveResponse{
			Decision: "approve",
			Reason:   "WorktreeRemove: no payload",
		})
	}

	if inp.SessionID == "" {
		return writeJSON(w, worktreeRemoveResponse{
			Decision: "approve",
			Reason:   "WorktreeRemove: no session_id",
		})
	}

	// Codex プロンプト一時ファイルを削除
	removeTmpGlob("/tmp/codex-prompt-*.md")

	// Harness Codex ログを削除
	removeTmpGlob("/tmp/harness-codex-*.log")

	// worktree-info.json のクリーンアップ
	if inp.CWD != "" {
		infoFile := filepath.Join(inp.CWD, ".claude", "state", "worktree-info.json")
		_ = os.Remove(infoFile)
	}

	return writeJSON(w, worktreeRemoveResponse{
		Decision: "approve",
		Reason:   "WorktreeRemove: cleaned up worktree resources",
	})
}

// removeTmpGlob はグロブパターンにマッチする /tmp 以下のファイルを削除する。
func removeTmpGlob(pattern string) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	for _, path := range matches {
		_ = os.Remove(path)
	}
}
