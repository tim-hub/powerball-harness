package hookhandler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// AutoCleanupHandler は PostToolUse フックハンドラ（自動サイズチェック）。
// Write/Edit ツールで書き込まれたファイルのサイズ（行数）をチェックし、
// Plans.md / session-log.md / CLAUDE.md が閾値を超えた場合に systemMessage で警告する。
//
// shell 版: scripts/auto-cleanup-hook.sh
type AutoCleanupHandler struct {
	// ProjectRoot はプロジェクトルートのパス。空の場合は cwd を使用する。
	ProjectRoot string

	// 閾値（0 の場合はデフォルト値を使用）
	PlansMaxLines      int
	SessionLogMaxLines int
	ClaudeMdMaxLines   int
}

const (
	defaultPlansMaxLines      = 200
	defaultSessionLogMaxLines = 500
	defaultClaudeMdMaxLines   = 100
)

// autoCleanupInput は PostToolUse フックの stdin JSON。
type autoCleanupInput struct {
	ToolInput    autoCleanupToolInput    `json:"tool_input"`
	ToolResponse autoCleanupToolResponse `json:"tool_response"`
	CWD          string                  `json:"cwd"`
}

type autoCleanupToolInput struct {
	FilePath string `json:"file_path"`
}

type autoCleanupToolResponse struct {
	FilePath string `json:"filePath"`
}

// Handle は stdin から PostToolUse ペイロードを読み取り、ファイルサイズをチェックする。
func (h *AutoCleanupHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	if len(data) == 0 {
		return nil
	}

	var inp autoCleanupInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil
	}

	filePath := inp.ToolInput.FilePath
	if filePath == "" {
		filePath = inp.ToolResponse.FilePath
	}
	if filePath == "" {
		return nil
	}

	cwd := inp.CWD
	if cwd == "" {
		if h.ProjectRoot != "" {
			cwd = h.ProjectRoot
		} else {
			cwd, _ = os.Getwd()
		}
	}

	// プロジェクト相対パスへ正規化
	if strings.HasPrefix(filePath, cwd+"/") {
		filePath = filePath[len(cwd)+1:]
	}

	// 閾値を決定
	plansMax := h.PlansMaxLines
	if plansMax == 0 {
		plansMax = h.envInt("PLANS_MAX_LINES", defaultPlansMaxLines)
	}
	sessionMax := h.SessionLogMaxLines
	if sessionMax == 0 {
		sessionMax = h.envInt("SESSION_LOG_MAX_LINES", defaultSessionLogMaxLines)
	}
	claudeMax := h.ClaudeMdMaxLines
	if claudeMax == 0 {
		claudeMax = h.envInt("CLAUDE_MD_MAX_LINES", defaultClaudeMdMaxLines)
	}

	// 絶対パスを解決（ファイルの存在確認に使う）
	absPath := filePath
	if !filepath.IsAbs(absPath) {
		absPath = filepath.Join(cwd, filePath)
	}

	feedback := h.checkFile(filePath, absPath, plansMax, sessionMax, claudeMax, cwd, resolveHarnessLocale(cwd))
	if feedback == "" {
		return nil
	}

	return writeCleanupOutput(w, feedback)
}

// checkFile はファイルを判別してサイズチェックを行い、フィードバック文字列を返す。
func (h *AutoCleanupHandler) checkFile(relPath, absPath string, plansMax, sessionMax, claudeMax int, cwd, locale string) string {
	lower := strings.ToLower(relPath)
	var feedback string

	switch {
	case strings.Contains(lower, "plans.md"):
		feedback = h.checkPlans(absPath, plansMax, cwd, locale)
	case strings.Contains(lower, "session-log.md"):
		feedback = h.checkSessionLog(absPath, sessionMax, locale)
	case strings.Contains(lower, "claude.md"):
		feedback = h.checkClaudeMd(absPath, claudeMax, locale)
	}

	return feedback
}

// checkPlans は Plans.md の行数をチェックし、アーカイブ検知も行う。
func (h *AutoCleanupHandler) checkPlans(absPath string, maxLines int, cwd, locale string) string {
	lines, err := countLines(absPath)
	if err != nil {
		return ""
	}

	var feedback string
	if lines > maxLines {
		feedback = localizedHarnessMessage(locale,
			fmt.Sprintf("Warning: Plans.md has %d lines (limit: %d). Consider archiving old tasks with /maintenance.", lines, maxLines),
			fmt.Sprintf("⚠️ Plans.md が %d 行です（上限: %d行）。/maintenance で古いタスクをアーカイブすることを推奨します。", lines, maxLines))
	}

	// アーカイブセクション検知 + SSOT フラグチェック
	if containsArchiveSection(absPath) {
		// リポジトリルートの stateDir を使用
		repoRoot := cwd
		if root, err := gitRepoRoot(cwd); err == nil {
			repoRoot = root
		}
		stateDir := filepath.Join(repoRoot, ".claude", "state")
		ssotFlag := filepath.Join(stateDir, ".ssot-synced-this-session")

		if !fileExists(ssotFlag) {
			ssotWarning := localizedHarnessMessage(locale,
				"**Run /memory sync before cleaning up Plans.md** - important decisions or learnings may not be reflected in the SSOT (decisions.md/patterns.md).",
				"**Plans.md クリーンアップ前に /memory sync を実行してください** - 重要な決定や学習事項が SSOT (decisions.md/patterns.md) に反映されていない可能性があります。")
			if feedback != "" {
				feedback = feedback + localizedHarnessMessage(locale, " | Warning: ", " | ⚠️ ") + ssotWarning
			} else {
				feedback = localizedHarnessMessage(locale, "Warning: ", "⚠️ ") + ssotWarning
			}
		}
	}

	return feedback
}

// checkSessionLog は session-log.md の行数をチェックする。
func (h *AutoCleanupHandler) checkSessionLog(absPath string, maxLines int, locale string) string {
	lines, err := countLines(absPath)
	if err != nil {
		return ""
	}
	if lines > maxLines {
		return localizedHarnessMessage(locale,
			fmt.Sprintf("Warning: session-log.md has %d lines (limit: %d). Consider splitting it by month with /maintenance.", lines, maxLines),
			fmt.Sprintf("⚠️ session-log.md が %d 行です（上限: %d行）。/maintenance で月別に分割することを推奨します。", lines, maxLines))
	}
	return ""
}

// checkClaudeMd は CLAUDE.md の行数をチェックする。
func (h *AutoCleanupHandler) checkClaudeMd(absPath string, maxLines int, locale string) string {
	lines, err := countLines(absPath)
	if err != nil {
		return ""
	}
	if lines > maxLines {
		return localizedHarnessMessage(locale,
			fmt.Sprintf("Warning: CLAUDE.md has %d lines. Consider splitting rules into .claude/rules/ or moving long content to docs/ and referencing it with @docs/filename.md.", lines),
			fmt.Sprintf("⚠️ CLAUDE.md が %d 行です。.claude/rules/ への分割、または docs/ に移動して @docs/filename.md で参照することを検討してください。", lines))
	}
	return ""
}

// countLines はファイルの行数を数える。
func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		count++
	}
	return count, sc.Err()
}

// containsArchiveSection はファイルにアーカイブセクションが含まれているかを確認する。
func containsArchiveSection(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, "📦 アーカイブ") ||
			strings.Contains(line, "## アーカイブ") ||
			strings.Contains(line, "Archive") {
			return true
		}
	}
	return false
}

// envInt は環境変数を整数として取得し、未設定またはパース失敗時はデフォルト値を返す。
func (h *AutoCleanupHandler) envInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	var n int
	if _, err := fmt.Sscanf(val, "%d", &n); err != nil {
		return defaultVal
	}
	return n
}

// writeCleanupOutput は feedback を additionalContext として JSON 出力する。
// bash は単純な JSON 文字列として出力しているため、同じ形式で出力する。
func writeCleanupOutput(w io.Writer, feedback string) error {
	type hookOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	}
	type output struct {
		HookSpecificOutput hookOutput `json:"hookSpecificOutput"`
	}
	out := output{
		HookSpecificOutput: hookOutput{
			HookEventName:     "PostToolUse",
			AdditionalContext: feedback,
		},
	}
	data, err := json.Marshal(out)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
