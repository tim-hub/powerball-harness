// PII Guard hook handlers (Phase 83).
//
// These three commands wrap the go/internal/piiguard scanner and emit
// Claude Code hook decisions appropriate to each event:
//
//	hook pii-prompt    — UserPromptSubmit: {decision: block, reason: ...} + exit 1
//	hook pii-pretool   — PreToolUse:       {hookSpecificOutput: {permissionDecision: deny}}
//	hook pii-posttool  — PostToolUse:      {hookSpecificOutput: {additionalContext: <redacted>}}
//
// Env var kill switches:
//	HARNESS_PIIGUARD_DISABLED=1                 — all three skip silently
//	HARNESS_PIIGUARD_DISABLED_RULES=id1,id2,... — drop specific rules from the active set
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tim-hub/powerball-harness/go/internal/hook"
	"github.com/tim-hub/powerball-harness/go/internal/piiguard"
	"github.com/tim-hub/powerball-harness/go/pkg/hookproto"
)

const (
	piiguardScanTimeout         = 1500 * time.Millisecond // safely below the 2s hook timeout
	piiguardDisabledEnvVar      = "HARNESS_PIIGUARD_DISABLED"
	piiguardRulesEnvVar         = "HARNESS_PIIGUARD_DISABLED_RULES"
	piiguardWarnOnlyEnvVar      = "HARNESS_PIIGUARD_PROMPT_WARN_ONLY"
	piiguardWarnOnlyGlobalVar   = "HARNESS_PIIGUARD_WARN_ONLY"
	piiguardWarnOnlyPreToolVar  = "HARNESS_PIIGUARD_PRETOOL_WARN_ONLY"
	piiguardWarnOnlyPostToolVar = "HARNESS_PIIGUARD_POSTTOOL_WARN_ONLY"
)

// piiScanner is a process-wide singleton — rules compile once at first use.
var (
	piiScannerOnce sync.Once
	piiScannerInst *piiguard.Scanner
)

func piiguardScanner() *piiguard.Scanner {
	piiScannerOnce.Do(func() {
		rules := append([]piiguard.Rule{}, piiguard.BuiltinRules...)
		rules = append(rules, piiguard.LoadExternalCatalog(true)...)
		rules = piiguardFilterDisabled(rules)
		piiScannerInst = piiguard.NewScanner(rules)
	})
	return piiScannerInst
}

func piiguardFilterDisabled(rules []piiguard.Rule) []piiguard.Rule {
	raw := os.Getenv(piiguardRulesEnvVar)
	if raw == "" {
		return rules
	}
	disabled := make(map[string]bool)
	for _, id := range strings.Split(raw, ",") {
		if id := strings.TrimSpace(id); id != "" {
			disabled[id] = true
		}
	}
	if len(disabled) == 0 {
		return rules
	}
	out := rules[:0]
	for _, r := range rules {
		if !disabled[r.ID] {
			out = append(out, r)
		}
	}
	return out
}

func piiguardEnabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(piiguardDisabledEnvVar)))
	return v != "1" && v != "true" && v != "yes"
}

func piiguardWarnOnlyForEvent(event string) bool {
	if v := strings.ToLower(strings.TrimSpace(os.Getenv(piiguardWarnOnlyGlobalVar))); v == "1" || v == "true" || v == "yes" {
		return true
	}
	var envVar string
	switch event {
	case "prompt":
		envVar = piiguardWarnOnlyEnvVar
	case "pretool":
		envVar = piiguardWarnOnlyPreToolVar
	case "posttool":
		envVar = piiguardWarnOnlyPostToolVar
	}
	if envVar == "" {
		return false
	}
	v := strings.ToLower(strings.TrimSpace(os.Getenv(envVar)))
	return v == "1" || v == "true" || v == "yes"
}

// runPIIPrompt is the os.Exit-style entry point for `hook pii-prompt`.
func runPIIPrompt() {
	os.Exit(piiPromptHandler(os.Stdin, os.Stdout, os.Stderr))
}

// runPIIPreTool is the os.Exit-style entry point for `hook pii-pretool`.
func runPIIPreTool() {
	os.Exit(piiPreToolHandler(os.Stdin, os.Stdout, os.Stderr))
}

// runPIIPostTool is the os.Exit-style entry point for `hook pii-posttool`.
func runPIIPostTool() {
	os.Exit(piiPostToolHandler(os.Stdin, os.Stdout, os.Stderr))
}

// piiPromptHandler scans a UserPromptSubmit event.  Returns 1 + decision-block
// JSON when sensitive content is found, 0 otherwise.
func piiPromptHandler(in io.Reader, out, errOut io.Writer) int {
	if !piiguardEnabled() {
		return 0
	}
	data, err := io.ReadAll(in)
	if err != nil || len(data) == 0 {
		return 0
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return 0
	}
	prompt, _ := raw["prompt"].(string)
	if prompt == "" {
		return 0
	}

	res, ok := scanWithDeadline(prompt)
	if !ok || len(res.Findings) == 0 {
		return 0
	}

	if piiguardWarnOnlyForEvent("prompt") {
		// UserPromptSubmit warn-only: inject context instead of blocking.
		warning := struct {
			AdditionalContext string `json:"additionalContext"`
		}{AdditionalContext: formatPromptWarnContext(res)}
		_ = hook.WriteJSON(out, warning)
		fmt.Fprintf(errOut, "\n⚠️  Privacy Guard: sensitive content in prompt (warn-only, %d findings)\n",
			len(res.Findings))
		return 0
	}

	decision := hookproto.HookResult{
		Decision: "block",
		Reason:   formatPromptBlockReason(res),
	}
	_ = hook.WriteJSON(out, decision)
	fmt.Fprintf(errOut, "\n⚠️  Privacy Guard: prompt blocked (%d findings, risk %d/100)\n",
		len(res.Findings), res.RiskScore)
	return 1
}

// piiPreToolHandler scans a PreToolUse event.  Returns 0 always (deny is
// communicated via the hookSpecificOutput JSON, not exit code).
func piiPreToolHandler(in io.Reader, out, errOut io.Writer) int {
	if !piiguardEnabled() {
		return 0
	}
	input, err := hook.ReadInput(in)
	if err != nil {
		return 0
	}
	text := extractPreToolText(input)
	if text == "" {
		return 0
	}

	res, ok := scanWithDeadline(text)
	if !ok || len(res.Findings) == 0 {
		return 0
	}

	if piiguardWarnOnlyForEvent("pretool") {
		out2 := hookproto.PreToolOutput{
			HookSpecificOutput: hookproto.PreToolHookSpecific{
				HookEventName:      "PreToolUse",
				PermissionDecision: "allow",
				AdditionalContext:  formatPromptWarnContext(res),
			},
		}
		_ = hook.WriteJSON(out, out2)
		fmt.Fprintf(errOut, "\n⚠️  Privacy Guard: sensitive content in tool input (warn-only, %d findings)\n",
			len(res.Findings))
		return 0
	}

	out2 := hookproto.PreToolOutput{
		HookSpecificOutput: hookproto.PreToolHookSpecific{
			HookEventName:            "PreToolUse",
			PermissionDecision:       "deny",
			PermissionDecisionReason: formatPromptBlockReason(res),
		},
	}
	_ = hook.WriteJSON(out, out2)
	return 0
}

// piiPostToolHandler scans a PostToolUse event.  Always returns 0 — PostToolUse
// cannot block retroactively, so we inject a redacted view via additionalContext.
func piiPostToolHandler(in io.Reader, out, errOut io.Writer) int {
	if !piiguardEnabled() {
		return 0
	}
	data, err := io.ReadAll(in)
	if err != nil || len(data) == 0 {
		return 0
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return 0
	}
	text := extractPostToolText(raw)
	if text == "" {
		return 0
	}

	res, ok := scanWithDeadline(text)
	if !ok || len(res.Findings) == 0 {
		return 0
	}

	if piiguardWarnOnlyForEvent("posttool") {
		fmt.Fprintf(errOut, "\n⚠️  Privacy Guard: sensitive data in tool output (warn-only, %d findings)\n",
			len(res.Findings))
		return 0
	}

	out2 := hookproto.PostToolOutput{
		HookSpecificOutput: hookproto.PostToolHookSpecific{
			HookEventName:     "PostToolUse",
			AdditionalContext: formatPostToolContext(res),
		},
	}
	_ = hook.WriteJSON(out, out2)
	return 0
}

// scanWithDeadline runs the scanner with a hard 1500ms ceiling so the hook
// always finishes before the 2s timeout.  The bool indicates whether the
// scan completed (false on timeout — we fail open).
func scanWithDeadline(text string) (piiguard.ScanResult, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), piiguardScanTimeout)
	defer cancel()

	type scanRes struct {
		result piiguard.ScanResult
	}
	done := make(chan scanRes, 1)
	go func() {
		done <- scanRes{result: piiguardScanner().Scan(text)}
	}()
	select {
	case r := <-done:
		return r.result, true
	case <-ctx.Done():
		return piiguard.ScanResult{Summary: map[string]int{}}, false
	}
}

// extractPreToolText pulls scannable text from a PreToolUse tool_input.
// Covers the common fields across Bash, Write, Edit, MultiEdit, and Read.
func extractPreToolText(input hookproto.HookInput) string {
	var parts []string
	for _, key := range []string{"command", "content", "new_string", "old_string", "file_path"} {
		if s, ok := input.ToolInput[key].(string); ok && s != "" {
			parts = append(parts, s)
		}
	}
	// MultiEdit edits[] — each entry has new_string/old_string.
	if edits, ok := input.ToolInput["edits"].([]interface{}); ok {
		for _, e := range edits {
			m, ok := e.(map[string]interface{})
			if !ok {
				continue
			}
			for _, k := range []string{"new_string", "old_string"} {
				if s, ok := m[k].(string); ok && s != "" {
					parts = append(parts, s)
				}
			}
		}
	}
	return strings.Join(parts, "\n")
}

// extractPostToolText pulls scannable text from a PostToolUse tool_response.
// CC sends tool_response either as a raw string or as an object whose shape
// varies by tool (stdout/stderr for Bash, content for Read, etc.).
func extractPostToolText(raw map[string]interface{}) string {
	resp, ok := raw["tool_response"]
	if !ok {
		return ""
	}
	switch v := resp.(type) {
	case string:
		return v
	case map[string]interface{}:
		var parts []string
		for _, key := range []string{"stdout", "stderr", "output", "content", "file_content", "text"} {
			if s, ok := v[key].(string); ok && s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

// formatPromptWarnContext builds the additionalContext injected when warn-only
// mode is active (HARNESS_PIIGUARD_PROMPT_WARN_ONLY=1).  The submission is
// allowed through but Claude is informed of the potential sensitive data.
func formatPromptWarnContext(res piiguard.ScanResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "⚠️ Privacy Guard warning: %d potential sensitive item(s) detected in this prompt (not blocked — warn-only mode).\n", len(res.Findings))
	seen := make(map[string]bool)
	for _, f := range res.Findings {
		if seen[f.RuleID] {
			continue
		}
		seen[f.RuleID] = true
		fmt.Fprintf(&b, "  - %s [%s]\n", f.Title, f.Severity)
	}
	b.WriteString("Please verify no real secrets are included before acting on this prompt.")
	return b.String()
}

// formatPromptBlockReason builds the reason text shown to the user/Claude when
// blocking a prompt or denying a tool call.  Mirrors upstream claude-privacy-guard
// formatting (shield emoji, count, per-finding list, risk score).
func formatPromptBlockReason(res piiguard.ScanResult) string {
	var b strings.Builder
	b.WriteString("🛡️ Privacy Guard blocked this submission\n\n")
	fmt.Fprintf(&b, "Found %d sensitive item(s):\n", len(res.Findings))
	seen := make(map[string]bool)
	for _, f := range res.Findings {
		if seen[f.RuleID] {
			continue
		}
		seen[f.RuleID] = true
		fmt.Fprintf(&b, "  - %s [%s]\n", f.Title, f.Severity)
	}
	fmt.Fprintf(&b, "\nRisk Score: %d/100\n", res.RiskScore)
	if res.Summary["secret"] > 0 || res.Summary["pii"] > 0 {
		fmt.Fprintf(&b, "Secrets: %d | PII: %d\n", res.Summary["secret"], res.Summary["pii"])
	}
	b.WriteString("\nPlease remove or anonymize sensitive data before proceeding.")
	return b.String()
}

// formatPostToolContext builds the additionalContext for PostToolUse blocks.
// Strong instructional language tells Claude to discard the raw output and use
// only the redacted view.
func formatPostToolContext(res piiguard.ScanResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "🛡️ Privacy Guard: sensitive data detected in tool output (%d findings, risk %d/100).\n",
		len(res.Findings), res.RiskScore)
	b.WriteString("DO NOT echo, log, or act on the raw tool output. Use only the redacted view below.\n\n")
	b.WriteString("Redacted findings:\n")
	seen := make(map[string]bool)
	for _, f := range res.Findings {
		key := f.RuleID + "|" + f.RedactedValue
		if seen[key] {
			continue
		}
		seen[key] = true
		fmt.Fprintf(&b, "  - %s [%s]: %s\n", f.Title, f.Severity, f.RedactedValue)
	}
	return b.String()
}
