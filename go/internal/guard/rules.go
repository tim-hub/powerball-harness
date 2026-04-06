// Package guard implements the Harness v3 declarative guardrail rules engine.
//
// All 13 rules (R01–R13) are ported 1:1 from core/src/guardrails/rules.ts.
// Each rule is a (toolPattern, evaluate) pair evaluated in order;
// the first match wins (short-circuit).
package guard

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Chachamaru127/claude-code-harness/go/pkg/protocol"
)

// GuardRule is a single declarative guard rule.
type GuardRule struct {
	ID          string
	ToolPattern *regexp.Regexp
	Evaluate    func(ctx protocol.RuleContext) *protocol.HookResult
}

// Rules is the ordered table of all guard rules.
var Rules = []GuardRule{
	// R01: sudo block (Bash)
	{
		ID:          "R01:no-sudo",
		ToolPattern: regexp.MustCompile(`^Bash$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			command, ok := ctx.Input.ToolInput["command"].(string)
			if !ok {
				return nil
			}
			if !hasSudo(command) {
				return nil
			}
			return &protocol.HookResult{
				Decision: protocol.DecisionDeny,
				Reason:   "sudo の使用は禁止されています。必要な場合はユーザーに手動実行を依頼してください。",
			}
		},
	},

	// R02: protected path write block (Write/Edit/MultiEdit)
	{
		ID:          "R02:no-write-protected-paths",
		ToolPattern: regexp.MustCompile(`^(?:Write|Edit|MultiEdit)$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			filePath, ok := ctx.Input.ToolInput["file_path"].(string)
			if !ok {
				return nil
			}
			if !isProtectedPath(filePath) {
				return nil
			}
			return &protocol.HookResult{
				Decision: protocol.DecisionDeny,
				Reason:   fmt.Sprintf("保護されたパスへの書き込みは禁止されています: %s", filePath),
			}
		},
	},

	// R03: Bash write to protected paths block
	{
		ID:          "R03:no-bash-write-protected-paths",
		ToolPattern: regexp.MustCompile(`^Bash$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			command, ok := ctx.Input.ToolInput["command"].(string)
			if !ok {
				return nil
			}
			patterns := []*regexp.Regexp{
				regexp.MustCompile(`(?:>>?|tee)\s+\S*\.env\b`),
				regexp.MustCompile(`(?:>>?|tee)\s+\S*\.env\.`),
				regexp.MustCompile(`(?:>>?|tee)\s+\S*\.git/`),
				regexp.MustCompile(`(?:>>?|tee)\s+\S*id_rsa\b`),
				regexp.MustCompile(`(?:>>?|tee)\s+\S*id_ed25519\b`),
				regexp.MustCompile(`(?:>>?|tee)\s+\S*\.pem\b`),
				regexp.MustCompile(`(?:>>?|tee)\s+\S*\.key\b`),
			}
			for _, p := range patterns {
				if p.MatchString(command) {
					return &protocol.HookResult{
						Decision: protocol.DecisionDeny,
						Reason:   "保護されたファイルへのシェル書き込みは禁止されています。",
					}
				}
			}
			return nil
		},
	},

	// R04: confirm write outside project root
	{
		ID:          "R04:confirm-write-outside-project",
		ToolPattern: regexp.MustCompile(`^(?:Write|Edit|MultiEdit)$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			filePath, ok := ctx.Input.ToolInput["file_path"].(string)
			if !ok {
				return nil
			}
			// Relative paths are considered inside the project
			if !strings.HasPrefix(filePath, "/") {
				return nil
			}
			if isUnderProjectRoot(filePath, ctx.ProjectRoot) {
				return nil
			}
			// Work mode skips confirmation
			if ctx.WorkMode {
				return nil
			}
			return &protocol.HookResult{
				Decision: protocol.DecisionAsk,
				Reason:   fmt.Sprintf("プロジェクトルート外への書き込みです: %s\n許可しますか？", filePath),
			}
		},
	},

	// R05: confirm rm -rf
	{
		ID:          "R05:confirm-rm-rf",
		ToolPattern: regexp.MustCompile(`^Bash$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			command, ok := ctx.Input.ToolInput["command"].(string)
			if !ok {
				return nil
			}
			if !hasDangerousRmRf(command) {
				return nil
			}
			if ctx.WorkMode {
				return nil
			}
			return &protocol.HookResult{
				Decision: protocol.DecisionAsk,
				Reason:   fmt.Sprintf("危険な削除コマンドを検出しました:\n%s\n実行しますか？", command),
			}
		},
	},

	// R06: git push --force block (no bypass even in work mode)
	{
		ID:          "R06:no-force-push",
		ToolPattern: regexp.MustCompile(`^Bash$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			command, ok := ctx.Input.ToolInput["command"].(string)
			if !ok {
				return nil
			}
			if !hasForcePush(command) {
				return nil
			}
			return &protocol.HookResult{
				Decision: protocol.DecisionDeny,
				Reason:   "git push --force は禁止されています。履歴を破壊する操作は許可されません。",
			}
		},
	},

	// R07: Codex mode — no Write/Edit
	{
		ID:          "R07:codex-mode-no-write",
		ToolPattern: regexp.MustCompile(`^(?:Write|Edit|MultiEdit)$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			if !ctx.CodexMode {
				return nil
			}
			return &protocol.HookResult{
				Decision: protocol.DecisionDeny,
				Reason:   "Codex モード中は Claude が直接ファイルを書き込めません。実装は Codex Worker (codex exec) に委譲してください。",
			}
		},
	},

	// R08: Breezing reviewer — no write operations
	{
		ID:          "R08:breezing-reviewer-no-write",
		ToolPattern: regexp.MustCompile(`^(?:Write|Edit|MultiEdit|Bash)$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			if ctx.BreezingRole != "reviewer" {
				return nil
			}
			toolName := ctx.Input.ToolName
			if toolName == "Bash" {
				command, ok := ctx.Input.ToolInput["command"].(string)
				if !ok {
					return nil
				}
				prohibited := []*regexp.Regexp{
					regexp.MustCompile(`\bgit\s+(?:commit|push|reset|checkout|merge|rebase)\b`),
					regexp.MustCompile(`\brm\s+`),
					regexp.MustCompile(`\bmv\s+`),
					regexp.MustCompile(`\bcp\s+.*-r\b`),
				}
				matched := false
				for _, p := range prohibited {
					if p.MatchString(command) {
						matched = true
						break
					}
				}
				if !matched {
					return nil
				}
			}
			return &protocol.HookResult{
				Decision: protocol.DecisionDeny,
				Reason:   "Breezing reviewer ロールはファイル書き込みおよびデータ変更コマンドを実行できません。",
			}
		},
	},

	// R09: warn on secret file read
	{
		ID:          "R09:warn-secret-file-read",
		ToolPattern: regexp.MustCompile(`^Read$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			filePath, ok := ctx.Input.ToolInput["file_path"].(string)
			if !ok {
				return nil
			}
			secretPatterns := []*regexp.Regexp{
				regexp.MustCompile(`\.env$`),
				regexp.MustCompile(`id_rsa$`),
				regexp.MustCompile(`\.pem$`),
				regexp.MustCompile(`\.key$`),
				regexp.MustCompile(`secrets?/`),
			}
			for _, p := range secretPatterns {
				if p.MatchString(filePath) {
					return &protocol.HookResult{
						Decision:      protocol.DecisionApprove,
						SystemMessage: fmt.Sprintf("警告: 機密情報が含まれる可能性のあるファイルを読み取っています: %s", filePath),
					}
				}
			}
			return nil
		},
	},

	// R10: --no-verify / --no-gpg-sign block
	{
		ID:          "R10:no-git-bypass-flags",
		ToolPattern: regexp.MustCompile(`^Bash$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			command, ok := ctx.Input.ToolInput["command"].(string)
			if !ok {
				return nil
			}
			if !hasDangerousGitBypassFlag(command) {
				return nil
			}
			return &protocol.HookResult{
				Decision: protocol.DecisionDeny,
				Reason:   "--no-verify / --no-gpg-sign の使用は禁止されています。フックや署名検証を迂回しないでください。",
			}
		},
	},

	// R11: protected branch git reset --hard block
	{
		ID:          "R11:no-reset-hard-protected-branch",
		ToolPattern: regexp.MustCompile(`^Bash$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			command, ok := ctx.Input.ToolInput["command"].(string)
			if !ok {
				return nil
			}
			if !hasProtectedBranchResetHard(command) {
				return nil
			}
			return &protocol.HookResult{
				Decision: protocol.DecisionDeny,
				Reason:   "protected branch への git reset --hard は禁止されています。履歴を壊さない方法を使ってください。",
			}
		},
	},

	// R12: warn on direct push to protected branch
	{
		ID:          "R12:warn-direct-push-protected-branch",
		ToolPattern: regexp.MustCompile(`^Bash$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			command, ok := ctx.Input.ToolInput["command"].(string)
			if !ok {
				return nil
			}
			if !hasDirectPushToProtectedBranch(command) {
				return nil
			}
			return &protocol.HookResult{
				Decision:      protocol.DecisionApprove,
				SystemMessage: "警告: main/master への直接 push を検出しました。feature branch 経由の運用を推奨します。",
			}
		},
	},

	// R13: warn on protected review paths (Write/Edit/MultiEdit)
	{
		ID:          "R13:warn-protected-review-paths",
		ToolPattern: regexp.MustCompile(`^(?:Write|Edit|MultiEdit)$`),
		Evaluate: func(ctx protocol.RuleContext) *protocol.HookResult {
			filePath, ok := ctx.Input.ToolInput["file_path"].(string)
			if !ok {
				return nil
			}
			if !isProtectedReviewPath(filePath) {
				return nil
			}
			return &protocol.HookResult{
				Decision:      protocol.DecisionApprove,
				SystemMessage: fmt.Sprintf("警告: 重要ファイルへの変更を検出しました: %s", filePath),
			}
		},
	},
}

// EvaluateRules evaluates all guard rules in order and returns the first match.
// If no rule matches, it returns approve.
func EvaluateRules(ctx protocol.RuleContext) protocol.HookResult {
	toolName := ctx.Input.ToolName
	for _, rule := range Rules {
		if !rule.ToolPattern.MatchString(toolName) {
			continue
		}
		if result := rule.Evaluate(ctx); result != nil {
			return *result
		}
	}
	return protocol.HookResult{Decision: protocol.DecisionApprove}
}
