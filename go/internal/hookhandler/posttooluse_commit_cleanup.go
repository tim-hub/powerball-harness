package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// CommitCleanupHandler は PostToolUse フックハンドラ（git commit 後のクリーンアップ）。
// git commit コマンドが成功した後に、レビュー承認状態ファイルを削除する。
//
// shell 版: scripts/posttooluse-commit-cleanup.sh
type CommitCleanupHandler struct {
	// ProjectRoot はプロジェクトルートのパス。空の場合は cwd を使用する。
	ProjectRoot string
}

// commitCleanupInput は PostToolUse フックの stdin JSON。
type commitCleanupInput struct {
	ToolName   string                 `json:"tool_name,omitempty"`
	ToolInput  map[string]interface{} `json:"tool_input,omitempty"`
	ToolResult interface{}            `json:"tool_result,omitempty"`
}

// Handle は stdin から PostToolUse ペイロードを読み取り、
// git commit コマンドが成功していた場合にレビュー承認状態ファイルを削除する。
// このハンドラは標準出力にはログメッセージのみ書き出す（JSON 不要）。
func (h *CommitCleanupHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	if len(data) == 0 {
		return nil
	}

	var inp commitCleanupInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil
	}

	// Bash ツール以外はスキップ
	if inp.ToolName != "Bash" {
		return nil
	}

	// コマンドを取得
	command := ""
	if v, ok := inp.ToolInput["command"]; ok {
		if s, ok := v.(string); ok {
			command = s
		}
	}
	if command == "" {
		return nil
	}

	// git commit コマンドかどうかを確認（大文字小文字を区別しない）
	if !isGitCommitCommand(command) {
		return nil
	}

	// ツール結果を文字列に変換
	toolResult := ""
	switch v := inp.ToolResult.(type) {
	case string:
		toolResult = v
	case map[string]interface{}:
		if b, err := json.Marshal(v); err == nil {
			toolResult = string(b)
		}
	}

	// エラーが含まれている場合はスキップ
	if containsErrorIndicator(toolResult) {
		return nil
	}

	// レビュー承認状態ファイルを削除
	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	reviewStateFile := projectRoot + "/.claude/state/review-approved.json"
	reviewResultFile := projectRoot + "/.claude/state/review-result.json"

	stateFileExists := fileExists(reviewStateFile)
	resultFileExists := fileExists(reviewResultFile)

	if stateFileExists || resultFileExists {
		_ = os.Remove(reviewStateFile)
		_ = os.Remove(reviewResultFile)

		_, _ = fmt.Fprintf(w, "[Commit Guard] レビュー承認状態をクリアしました。次回のコミット前に再度独立レビューを実行してください。\n")
	}

	return nil
}

// isGitCommitCommand は command 文字列に git commit が含まれているかを判定する。
// bash の grep -Eiq と同等: '(^|[[:space:]])git[[:space:]]+commit([[:space:]]|$)'
func isGitCommitCommand(command string) bool {
	lower := strings.ToLower(command)
	// "git commit" のパターンを順次探索
	searchFrom := 0
	for searchFrom < len(lower) {
		idx := strings.Index(lower[searchFrom:], "git")
		if idx < 0 {
			break
		}
		absIdx := searchFrom + idx

		// "git" の前が行頭またはスペース
		if absIdx > 0 && !isWordBoundaryBefore(lower[absIdx-1]) {
			searchFrom = absIdx + 1
			continue
		}

		// "git" の後にスペースがある
		afterGit := absIdx + 3
		if afterGit >= len(lower) || !isWordBoundaryBefore(lower[afterGit]) {
			searchFrom = absIdx + 1
			continue
		}

		// スペースをスキップして "commit" を探す
		i := afterGit
		for i < len(lower) && isWordBoundaryBefore(lower[i]) {
			i++
		}
		if strings.HasPrefix(lower[i:], "commit") {
			after := i + 6
			if after >= len(lower) || isWordBoundaryBefore(lower[after]) {
				return true
			}
		}
		searchFrom = absIdx + 1
	}
	return false
}

// isWordBoundaryBefore は c がスペース（単語境界）かどうかを返す。
func isWordBoundaryBefore(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// containsErrorIndicator はツール結果にエラーの兆候が含まれているかを判定する。
func containsErrorIndicator(result string) bool {
	lower := strings.ToLower(result)
	for _, indicator := range []string{"error", "fatal", "failed", "nothing to commit"} {
		if strings.Contains(lower, indicator) {
			return true
		}
	}
	return false
}
