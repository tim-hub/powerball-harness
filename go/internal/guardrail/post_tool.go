package guardrail

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Chachamaru127/claude-code-harness/go/pkg/hookproto"
)

// ---------------------------------------------------------------------------
// Security risk patterns (ported from post-tool.ts detectSecurityRisks)
// ---------------------------------------------------------------------------

type securityPattern struct {
	Pattern *regexp.Regexp
	Message string
}

var securityPatterns = []securityPattern{
	{
		Pattern: regexp.MustCompile(`(?i)process\.env\.[A-Z_]+.*(?:password|secret|key|token)`),
		Message: "機密情報を環境変数から直接文字列に埋め込んでいる可能性があります",
	},
	{
		Pattern: regexp.MustCompile(`(?i)eval\s*\(\s*(?:request|req|input|param|query)`),
		Message: "ユーザー入力を eval() に渡すコードを検出しました（RCE リスク）",
	},
	{
		Pattern: regexp.MustCompile(`exec\s*\(\s*` + "`" + `[^` + "`" + `]*\$\{`),
		Message: "テンプレートリテラルを exec() に渡すコードを検出しました（コマンドインジェクションリスク）",
	},
	{
		Pattern: regexp.MustCompile(`innerHTML\s*=\s*(?:.*\+.*|` + "`" + `[^` + "`" + `]*\$\{)`),
		Message: "ユーザー入力を innerHTML に設定しているコードを検出しました（XSS リスク）",
	},
	{
		Pattern: regexp.MustCompile(`(?i)(?:password|passwd|secret|api_key|apikey)\s*=\s*["'][^"']{8,}["']`),
		Message: "ハードコードされた機密情報（パスワード/APIキー）を検出しました",
	},
}

func detectSecurityRisks(content string) []string {
	var warnings []string
	for _, sp := range securityPatterns {
		if sp.Pattern.MatchString(content) {
			warnings = append(warnings, sp.Message)
		}
	}
	return warnings
}

// ---------------------------------------------------------------------------
// EvaluatePostTool — PostToolUse hook entry point
// ---------------------------------------------------------------------------

// EvaluatePostTool evaluates post-tool checks (tampering detection + security review).
// Only Write/Edit/MultiEdit are inspected; all other tools get immediate approve.
func EvaluatePostTool(input hookproto.HookInput) hookproto.HookResult {
	if !matchesWriteEditMultiEdit(input.ToolName) {
		return hookproto.HookResult{Decision: hookproto.DecisionApprove}
	}

	var systemMessages []string

	// Tampering detection
	filePath, _ := getStringField(input.ToolInput, "file_path")
	if filePath != "" {
		isTest := isTestFile(filePath)
		isConfig := isConfigFile(filePath)

		if isTest || isConfig {
			content := getChangedContent(input.ToolInput)
			if content != "" {
				warnings := detectTampering(content, isTest)
				if len(warnings) > 0 {
					fileType := "テストファイル"
					if !isTest {
						fileType = "CI/設定ファイル"
					}
					var lines []string
					for _, w := range warnings {
						lines = append(lines, fmt.Sprintf("- [%s] %s\n  検出箇所: %s", w.PatternID, w.Description, w.MatchedText))
					}
					msg := fmt.Sprintf("[v4] テスト改ざん検出警告\n\n%s `%s` に疑わしいパターンが検出されました:\n\n%s\n\n【確認してください】\nこの変更がテストを意図的に無効化したり、実装品質を下げるものでないかを確認してください。\n改ざんと判断した場合は変更を元に戻してください。",
						fileType, filePath, strings.Join(lines, "\n"))
					systemMessages = append(systemMessages, msg)
				}
			}
		}
	}

	// Security risk detection
	content := getChangedContent(input.ToolInput)
	if content != "" {
		secWarnings := detectSecurityRisks(content)
		if len(secWarnings) > 0 {
			var lines []string
			for _, w := range secWarnings {
				lines = append(lines, fmt.Sprintf("- %s", w))
			}
			msg := fmt.Sprintf("[v4] セキュリティリスク検出:\n%s", strings.Join(lines, "\n"))
			systemMessages = append(systemMessages, msg)
		}
	}

	if len(systemMessages) == 0 {
		return hookproto.HookResult{Decision: hookproto.DecisionApprove}
	}

	return hookproto.HookResult{
		Decision:      hookproto.DecisionApprove,
		SystemMessage: strings.Join(systemMessages, "\n\n---\n\n"),
	}
}
