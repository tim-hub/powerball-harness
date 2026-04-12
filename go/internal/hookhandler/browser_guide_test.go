package hookhandler

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestHandleBrowserGuide_EmptyInput(t *testing.T) {
	// Empty stdin → no output (matches bash `[ -z "$INPUT" ] && exit 0`).
	var out bytes.Buffer
	if err := HandleBrowserGuide(strings.NewReader(""), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for empty stdin, got: %s", out.String())
	}
}

func TestHandleBrowserGuide_AgentBrowserNotInstalled(t *testing.T) {
	// When agent-browser is not on PATH, the handler should produce no output.
	// We cannot easily control PATH in unit tests, so we test the internal
	// agentBrowserInstalled() helper separately and mock the function.

	// Save original and restore.
	orig := agentBrowserLookupFn
	defer func() { agentBrowserLookupFn = orig }()

	// Simulate not installed.
	agentBrowserLookupFn = func() bool { return false }

	var out bytes.Buffer
	if err := HandleBrowserGuide(strings.NewReader(`{"tool_name":"mcp__playwright__navigate"}`), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output when agent-browser not installed, got: %s", out.String())
	}
}

func TestHandleBrowserGuide_AgentBrowserInstalled(t *testing.T) {
	orig := agentBrowserLookupFn
	defer func() { agentBrowserLookupFn = orig }()

	// Simulate installed.
	agentBrowserLookupFn = func() bool { return true }

	var out bytes.Buffer
	if err := HandleBrowserGuide(strings.NewReader(`{"tool_name":"mcp__playwright__navigate"}`), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("expected output when agent-browser is installed")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}

	hso, ok := result["hookSpecificOutput"].(map[string]interface{})
	if !ok {
		t.Fatal("missing hookSpecificOutput field")
	}
	if hso["hookEventName"] != "PreToolUse" {
		t.Errorf("hookEventName = %v, want PreToolUse", hso["hookEventName"])
	}
	ctx, _ := hso["additionalContext"].(string)
	if !strings.Contains(ctx, "agent-browser") {
		t.Errorf("additionalContext does not mention agent-browser: %s", ctx)
	}
}
