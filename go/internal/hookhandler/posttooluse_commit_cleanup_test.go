package hookhandler

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommitCleanupHandler_EmptyInput(t *testing.T) {
	h := &CommitCleanupHandler{}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// サイレントクリーンアップ: 出力なし
	if out.Len() != 0 {
		t.Errorf("expected no output, got %q", out.String())
	}
}

func TestCommitCleanupHandler_NotBashTool(t *testing.T) {
	h := &CommitCleanupHandler{}
	input := `{"tool_name":"Read","tool_input":{"command":"git commit -m test"}}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for non-Bash tool, got %q", out.String())
	}
}

func TestCommitCleanupHandler_NotGitCommit(t *testing.T) {
	h := &CommitCleanupHandler{}
	input := `{"tool_name":"Bash","tool_input":{"command":"git status"},"tool_result":"On branch main"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output for non-commit command, got %q", out.String())
	}
}

func TestCommitCleanupHandler_GitCommitSuccess_ClearsFiles(t *testing.T) {
	dir := t.TempDir()

	// レビューファイルを作成
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	reviewState := filepath.Join(stateDir, "review-approved.json")
	reviewResult := filepath.Join(stateDir, "review-result.json")
	if err := os.WriteFile(reviewState, []byte(`{"approved":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(reviewResult, []byte(`{"verdict":"APPROVE"}`), 0600); err != nil {
		t.Fatal(err)
	}

	h := &CommitCleanupHandler{ProjectRoot: dir}
	input := `{"tool_name":"Bash","tool_input":{"command":"git commit -m 'feat: add feature'"},"tool_result":"[main abc1234] feat: add feature"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ファイルが削除されているか確認
	if _, err := os.Stat(reviewState); err == nil {
		t.Errorf("expected review-approved.json to be deleted")
	}
	if _, err := os.Stat(reviewResult); err == nil {
		t.Errorf("expected review-result.json to be deleted")
	}
}

func TestCommitCleanupHandler_GitCommitError_KeepsFiles(t *testing.T) {
	dir := t.TempDir()

	// レビューファイルを作成
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	reviewState := filepath.Join(stateDir, "review-approved.json")
	if err := os.WriteFile(reviewState, []byte(`{"approved":true}`), 0600); err != nil {
		t.Fatal(err)
	}

	h := &CommitCleanupHandler{ProjectRoot: dir}
	// エラーを含む tool_result
	input := `{"tool_name":"Bash","tool_input":{"command":"git commit -m test"},"tool_result":"error: nothing to commit, working tree clean"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// エラー時はファイルを保持
	if _, err := os.Stat(reviewState); err != nil {
		t.Errorf("expected review-approved.json to be kept on commit error")
	}
}

func TestCommitCleanupHandler_NoReviewFiles_NoError(t *testing.T) {
	dir := t.TempDir()

	h := &CommitCleanupHandler{ProjectRoot: dir}
	input := `{"tool_name":"Bash","tool_input":{"command":"git commit -m test"},"tool_result":"[main abc1234] test"}`

	var out bytes.Buffer
	// レビューファイルが存在しなくてもエラーにならないこと
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestIsGitCommitCommand(t *testing.T) {
	tests := []struct {
		command string
		want    bool
	}{
		{"git commit -m test", true},
		{"git commit", true},
		{"git commit --amend", true},
		{"  git commit -m 'message'", true},
		{"git status", false},
		{"git checkout main", false},
		{"echo 'git commit'", false},
		{"notgit commit", false},
		{"git commitish", false},
	}

	for _, tt := range tests {
		got := isGitCommitCommand(tt.command)
		if got != tt.want {
			t.Errorf("isGitCommitCommand(%q) = %v, want %v", tt.command, got, tt.want)
		}
	}
}

func TestContainsErrorIndicator(t *testing.T) {
	tests := []struct {
		result string
		want   bool
	}{
		{"[main abc] feat: done", false},
		{"error: nothing to commit", true},
		{"fatal: not a git repository", true},
		{"nothing to commit, working tree clean", true},
		{"failed to write", true},
		{"FAILED tests", true},
		{"", false},
	}

	for _, tt := range tests {
		got := containsErrorIndicator(tt.result)
		if got != tt.want {
			t.Errorf("containsErrorIndicator(%q) = %v, want %v", tt.result, got, tt.want)
		}
	}
}

func TestCommitCleanupHandler_StderrMessage(t *testing.T) {
	dir := t.TempDir()

	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "review-approved.json"), []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}

	h := &CommitCleanupHandler{ProjectRoot: dir}
	input := `{"tool_name":"Bash","tool_input":{"command":"git commit -m ok"},"tool_result":"[main 1234567] ok"}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// ログメッセージが出力されること
	if !strings.Contains(out.String(), "レビュー承認状態をクリア") {
		t.Errorf("expected cleanup log message, got %q", out.String())
	}
}
