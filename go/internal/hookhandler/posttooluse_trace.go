package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tim-hub/powerball-harness/go/internal/trace"
)

// posttooluseTraceInput is the stdin JSON shape passed by Claude Code to
// PostToolUse hooks. We only read the fields we need; unknown keys are
// ignored silently.
type posttooluseTraceInput struct {
	ToolName     string          `json:"tool_name"`
	CWD          string          `json:"cwd"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
}

// tracedTools is the allowlist of tools that produce trace events.
// Read-only tools (Read, Grep, Glob) are intentionally excluded so traces
// stay focused on state-changing activity — which is the signal the Advisor
// and code-space-search consumers actually want.
var tracedTools = map[string]bool{
	"Edit":      true,
	"Write":     true,
	"MultiEdit": true,
	"Bash":      true,
	"Task":      true,
}

// wipRowPattern matches a Plans.md v2 table row whose Status column is
// cc:WIP. Capture group 1 is the task id (first column), e.g. "72.3" or
// "72.1.fix". Anchored at "^|" so it only matches table rows, not prose.
var wipRowPattern = regexp.MustCompile(`^\|\s*([0-9]+(?:\.[0-9]+)*(?:\.[a-z]+)?)\s*\|.*\bcc:WIP\b`)

// maxBashArgsSummary caps Bash command text in args_summary per the
// trace.v1 schema privacy guidance (500 chars).
const maxBashArgsSummary = 500

// plansReadCap bounds how many bytes we read from Plans.md. Real plans are
// typically 5–50 KiB; 1 MiB is generous but finite.
const plansReadCap int64 = 1 * 1024 * 1024

// hookInputCap bounds how many bytes we accept on stdin. Hook inputs are
// small JSON objects; 256 KiB covers even pathological tool_response blobs.
const hookInputCap int64 = 256 * 1024

// HandlePostToolUseTrace writes one trace.v1 tool_call event per invocation
// to .claude/state/traces/<task_id>.jsonl, scoped to the currently-active
// Plans.md task (first cc:WIP row). If no task is active, or the tool is
// not in the traced allowlist, or anything goes wrong deriving the
// task_id, the hook returns silently without writing.
//
// This hook is observation-only: it never blocks the tool call, never
// modifies the user's view of the conversation, and swallows all
// non-fatal errors to stderr. A bug here must not cascade into a failed
// tool invocation for the user.
func HandlePostToolUseTrace(r io.Reader, w io.Writer) error {
	body, err := io.ReadAll(io.LimitReader(r, hookInputCap))
	if err != nil || len(body) == 0 {
		return nil
	}

	var input posttooluseTraceInput
	if err := json.Unmarshal(body, &input); err != nil {
		// Malformed input is not a reason to fail the tool call.
		return nil
	}

	if !tracedTools[input.ToolName] {
		return nil
	}

	cwd := input.CWD
	if cwd == "" {
		var cwdErr error
		cwd, cwdErr = os.Getwd()
		if cwdErr != nil || cwd == "" {
			return nil
		}
	}

	taskID, ok := findActiveWIPTask(filepath.Join(cwd, "Plans.md"))
	if !ok {
		return nil
	}

	argsSummary := summarizeToolArgs(input.ToolName, input.ToolInput)
	payload, err := trace.MarshalPayload(map[string]any{
		"tool":         input.ToolName,
		"args_summary": argsSummary,
	})
	if err != nil {
		return nil
	}

	ev := trace.Event{
		TaskID:    taskID,
		EventType: "tool_call",
		Agent:     "worker",
		Payload:   payload,
	}

	if appendErr := trace.NewWriter(cwd).AppendEvent(ev); appendErr != nil {
		fmt.Fprintf(os.Stderr, "[trace-posttool] append: %v\n", appendErr)
	}
	return nil
}

// findActiveWIPTask scans Plans.md for the first table row with cc:WIP in
// its Status column and returns that row's task id. Returns ("", false) if
// the file can't be read, is unreadable JSON-of-Plans, or has no cc:WIP row.
func findActiveWIPTask(plansPath string) (string, bool) {
	f, err := os.Open(plansPath)
	if err != nil {
		return "", false
	}
	defer f.Close()
	body, err := io.ReadAll(io.LimitReader(f, plansReadCap))
	if err != nil {
		return "", false
	}
	for line := range strings.SplitSeq(string(body), "\n") {
		if m := wipRowPattern.FindStringSubmatch(line); m != nil {
			return m[1], true
		}
	}
	return "", false
}

// summarizeToolArgs produces a short, privacy-safe description of a tool
// invocation. File paths are preserved (so traces pinpoint affected files);
// file contents, env values, and full argument payloads never appear.
//
// Shape per tool:
//
//	Edit / MultiEdit  -> "file_path=<path>"
//	Write             -> "file_path=<path> (new)"
//	Bash              -> "cmd=<first 500 chars of command>"
//	Task              -> "subagent=<type>"
//
// Unknown or malformed input yields an empty string; the trace event still
// gets written so the tool_call is observable, just without details.
func summarizeToolArgs(toolName string, toolInput json.RawMessage) string {
	if len(toolInput) == 0 {
		return ""
	}
	var raw map[string]any
	if err := json.Unmarshal(toolInput, &raw); err != nil {
		return ""
	}
	switch toolName {
	case "Edit", "MultiEdit":
		if fp, _ := raw["file_path"].(string); fp != "" {
			return "file_path=" + fp
		}
	case "Write":
		if fp, _ := raw["file_path"].(string); fp != "" {
			return "file_path=" + fp + " (new)"
		}
	case "Bash":
		if cmd, _ := raw["command"].(string); cmd != "" {
			return "cmd=" + truncateSummary(cmd, maxBashArgsSummary)
		}
	case "Task":
		if st, _ := raw["subagent_type"].(string); st != "" {
			return "subagent=" + st
		}
	}
	return ""
}

func truncateSummary(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
