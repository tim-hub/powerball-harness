package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ClearPendingHandler は PostToolUse フックハンドラ（pending-skills クリア）。
// Skill ツール実行後に .claude/state/pending-skills/*.pending ファイルを削除する。
// Skill の呼び出しをもって品質ゲート実行済みとみなし、pending 状態を解消する。
//
// shell 版: scripts/posttooluse-clear-pending.sh
type ClearPendingHandler struct {
	// ProjectRoot はプロジェクトルートのパス。空の場合は cwd を使用する。
	ProjectRoot string
}

// clearPendingResponse は ClearPending フックのレスポンス。
type clearPendingResponse struct {
	Continue bool `json:"continue"`
}

// Handle は stdin からペイロードを読み取り（使用しない）、
// pending-skills ディレクトリの *.pending ファイルをすべて削除する。
func (h *ClearPendingHandler) Handle(r io.Reader, w io.Writer) error {
	// stdin は読み捨て（このハンドラは入力を使用しない）
	_, _ = io.ReadAll(r)

	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	pendingDir := filepath.Join(projectRoot, ".claude", "state", "pending-skills")

	// pending ディレクトリが存在しない場合はスキップ
	if _, err := os.Stat(pendingDir); os.IsNotExist(err) {
		return writePendingJSON(w, clearPendingResponse{Continue: true})
	}

	// *.pending ファイルをすべて削除
	matches, err := filepath.Glob(filepath.Join(pendingDir, "*.pending"))
	if err == nil {
		for _, path := range matches {
			_ = os.Remove(path)
		}
	}

	return writePendingJSON(w, clearPendingResponse{Continue: true})
}

// writePendingJSON は v を JSON として w に書き出す。
func writePendingJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
