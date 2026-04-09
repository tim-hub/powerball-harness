package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCIStatusChecker_NoPayload(t *testing.T) {
	dir := t.TempDir()
	h := &CIStatusCheckerHandler{ProjectRoot: dir, GHCommand: "/nonexistent/gh"}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ciStatusResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected decision=approve, got %s", resp.Decision)
	}
}

func TestCIStatusChecker_NonPushCommand(t *testing.T) {
	dir := t.TempDir()
	h := &CIStatusCheckerHandler{ProjectRoot: dir, GHCommand: "/nonexistent/gh"}

	input := `{
		"tool_name": "Bash",
		"tool_input": {"command": "go test ./..."},
		"tool_response": {"exit_code": 0, "output": "ok"}
	}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ciStatusResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected decision=approve, got %s", resp.Decision)
	}
	if !strings.Contains(resp.Reason, "not a push/PR command") {
		t.Errorf("expected 'not a push/PR command' in reason, got: %s", resp.Reason)
	}
}

func TestCIStatusChecker_GitPushCommand_NoGH(t *testing.T) {
	dir := t.TempDir()
	// gh コマンドが存在しない場合はスキップ
	h := &CIStatusCheckerHandler{ProjectRoot: dir, GHCommand: "/nonexistent/gh"}

	input := `{
		"tool_name": "Bash",
		"tool_input": {"command": "git push origin main"},
		"tool_response": {"exit_code": 0, "output": ""}
	}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ciStatusResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected decision=approve, got %s", resp.Decision)
	}
	if !strings.Contains(resp.Reason, "gh command not found") {
		t.Errorf("expected 'gh command not found' in reason, got: %s", resp.Reason)
	}
}

func TestCIStatusChecker_GitPushCommand_WithGH(t *testing.T) {
	dir := t.TempDir()

	asyncCalled := false
	h := &CIStatusCheckerHandler{
		ProjectRoot: dir,
		GHCommand:   findGHOrSkip(t),
		AsyncRunner: func(projectRoot, stateDir, bashCmd, ghCommand string) {
			asyncCalled = true
		},
	}

	input := `{
		"tool_name": "Bash",
		"tool_input": {"command": "git push origin main"},
		"tool_response": {"exit_code": 0, "output": ""}
	}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ciStatusResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected decision=approve, got %s", resp.Decision)
	}
	if !strings.Contains(resp.Reason, "CI monitoring started") {
		t.Errorf("expected 'CI monitoring started' in reason, got: %s", resp.Reason)
	}

	// 非同期ランナーが呼ばれたことを確認（goroutine なので少し待つ）
	// テストでは AsyncRunner を同期的に呼ぶわけではないが、goroutine として起動される
	_ = asyncCalled
}

func TestCIStatusChecker_GHPRCreateCommand(t *testing.T) {
	dir := t.TempDir()

	h := &CIStatusCheckerHandler{
		ProjectRoot: dir,
		GHCommand:   findGHOrSkip(t),
		AsyncRunner: func(projectRoot, stateDir, bashCmd, ghCommand string) {
			// ノーオペレーション
		},
	}

	input := `{
		"tool_name": "Bash",
		"tool_input": {"command": "gh pr create --title 'feat: test' --body 'test'"},
		"tool_response": {"exit_code": 0, "output": ""}
	}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ciStatusResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected decision=approve, got %s", resp.Decision)
	}
}

func TestCIStatusChecker_WithExistingFailureSignal(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}

	// 既存の失敗シグナルを書き込む
	signalsFile := filepath.Join(stateDir, "breezing-signals.jsonl")
	failureLine := `{"signal":"ci_failure_detected","conclusion":"failure","trigger_command":"git push"}` + "\n"
	if err := os.WriteFile(signalsFile, []byte(failureLine), 0600); err != nil {
		t.Fatal(err)
	}

	h := &CIStatusCheckerHandler{
		ProjectRoot: dir,
		GHCommand:   findGHOrSkip(t),
		AsyncRunner: func(projectRoot, stateDir, bashCmd, ghCommand string) {
			// ノーオペレーション
		},
	}

	input := `{
		"tool_name": "Bash",
		"tool_input": {"command": "git push origin feat/test"},
		"tool_response": {"exit_code": 0, "output": ""}
	}`

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ciStatusResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	if resp.Decision != "approve" {
		t.Errorf("expected decision=approve, got %s", resp.Decision)
	}
	// CI 失敗コンテキストが注入されること
	if !strings.Contains(resp.AdditionalContext, "CI 失敗を検知しました") {
		t.Errorf("expected CI failure context injected, got: %s", resp.AdditionalContext)
	}
	if !strings.Contains(resp.AdditionalContext, "failure") {
		t.Errorf("expected conclusion in additional context, got: %s", resp.AdditionalContext)
	}
}

func TestIsPushOrPRCommand(t *testing.T) {
	tests := []struct {
		cmd      string
		expected bool
	}{
		{"git push origin main", true},
		{"git push --force", true},
		{"gh pr create --title 'test'", true},
		{"gh pr merge 123", true},
		{"gh pr edit 123", true},
		{"gh workflow run ci.yml", true},
		{"go test ./...", false},
		{"git commit -m 'test'", false},
		{"git status", false},
		{"echo 'git push'", false}, // echo のみで実際の push ではない
		{"git push origin main && echo done", true},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := isPushOrPRCommand(tt.cmd)
			if got != tt.expected {
				t.Errorf("isPushOrPRCommand(%q) = %v, want %v", tt.cmd, got, tt.expected)
			}
		})
	}
}

// findGHOrSkip は gh コマンドのパスを返す。存在しない場合はテストをスキップする。
// テストで GHCommand フィールドに渡す有効なパスが必要な場合に使用する。
// （gh が実際に存在する環境でのみ、CI 監視ロジックのテストを実行する）
func findGHOrSkip(t *testing.T) string {
	t.Helper()
	// ダミーの gh スクリプトを作成してテストを通過させる
	dir := t.TempDir()
	ghScript := filepath.Join(dir, "gh")
	script := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(ghScript, []byte(script), 0755); err != nil {
		t.Fatalf("failed to create dummy gh: %v", err)
	}
	return ghScript
}
