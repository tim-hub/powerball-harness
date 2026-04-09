package hookhandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHandleInboxCheck_EmptyInbox(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	// Create broadcast.md so the handler proceeds past the early-exit check.
	sessionsDir := filepath.Join(dir, ".claude", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	broadcastPath := filepath.Join(sessionsDir, "broadcast.md")
	if err := os.WriteFile(broadcastPath, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	// No inbox file → expect silent (no output).
	var out bytes.Buffer
	if err := HandleInboxCheck(strings.NewReader("{}"), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for empty inbox, got: %s", out.String())
	}
}

func TestHandleInboxCheck_WithUnreadMessages(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	sessionsDir := filepath.Join(dir, ".claude", "sessions")
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// broadcast.md must exist.
	broadcastPath := filepath.Join(sessionsDir, "broadcast.md")
	if err := os.WriteFile(broadcastPath, []byte("exists"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write two unread messages to inbox JSONL.
	inboxPath := filepath.Join(stateDir, "session-inbox.jsonl")
	content := `{"read":false,"msg":"message one"}` + "\n" +
		`{"read":false,"msg":"message two"}` + "\n"
	if err := os.WriteFile(inboxPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := HandleInboxCheck(strings.NewReader("{}"), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Len() == 0 {
		t.Fatal("expected output for unread messages, got nothing")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}

	hso, ok := result["hookSpecificOutput"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing hookSpecificOutput field")
	}
	if hso["hookEventName"] != "PreToolUse" {
		t.Errorf("hookEventName = %v, want PreToolUse", hso["hookEventName"])
	}
	if hso["permissionDecision"] != "allow" {
		t.Errorf("permissionDecision = %v, want allow", hso["permissionDecision"])
	}
	ctx, _ := hso["additionalContext"].(string)
	if !strings.Contains(ctx, "message one") {
		t.Errorf("additionalContext does not contain expected message: %s", ctx)
	}
}

func TestHandleInboxCheck_Throttle(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	sessionsDir := filepath.Join(dir, ".claude", "sessions")
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// broadcast.md must exist.
	if err := os.WriteFile(filepath.Join(sessionsDir, "broadcast.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write an unread message.
	inboxPath := filepath.Join(stateDir, "session-inbox.jsonl")
	if err := os.WriteFile(inboxPath, []byte(`{"read":false,"msg":"hello"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Write a recent last-check timestamp (now − 1 minute → within throttle window).
	recent := time.Now().Add(-1 * time.Minute).Unix()
	checkFile := filepath.Join(sessionsDir, ".last_inbox_check")
	tsStr := fmt.Sprintf("%d", recent)
	if err := os.WriteFile(checkFile, []byte(tsStr+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := HandleInboxCheck(strings.NewReader("{}"), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Within throttle window → expect no output.
	if out.Len() != 0 {
		t.Errorf("expected no output within throttle window, got: %s", out.String())
	}
}

func TestHandleInboxCheck_ReadMessages_Filtered(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	sessionsDir := filepath.Join(dir, ".claude", "sessions")
	stateDir := filepath.Join(dir, ".claude", "state")
	os.MkdirAll(sessionsDir, 0o755)  //nolint:errcheck
	os.MkdirAll(stateDir, 0o755)     //nolint:errcheck
	os.WriteFile(filepath.Join(sessionsDir, "broadcast.md"), []byte("x"), 0o644) //nolint:errcheck

	// Mix of read and unread messages.
	inboxPath := filepath.Join(stateDir, "session-inbox.jsonl")
	content := `{"read":true,"msg":"already read"}` + "\n" +
		`{"read":false,"msg":"unread msg"}` + "\n"
	os.WriteFile(inboxPath, []byte(content), 0o644) //nolint:errcheck

	var out bytes.Buffer
	HandleInboxCheck(strings.NewReader("{}"), &out) //nolint:errcheck

	if out.Len() == 0 {
		t.Fatal("expected output for unread message")
	}
	outStr := out.String()
	if strings.Contains(outStr, "already read") {
		t.Error("output should not contain already-read messages")
	}
	if !strings.Contains(outStr, "unread msg") {
		t.Error("output should contain unread message")
	}
}

func TestNoBroadcastFile_NoOutput(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	// No broadcast.md → early exit, no output.
	var out bytes.Buffer
	if err := HandleInboxCheck(strings.NewReader("{}"), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output when broadcast.md absent, got: %s", out.String())
	}
}

