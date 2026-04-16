package guardrail

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tim-hub/powerball-harness/go/pkg/hookproto"
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
		Message: "Sensitive information may be embedded directly from environment variables into a string",
	},
	{
		Pattern: regexp.MustCompile(`(?i)eval\s*\(\s*(?:request|req|input|param|query)`),
		Message: "Detected code passing user input to eval() (RCE risk)",
	},
	{
		Pattern: regexp.MustCompile(`exec\s*\(\s*` + "`" + `[^` + "`" + `]*\$\{`),
		Message: "Detected code passing a template literal to exec() (command injection risk)",
	},
	{
		Pattern: regexp.MustCompile(`innerHTML\s*=\s*(?:.*\+.*|` + "`" + `[^` + "`" + `]*\$\{)`),
		Message: "Detected code setting user input to innerHTML (XSS risk)",
	},
	{
		Pattern: regexp.MustCompile(`(?i)(?:password|passwd|secret|api_key|apikey)\s*=\s*["'][^"']{8,}["']`),
		Message: "Hardcoded sensitive information detected (password/API key)",
	},
	// Anthropic API keys (sk-ant-...)
	{
		Pattern: regexp.MustCompile(`sk-ant-[a-zA-Z0-9\-]{20,}`),
		Message: "Hardcoded Anthropic API key detected (sk-ant-...)",
	},
	// OpenAI API keys (sk-... but not sk-ant-)
	{
		Pattern: regexp.MustCompile(`\bsk-[a-zA-Z0-9]{20,}\b`),
		Message: "Hardcoded OpenAI/generic API key detected (sk-...)",
	},
	// AWS access keys
	{
		Pattern: regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		Message: "Hardcoded AWS access key detected (AKIA...)",
	},
	// GitHub personal access tokens (classic and fine-grained)
	{
		Pattern: regexp.MustCompile(`gh[pousre]_[A-Za-z0-9_]{36,}`),
		Message: "Hardcoded GitHub token detected (ghp_/gho_/ghu_/ghs_/ghr_/ghe_...)",
	},
	// Stripe keys
	{
		Pattern: regexp.MustCompile(`[rs]k_live_[a-zA-Z0-9]{20,}`),
		Message: "Hardcoded Stripe live key detected (sk_live_/rk_live_...)",
	},
	// JWT tokens (header.payload format — both parts base64url encoded)
	{
		Pattern: regexp.MustCompile(`eyJ[A-Za-z0-9_\-]{10,}\.eyJ[A-Za-z0-9_\-]{10,}`),
		Message: "Hardcoded JWT token detected (eyJ...eyJ...)",
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

// hasSuspiciousContent performs cheap string pre-screening before running
// expensive regex patterns. Returns false if no pattern can possibly match,
// allowing detectSecurityRisks to be skipped entirely for clean files.
// This function is a superset of what the regexes can match — it must never
// return false when a pattern would match. False positives are acceptable;
// false negatives are not.
func hasSuspiciousContent(content string) bool {
	return strings.Contains(content, "process.env.") ||
		strings.Contains(content, "eval(") ||
		strings.Contains(content, "exec(") ||
		strings.Contains(content, "innerHTML") ||
		strings.Contains(content, "password") ||
		strings.Contains(content, "passwd") ||
		strings.Contains(content, "secret") ||
		strings.Contains(content, "api_key") ||
		strings.Contains(content, "apikey") ||
		strings.Contains(content, "sk-ant-") ||
		strings.Contains(content, "sk-") ||
		strings.Contains(content, "AKIA") ||
		strings.Contains(content, "ghp_") ||
		strings.Contains(content, "gho_") ||
		strings.Contains(content, "ghu_") ||
		strings.Contains(content, "ghs_") ||
		strings.Contains(content, "ghr_") ||
		strings.Contains(content, "ghe_") ||
		strings.Contains(content, "_live_") ||
		strings.Contains(content, "eyJ")
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

	filePath, _ := getStringField(input.ToolInput, "file_path")

	// Read content once — reused for both tampering and security checks
	content := getChangedContent(input.ToolInput)

	// Tampering detection — only for test and CI/config files
	if filePath != "" && content != "" {
		isTest := isTestFile(filePath)
		isConfig := isConfigFile(filePath)

		if isTest || isConfig {
			warnings := detectTampering(content, isTest)
			if len(warnings) > 0 {
				fileType := "test file"
				if !isTest {
					fileType = "CI/config file"
				}
				var lines []string
				for _, w := range warnings {
					lines = append(lines, fmt.Sprintf("- [%s] %s\n  Detected at: %s", w.PatternID, w.Description, w.MatchedText))
				}
				msg := fmt.Sprintf("[v4] Test tampering warning\n\nSuspicious pattern detected in %s `%s`:\n\n%s\n\n[Please verify]\nCheck that this change does not intentionally disable tests or lower implementation quality.\nIf tampering is determined, revert the change.",
					fileType, filePath, strings.Join(lines, "\n"))
				systemMessages = append(systemMessages, msg)
			}
		}
	}

	// Security risk detection — all files, pre-screened for efficiency
	if content != "" && hasSuspiciousContent(content) {
		secWarnings := detectSecurityRisks(content)
		if len(secWarnings) > 0 {
			var lines []string
			for _, w := range secWarnings {
				lines = append(lines, fmt.Sprintf("- %s", w))
			}
			msg := fmt.Sprintf("[v4] Security risk detected:\n%s", strings.Join(lines, "\n"))
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
