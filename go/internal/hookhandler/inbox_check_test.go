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

func TestReadBroadcastMessages_MarkdownFormat(t *testing.T) {
	dir := t.TempDir()
	broadcastPath := filepath.Join(dir, "broadcast.md")

	// bash 版 session-inbox-check.sh が生成するマークダウン形式
	content := "## 2026-04-09T12:00:00Z [abc123456def]\nhello from session A\n\n## 2026-04-09T12:05:00Z [xyz789012abc]\nupdate: task completed\n"
	if err := os.WriteFile(broadcastPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	msgs, err := readBroadcastMessages(broadcastPath, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d: %v", len(msgs), msgs)
	}
	if !strings.Contains(msgs[0], "hello from session A") {
		t.Errorf("message 0 should contain 'hello from session A', got: %s", msgs[0])
	}
	if !strings.Contains(msgs[1], "update: task completed") {
		t.Errorf("message 1 should contain 'update: task completed', got: %s", msgs[1])
	}
	// タイムスタンプは HH:MM 形式で含まれるはず
	if !strings.Contains(msgs[0], "[12:00]") {
		t.Errorf("message 0 should contain '[12:00]', got: %s", msgs[0])
	}
}

func TestReadBroadcastMessages_MaxCount(t *testing.T) {
	dir := t.TempDir()
	broadcastPath := filepath.Join(dir, "broadcast.md")

	// 5件以上のメッセージ
	var content string
	for i := 0; i < 8; i++ {
		content += fmt.Sprintf("## 2026-04-09T12:0%dZ [sender%d]\nmessage %d\n\n", i, i, i)
	}
	if err := os.WriteFile(broadcastPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	msgs, err := readBroadcastMessages(broadcastPath, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) > 5 {
		t.Errorf("expected at most 5 messages, got %d", len(msgs))
	}
}

func TestHandleInboxCheck_BroadcastMdSource(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HARNESS_PROJECT_ROOT", dir)

	sessionsDir := filepath.Join(dir, ".claude", "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// broadcast.md にメッセージを書き込む（bash 版と同じソース）
	broadcastPath := filepath.Join(sessionsDir, "broadcast.md")
	content := "## 2026-04-09T10:30:00Z [remote-session-a1]\nplease check the CI status\n"
	if err := os.WriteFile(broadcastPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := HandleInboxCheck(strings.NewReader("{}"), &out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Len() == 0 {
		t.Fatal("expected output for broadcast.md message, got nothing")
	}
	outStr := out.String()
	if !strings.Contains(outStr, "CI status") {
		t.Errorf("output should contain 'CI status', got: %s", outStr)
	}
}

