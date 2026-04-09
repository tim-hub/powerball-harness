package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorktreeRemoveHandler_EmptyPayload(t *testing.T) {
	h := &WorktreeRemoveHandler{}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp worktreeRemoveResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected decision=approve, got %q", resp.Decision)
	}
	if !strings.Contains(resp.Reason, "no payload") {
		t.Errorf("expected 'no payload' in reason, got %q", resp.Reason)
	}
}

func TestWorktreeRemoveHandler_NoSessionID(t *testing.T) {
	h := &WorktreeRemoveHandler{}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{"cwd":"/tmp/test"}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp worktreeRemoveResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected decision=approve, got %q", resp.Decision)
	}
	if !strings.Contains(resp.Reason, "no session_id") {
		t.Errorf("expected 'no session_id' in reason, got %q", resp.Reason)
	}
}

func TestWorktreeRemoveHandler_CleansWorktreeInfo(t *testing.T) {
	dir := t.TempDir()

	// worktree-info.json を作成
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	infoFile := filepath.Join(stateDir, "worktree-info.json")
	if err := os.WriteFile(infoFile, []byte(`{"worktree":"test"}`), 0600); err != nil {
		t.Fatal(err)
	}

	h := &WorktreeRemoveHandler{}
	input := `{"session_id":"sess-001","cwd":"` + dir + `"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp worktreeRemoveResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected decision=approve, got %q", resp.Decision)
	}
	if !strings.Contains(resp.Reason, "cleaned up") {
		t.Errorf("expected 'cleaned up' in reason, got %q", resp.Reason)
	}

	// worktree-info.json が削除されているか確認
	if _, err := os.Stat(infoFile); err == nil {
		t.Errorf("expected worktree-info.json to be deleted")
	}
}

func TestWorktreeRemoveHandler_NoCWD(t *testing.T) {
	h := &WorktreeRemoveHandler{}
	input := `{"session_id":"sess-001"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp worktreeRemoveResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected decision=approve, got %q", resp.Decision)
	}
}

func TestWorktreeRemoveHandler_MissingWorktreeInfo(t *testing.T) {
	dir := t.TempDir()
	// .claude/state ディレクトリを作らず worktree-info.json なし
	h := &WorktreeRemoveHandler{}
	input := `{"session_id":"sess-002","cwd":"` + dir + `"}`

	var out bytes.Buffer
	// ファイルが存在しなくてもエラーにならないこと
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp worktreeRemoveResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected decision=approve, got %q", resp.Decision)
	}
}
