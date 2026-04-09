package hookhandler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FixProposalInjectorHandler は UserPromptSubmit フックハンドラ。
// pending-fix-proposals.jsonl を読み込み、未表示の提案をユーザーに通知する。
// "approve fix" / "reject fix" コマンドを解釈して Plans.md に反映する。
//
// shell 版: scripts/hook-handlers/fix-proposal-injector.sh
type FixProposalInjectorHandler struct {
	// ProjectRoot はプロジェクトルートのパス。空の場合は cwd を使用する。
	ProjectRoot string
	// PlansPath は Plans.md のパス。空の場合は ProjectRoot/Plans.md を使用する。
	PlansPath string
}

// fixProposalInjectorInput は UserPromptSubmit フックの stdin JSON。
type fixProposalInjectorInput struct {
	Prompt string `json:"prompt"`
}

// fixProposalInjectorOutput は UserPromptSubmit フックのレスポンス。
// systemMessage でユーザーに通知する。
type fixProposalInjectorOutput struct {
	SystemMessage string `json:"systemMessage,omitempty"`
}

const (
	pendingFixProposalsFile = "pending-fix-proposals.jsonl"
	fixProposalMaxLines     = 500
)

// Handle は stdin から UserPromptSubmit ペイロードを読み取り、
// fix proposal を通知・処理する。
func (h *FixProposalInjectorHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)
	if len(data) == 0 {
		return nil
	}

	var inp fixProposalInjectorInput
	if err := json.Unmarshal(data, &inp); err != nil {
		return nil
	}

	projectRoot := h.resolveProjectRoot()
	stateDir := filepath.Join(projectRoot, ".claude", "state")
	proposalsFile := filepath.Join(stateDir, pendingFixProposalsFile)

	// proposals ファイルが存在しない場合はスキップ
	if _, err := os.Stat(proposalsFile); os.IsNotExist(err) {
		return nil
	}

	// symlink チェック（isSymlink は notification_handler.go で定義済み）
	if hasFixSymlinkComponent(stateDir, projectRoot) || isSymlink(proposalsFile) {
		return writeFixProposalJSON(w, fixProposalInjectorOutput{
			SystemMessage: "⚠️ fix proposal state path が symlink のため処理を中止しました。",
		})
	}

	plansPath := h.resolvePlansPath(projectRoot)
	if _, err := os.Stat(plansPath); err == nil {
		if hasFixSymlinkComponent(plansPath, projectRoot) {
			return writeFixProposalJSON(w, fixProposalInjectorOutput{
				SystemMessage: "⚠️ Plans.md path が symlink のため fix proposal を反映できません。",
			})
		}
	}

	// プロンプトを解析してアクションを決定
	firstLine := strings.TrimSpace(strings.SplitN(inp.Prompt, "\n", 2)[0])
	lower := strings.ToLower(firstLine)
	action, targetID := parseFixProposalAction(lower, firstLine)

	// pending な proposals を読み込む
	proposals, err := loadPendingFixProposals(proposalsFile)
	if err != nil || len(proposals) == 0 {
		return nil
	}

	pendingCount := len(proposals)

	// アクションがあり、target が未指定で複数 proposal がある場合はエラー
	if action != "" && targetID == "" && pendingCount != 1 {
		return writeFixProposalJSON(w, fixProposalInjectorOutput{
			SystemMessage: fmt.Sprintf(
				"⚠️ 未処理の fix proposal が %d 件あります。approve fix <task_id> または reject fix <task_id> を使って対象を明示してください。",
				pendingCount,
			),
		})
	}

	// 対象 proposal を選択
	proposal, found := selectFixProposal(proposals, targetID)
	if !found {
		if targetID != "" {
			return writeFixProposalJSON(w, fixProposalInjectorOutput{
				SystemMessage: fmt.Sprintf("⚠️ 指定された fix proposal が見つかりません: %s", targetID),
			})
		}
		return nil
	}

	// approve 処理
	if action == "approve" {
		applyResult := applyFixProposalToPlans(plansPath, proposal)
		if applyResult == "applied" || applyResult == "already_present" {
			if err := consumeFixProposal(proposalsFile, proposal.SourceTaskID); err != nil {
				_ = err
			}
			return writeFixProposalJSON(w, fixProposalInjectorOutput{
				SystemMessage: fmt.Sprintf("✅ fix proposal を反映しました: %s\n内容: %s", proposal.FixTaskID, proposal.ProposalSubject),
			})
		} else if applyResult == "plans_missing" {
			return writeFixProposalJSON(w, fixProposalInjectorOutput{
				SystemMessage: "⚠️ fix proposal を反映できませんでした。Plans.md が見つかりません。",
			})
		} else {
			return writeFixProposalJSON(w, fixProposalInjectorOutput{
				SystemMessage: fmt.Sprintf("⚠️ fix proposal の反映に失敗しました。対象タスク %s が Plans.md で見つかりません。", proposal.SourceTaskID),
			})
		}
	}

	// reject 処理
	if action == "reject" {
		_ = consumeFixProposal(proposalsFile, proposal.SourceTaskID)
		return writeFixProposalJSON(w, fixProposalInjectorOutput{
			SystemMessage: fmt.Sprintf("ℹ️ fix proposal を却下しました: %s", proposal.FixTaskID),
		})
	}

	// アクションなし → リマインダーを表示
	reminder := buildFixProposalReminder(proposal, pendingCount)
	return writeFixProposalJSON(w, fixProposalInjectorOutput{SystemMessage: reminder})
}

// resolveProjectRoot はプロジェクトルートを解決する。
func (h *FixProposalInjectorHandler) resolveProjectRoot() string {
	if h.ProjectRoot != "" {
		return h.ProjectRoot
	}
	wd, _ := os.Getwd()
	return wd
}

// resolvePlansPath は Plans.md のパスを解決する。
func (h *FixProposalInjectorHandler) resolvePlansPath(projectRoot string) string {
	if h.PlansPath != "" {
		return h.PlansPath
	}
	return filepath.Join(projectRoot, "Plans.md")
}

// parseFixProposalAction はプロンプト行からアクションと対象 ID を解析する。
func parseFixProposalAction(lower, original string) (action, targetID string) {
	switch {
	case lower == "approve fix" || strings.HasPrefix(lower, "approve fix "):
		action = "approve"
		re := regexp.MustCompile(`(?i)^approve fix\s*(.*)$`)
		if m := re.FindStringSubmatch(original); m != nil {
			targetID = strings.TrimSpace(m[1])
		}
	case lower == "reject fix" || strings.HasPrefix(lower, "reject fix "):
		action = "reject"
		re := regexp.MustCompile(`(?i)^reject fix\s*(.*)$`)
		if m := re.FindStringSubmatch(original); m != nil {
			targetID = strings.TrimSpace(m[1])
		}
	case lower == "yes" || lower == "はい" || lower == "承認":
		action = "approve"
	case lower == "no" || lower == "いいえ" || lower == "却下":
		action = "reject"
	}
	return action, targetID
}

// loadPendingFixProposals は JSONL ファイルから status=pending の fixProposal を読み込む。
// fixProposal 型は task_completed_escalation.go で定義されている。
func loadPendingFixProposals(path string) ([]fixProposal, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var result []fixProposal
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var p fixProposal
		if err := json.Unmarshal([]byte(line), &p); err != nil {
			continue
		}
		if p.Status == "" || p.Status == "pending" {
			result = append(result, p)
		}
	}
	return result, scanner.Err()
}

// selectFixProposal は proposals から selector に一致する fixProposal を返す。
// selector が空の場合は最初の fixProposal を返す。
func selectFixProposal(proposals []fixProposal, selector string) (fixProposal, bool) {
	if len(proposals) == 0 {
		return fixProposal{}, false
	}
	if selector == "" {
		return proposals[0], true
	}
	for _, p := range proposals {
		if p.SourceTaskID == selector || p.FixTaskID == selector {
			return p, true
		}
	}
	return fixProposal{}, false
}

// consumeFixProposal は JSONL から指定 source_task_id の行を削除する。
func consumeFixProposal(path, sourceTaskID string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	var remaining []fixProposal
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var p fixProposal
		if err := json.Unmarshal([]byte(line), &p); err != nil {
			continue
		}
		if p.SourceTaskID == sourceTaskID {
			continue // 削除対象をスキップ
		}
		remaining = append(remaining, p)
	}
	f.Close()
	if err := scanner.Err(); err != nil {
		return err
	}

	// ファイルを書き直す（JSONL ローテーション: 500 行超なら末尾から切り捨て）
	if len(remaining) > fixProposalMaxLines {
		remaining = remaining[len(remaining)-fixProposalMaxLines:]
	}

	tmp := path + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	for _, p := range remaining {
		line, _ := json.Marshal(p)
		_, _ = fmt.Fprintf(out, "%s\n", line)
	}
	out.Close()
	return os.Rename(tmp, path)
}

// applyFixProposalToPlans は proposal を Plans.md の source_task_id 行の直後に挿入する。
// 戻り値: "applied" / "already_present" / "plans_missing" / "source_not_found"
func applyFixProposalToPlans(plansPath string, proposal fixProposal) string {
	rawData, err := os.ReadFile(plansPath)
	if err != nil {
		return "plans_missing"
	}

	text := string(rawData)

	// fix_task_id が既に存在するか確認
	fixPattern := regexp.MustCompile(`(?m)^\|\s*` + regexp.QuoteMeta(proposal.FixTaskID) + `\s*\|`)
	if fixPattern.MatchString(text) {
		return "already_present"
	}

	// source_task_id の行を探して直後に挿入
	sourcePattern := regexp.MustCompile(`(?m)^\|\s*` + regexp.QuoteMeta(proposal.SourceTaskID) + `\s*\|`)

	subject := strings.ReplaceAll(proposal.ProposalSubject, "|", "/")
	dod := strings.ReplaceAll(proposal.DoD, "|", "/")
	depends := strings.ReplaceAll(proposal.Depends, "|", "/")
	newRow := fmt.Sprintf("| %s | %s | %s | %s | cc:TODO |", proposal.FixTaskID, subject, dod, depends)

	lines := strings.Split(text, "\n")
	inserted := false
	result := make([]string, 0, len(lines)+1)
	for _, line := range lines {
		result = append(result, line)
		if !inserted && sourcePattern.MatchString(line) {
			result = append(result, newRow)
			inserted = true
		}
	}

	if !inserted {
		return "source_not_found"
	}

	content := strings.Join(result, "\n")
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	tmp := plansPath + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return "source_not_found"
	}
	if err := os.Rename(tmp, plansPath); err != nil {
		_ = os.Remove(tmp)
		return "source_not_found"
	}
	return "applied"
}

// buildFixProposalReminder はリマインダーメッセージを構築する。
func buildFixProposalReminder(proposal fixProposal, pendingCount int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[FIX PROPOSAL] 未処理の修正タスク案があります (%d件)\n", pendingCount))
	sb.WriteString(fmt.Sprintf("対象: %s — %s\n", proposal.FixTaskID, proposal.ProposalSubject))
	sb.WriteString(fmt.Sprintf("失敗カテゴリ: %s\n", proposal.FailureCategory))
	sb.WriteString(fmt.Sprintf("DoD: %s\n", proposal.DoD))
	if proposal.RecommendedAction != "" {
		sb.WriteString(fmt.Sprintf("推奨アクション: %s\n", proposal.RecommendedAction))
	}
	sb.WriteString(fmt.Sprintf("承認: approve fix %s\n", proposal.SourceTaskID))
	sb.WriteString(fmt.Sprintf("却下: reject fix %s", proposal.SourceTaskID))
	return sb.String()
}

// hasFixSymlinkComponent はパスがプロジェクトルート内でシンボリックリンクコンポーネントを含むか確認する。
// isSymlink は userprompt_track_command.go (notification_handler.go) で定義済み。
func hasFixSymlinkComponent(path, root string) bool {
	path = strings.TrimSuffix(path, "/")
	root = strings.TrimSuffix(root, "/")

	for path != "" && path != root {
		if isSymlink(path) {
			return true
		}
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return isSymlink(root)
}

// writeFixProposalJSON は v を JSON として w に書き出す。
func writeFixProposalJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
