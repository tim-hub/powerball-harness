package session

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CleanupHandler は SessionEnd フックハンドラ。
// セッション終了時に一時ファイルを削除する。
//
// shell 版: scripts/session-cleanup.sh
type CleanupHandler struct {
	// StateDir はステートディレクトリのパス。空の場合は cwd から推定する。
	StateDir string
}

// cleanupInput は SessionEnd フックの stdin JSON。
type cleanupInput struct {
	CWD string `json:"cwd,omitempty"`
}

// cleanupResponse はクリーンアップ結果のレスポンス。
type cleanupResponse struct {
	Continue bool   `json:"continue"`
	Message  string `json:"message"`
}

// Handle は stdin から SessionEnd ペイロードを読み取り、
// 一時ファイルを削除して結果を stdout に書き出す。
func (h *CleanupHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	var inp cleanupInput
	if len(data) > 0 {
		_ = json.Unmarshal(data, &inp)
	}

	// ステートディレクトリを決定
	stateDir := h.StateDir
	if stateDir == "" {
		projectRoot := resolveProjectRoot(inp.CWD)
		stateDir = filepath.Join(projectRoot, ".claude", "state")
	}

	// ステートディレクトリが存在しない場合は早期リターン
	if _, err := os.Stat(stateDir); err != nil {
		return writeJSON(w, cleanupResponse{Continue: true, Message: "No state directory"})
	}

	// シンボリックリンクチェック（セキュリティ）
	if isSymlink(stateDir) {
		return writeJSON(w, cleanupResponse{Continue: true, Message: "State directory is symlink, skipping"})
	}

	// 固定の一時ファイルを削除
	tempFiles := []string{
		"pending-skill.json",
		"current-operation.json",
	}
	for _, name := range tempFiles {
		path := filepath.Join(stateDir, name)
		if isRegularFile(path) {
			_ = os.Remove(path)
		}
	}

	// inbox-*.tmp ファイルをクリーンアップ
	h.cleanupGlob(stateDir, "inbox-*.tmp")

	return writeJSON(w, cleanupResponse{Continue: true, Message: "Session cleanup completed"})
}

// cleanupGlob はステートディレクトリ内の glob パターンにマッチするファイルを削除する。
func (h *CleanupHandler) cleanupGlob(stateDir, pattern string) {
	matches, err := filepath.Glob(filepath.Join(stateDir, pattern))
	if err != nil {
		return
	}
	for _, path := range matches {
		if isRegularFile(path) {
			_ = os.Remove(path)
		}
	}
}

// isRegularFile はパスが通常ファイルかどうかを返す（シンボリックリンクは除外）。
func isRegularFile(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode().IsRegular()
}

// cleanupFilenameGlobMatch は pattern に対してファイル名が glob マッチするかを返す。
// filepath.Match を使用するが、パスセパレータなしのパターンのみ対応。
func cleanupFilenameGlobMatch(pattern, name string) bool {
	matched, err := filepath.Match(pattern, name)
	if err != nil {
		return false
	}
	return matched
}

// buildCleanupSummary はクリーンアップ対象のファイル一覧をログ用に構築する（デバッグ用）。
func buildCleanupSummary(files []string) string {
	if len(files) == 0 {
		return "none"
	}
	return strings.Join(files, ", ")
}

// formatCleanupResult はクリーンアップ結果の JSON を返す（エラー表示用）。
func formatCleanupResult(deleted int, err error) string {
	if err != nil {
		return fmt.Sprintf(`{"continue":true,"message":"cleanup partial: %d files removed, error: %v"}`, deleted, err)
	}
	return fmt.Sprintf(`{"continue":true,"message":"Session cleanup completed: %d files removed"}`, deleted)
}
