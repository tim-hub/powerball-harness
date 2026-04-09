package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeFixProposalFile は pending-fix-proposals.jsonl テスト用ファイルを作成する。
func makeFixProposalFile(t *testing.T, dir string, proposals []fixProposal) string {
	t.Helper()
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(stateDir, pendingFixProposalsFile)
	var sb strings.Builder
	for _, p := range proposals {
		line, _ := json.Marshal(p)
		sb.WriteString(string(line))
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

// sampleProposal はテスト用の fixProposal を返す。
func sampleProposal() fixProposal {
	return fixProposal{
		SourceTaskID:    "26",
		FixTaskID:       "26.fix",
		ProposalSubject: "fix: テスト修正タスク",
		DoD:             "テストが通ること",
		Depends:         "26",
		FailureCategory: "assertion_error",
		RecommendedAction: "アサーションを修正する",
		Status:          "pending",
	}
}

// TestFixProposalInjector_EmptyInput は空入力で何も出力しないことを確認する。
func TestFixProposalInjector_EmptyInput(t *testing.T) {
	dir := t.TempDir()
	h := &FixProposalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 空入力では何も出力しない
	if out.Len() > 0 {
		t.Errorf("expected no output for empty input, got: %s", out.String())
	}
}

// TestFixProposalInjector_NoPendingFile はファイルなしで何も出力しないことを確認する。
func TestFixProposalInjector_NoPendingFile(t *testing.T) {
	dir := t.TempDir()
	h := &FixProposalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{"prompt":"hello"}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() > 0 {
		t.Errorf("expected no output without proposals file, got: %s", out.String())
	}
}

// TestFixProposalInjector_Reminder は未処理の proposal がリマインダーとして表示されることを確認する。
func TestFixProposalInjector_Reminder(t *testing.T) {
	dir := t.TempDir()
	makeFixProposalFile(t, dir, []fixProposal{sampleProposal()})

	h := &FixProposalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{"prompt":"次の作業をして"}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp fixProposalInjectorOutput
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	if !strings.Contains(resp.SystemMessage, "FIX PROPOSAL") {
		t.Errorf("expected FIX PROPOSAL reminder, got: %s", resp.SystemMessage)
	}
	if !strings.Contains(resp.SystemMessage, "26.fix") {
		t.Errorf("expected fix task ID in message, got: %s", resp.SystemMessage)
	}
	if !strings.Contains(resp.SystemMessage, "approve fix 26") {
		t.Errorf("expected approve instruction, got: %s", resp.SystemMessage)
	}
}

// TestFixProposalInjector_Approve は "approve fix" コマンドで Plans.md に反映されることを確認する。
func TestFixProposalInjector_Approve(t *testing.T) {
	dir := t.TempDir()
	makeFixProposalFile(t, dir, []fixProposal{sampleProposal()})

	// Plans.md を作成
	plansContent := "| Task | 内容 | DoD | Depends | Status |\n|------|------|-----|---------|--------|\n| 26 | 元タスク | 完了 | - | cc:完了 |\n"
	plansPath := filepath.Join(dir, "Plans.md")
	if err := os.WriteFile(plansPath, []byte(plansContent), 0644); err != nil {
		t.Fatal(err)
	}

	h := &FixProposalInjectorHandler{ProjectRoot: dir, PlansPath: plansPath}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{"prompt":"approve fix 26"}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp fixProposalInjectorOutput
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	// 成功メッセージを確認
	if !strings.Contains(resp.SystemMessage, "fix proposal を反映しました") {
		t.Errorf("expected success message, got: %s", resp.SystemMessage)
	}

	// Plans.md に fix タスクが追加されていること
	plansData, err := os.ReadFile(plansPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(plansData), "26.fix") {
		t.Errorf("expected fix task in Plans.md, got: %s", string(plansData))
	}
	if !strings.Contains(string(plansData), "cc:TODO") {
		t.Errorf("expected cc:TODO status in Plans.md, got: %s", string(plansData))
	}

	// proposals ファイルから削除されていること
	proposals, _ := loadPendingFixProposals(filepath.Join(dir, ".claude", "state", pendingFixProposalsFile))
	if len(proposals) != 0 {
		t.Errorf("expected proposal to be consumed, but %d remain", len(proposals))
	}
}

// TestFixProposalInjector_Reject は "reject fix" コマンドで proposal が削除されることを確認する。
func TestFixProposalInjector_Reject(t *testing.T) {
	dir := t.TempDir()
	makeFixProposalFile(t, dir, []fixProposal{sampleProposal()})

	h := &FixProposalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{"prompt":"reject fix 26"}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp fixProposalInjectorOutput
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	if !strings.Contains(resp.SystemMessage, "却下しました") {
		t.Errorf("expected rejection message, got: %s", resp.SystemMessage)
	}

	// proposals ファイルから削除されていること
	proposals, _ := loadPendingFixProposals(filepath.Join(dir, ".claude", "state", pendingFixProposalsFile))
	if len(proposals) != 0 {
		t.Errorf("expected proposal to be consumed, but %d remain", len(proposals))
	}
}

// TestFixProposalInjector_YesApprove は "yes" で単一 proposal が承認されることを確認する。
func TestFixProposalInjector_YesApprove(t *testing.T) {
	dir := t.TempDir()
	makeFixProposalFile(t, dir, []fixProposal{sampleProposal()})

	// Plans.md を作成
	plansContent := "| Task | 内容 | DoD | Depends | Status |\n|------|------|-----|---------|--------|\n| 26 | 元タスク | 完了 | - | cc:完了 |\n"
	plansPath := filepath.Join(dir, "Plans.md")
	if err := os.WriteFile(plansPath, []byte(plansContent), 0644); err != nil {
		t.Fatal(err)
	}

	h := &FixProposalInjectorHandler{ProjectRoot: dir, PlansPath: plansPath}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{"prompt":"yes"}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp fixProposalInjectorOutput
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	if !strings.Contains(resp.SystemMessage, "反映しました") {
		t.Errorf("expected success for 'yes' command, got: %s", resp.SystemMessage)
	}
}

// TestFixProposalInjector_MultipleProposals_RequiresID は複数 proposal がある場合に ID 指定が必要なことを確認する。
func TestFixProposalInjector_MultipleProposals_RequiresID(t *testing.T) {
	dir := t.TempDir()
	p2 := fixProposal{
		SourceTaskID:    "27",
		FixTaskID:       "27.fix",
		ProposalSubject: "fix: 別タスク",
		Status:          "pending",
	}
	makeFixProposalFile(t, dir, []fixProposal{sampleProposal(), p2})

	h := &FixProposalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{"prompt":"yes"}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp fixProposalInjectorOutput
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	// "yes" は複数 proposal では使えない
	if !strings.Contains(resp.SystemMessage, "明示してください") {
		t.Errorf("expected disambiguation message for multiple proposals, got: %s", resp.SystemMessage)
	}
}

// TestFixProposalInjector_NotFoundID は存在しない ID を指定した場合のエラーを確認する。
func TestFixProposalInjector_NotFoundID(t *testing.T) {
	dir := t.TempDir()
	makeFixProposalFile(t, dir, []fixProposal{sampleProposal()})

	h := &FixProposalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{"prompt":"approve fix 99"}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp fixProposalInjectorOutput
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	if !strings.Contains(resp.SystemMessage, "見つかりません") {
		t.Errorf("expected not-found message, got: %s", resp.SystemMessage)
	}
}

// TestParseFixProposalAction は action 解析のロジックを確認する。
func TestParseFixProposalAction(t *testing.T) {
	tests := []struct {
		lower    string
		original string
		action   string
		targetID string
	}{
		{"approve fix 26", "approve fix 26", "approve", "26"},
		{"approve fix", "approve fix", "approve", ""},
		{"reject fix 27.fix", "reject fix 27.fix", "reject", "27.fix"},
		{"yes", "yes", "approve", ""},
		{"no", "no", "reject", ""},
		{"はい", "はい", "approve", ""},
		{"却下", "却下", "reject", ""},
		{"hello", "hello", "", ""},
	}
	for _, tc := range tests {
		action, targetID := parseFixProposalAction(tc.lower, tc.original)
		if action != tc.action {
			t.Errorf("parseFixProposalAction(%q): action=%q, want %q", tc.lower, action, tc.action)
		}
		if targetID != tc.targetID {
			t.Errorf("parseFixProposalAction(%q): targetID=%q, want %q", tc.lower, targetID, tc.targetID)
		}
	}
}

// TestApplyFixProposalToPlans は Plans.md への挿入ロジックを確認する。
func TestApplyFixProposalToPlans(t *testing.T) {
	dir := t.TempDir()
	plansPath := filepath.Join(dir, "Plans.md")

	content := "| Task | 内容 | DoD | Depends | Status |\n|------|------|-----|---------|--------|\n| 26 | 元タスク | 完了 | - | cc:完了 |\n| 28 | 別タスク | - | - | cc:TODO |\n"
	if err := os.WriteFile(plansPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	proposal := sampleProposal()
	result := applyFixProposalToPlans(plansPath, proposal)
	if result != "applied" {
		t.Fatalf("expected applied, got %s", result)
	}

	data, _ := os.ReadFile(plansPath)
	text := string(data)

	// 26.fix が 26 の直後に挿入されていること
	lines := strings.Split(text, "\n")
	found26 := -1
	found26fix := -1
	for i, line := range lines {
		// 正確に "| 26 |" にマッチ（"| 26.fix |" は除外）
		if strings.Contains(line, "| 26 |") && !strings.Contains(line, "26.fix") {
			found26 = i
		}
		if strings.Contains(line, "| 26.fix |") {
			found26fix = i
		}
	}
	if found26 < 0 {
		t.Fatal("original task 26 not found in Plans.md")
	}
	if found26fix < 0 {
		t.Fatal("fix task 26.fix not found in Plans.md")
	}
	if found26fix != found26+1 {
		t.Errorf("expected 26.fix to be right after 26, but 26 is at line %d and 26.fix is at line %d", found26, found26fix)
	}

	// 2回目は already_present
	result2 := applyFixProposalToPlans(plansPath, proposal)
	if result2 != "already_present" {
		t.Errorf("expected already_present on second apply, got %s", result2)
	}
}

// TestConsumeFixProposal は consume で指定 proposal が削除されることを確認する。
func TestConsumeFixProposal(t *testing.T) {
	dir := t.TempDir()
	p1 := sampleProposal()
	p2 := fixProposal{SourceTaskID: "27", FixTaskID: "27.fix", Status: "pending"}
	path := makeFixProposalFile(t, dir, []fixProposal{p1, p2})

	if err := consumeFixProposal(path, "26"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	remaining, err := loadPendingFixProposals(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 remaining, got %d", len(remaining))
	}
	if remaining[0].SourceTaskID != "27" {
		t.Errorf("expected task 27 to remain, got %s", remaining[0].SourceTaskID)
	}
}
