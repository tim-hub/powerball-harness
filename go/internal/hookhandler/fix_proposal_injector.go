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

// FixProposalInjectorHandler is the UserPromptSubmit hook handler.
// It reads pending-fix-proposals.jsonl and notifies the user of unseen proposals.
// It also interprets "approve fix" / "reject fix" commands and applies them to Plans.md.
//
// shell counterpart: scripts/hook-handlers/fix-proposal-injector.sh
type FixProposalInjectorHandler struct {
	// ProjectRoot is the project root path. Falls back to cwd when empty.
	ProjectRoot string
	// PlansPath is the path to Plans.md. Falls back to ProjectRoot/Plans.md when empty.
	PlansPath string
}

// fixProposalInjectorInput is the stdin JSON for the UserPromptSubmit hook.
type fixProposalInjectorInput struct {
	Prompt string `json:"prompt"`
}

// fixProposalInjectorOutput is the response for the UserPromptSubmit hook.
// Notifies the user via systemMessage.
type fixProposalInjectorOutput struct {
	SystemMessage string `json:"systemMessage,omitempty"`
}

const (
	pendingFixProposalsFile = "pending-fix-proposals.jsonl"
	fixProposalMaxLines     = 500
)

// Handle reads the UserPromptSubmit payload from stdin and
// notifies/processes fix proposals.
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

	// Skip if the proposals file does not exist.
	if _, err := os.Stat(proposalsFile); os.IsNotExist(err) {
		return nil
	}

	// Symlink check (isSymlink is defined in notification_handler.go).
	if hasFixSymlinkComponent(stateDir, projectRoot) || isSymlink(proposalsFile) {
		return writeFixProposalJSON(w, fixProposalInjectorOutput{
			SystemMessage: "⚠️ fix proposal state path is a symlink — aborting.",
		})
	}

	plansPath := h.resolvePlansPath(projectRoot)
	if _, err := os.Stat(plansPath); err == nil {
		if hasFixSymlinkComponent(plansPath, projectRoot) {
			return writeFixProposalJSON(w, fixProposalInjectorOutput{
				SystemMessage: "⚠️ Plans.md path is a symlink — cannot apply fix proposal.",
			})
		}
	}

	// Parse prompt to determine action.
	firstLine := strings.TrimSpace(strings.SplitN(inp.Prompt, "\n", 2)[0])
	lower := strings.ToLower(firstLine)
	action, targetID := parseFixProposalAction(lower, firstLine)

	// Load pending proposals.
	proposals, err := loadPendingFixProposals(proposalsFile)
	if err != nil || len(proposals) == 0 {
		return nil
	}

	pendingCount := len(proposals)

	// Error if an action is requested but no target ID is given and there are multiple proposals.
	if action != "" && targetID == "" && pendingCount != 1 {
		return writeFixProposalJSON(w, fixProposalInjectorOutput{
			SystemMessage: fmt.Sprintf(
				"⚠️ There are %d unprocessed fix proposals. Use 'approve fix <task_id>' or 'reject fix <task_id>' to specify the target.",
				pendingCount,
			),
		})
	}

	// Select target proposal.
	proposal, found := selectFixProposal(proposals, targetID)
	if !found {
		if targetID != "" {
			return writeFixProposalJSON(w, fixProposalInjectorOutput{
				SystemMessage: fmt.Sprintf("⚠️ Specified fix proposal not found: %s", targetID),
			})
		}
		return nil
	}

	// Approve processing.
	if action == "approve" {
		applyResult := applyFixProposalToPlans(plansPath, proposal)
		if applyResult == "applied" || applyResult == "already_present" {
			if err := consumeFixProposal(proposalsFile, proposal.SourceTaskID); err != nil {
				_ = err
			}
			return writeFixProposalJSON(w, fixProposalInjectorOutput{
				SystemMessage: fmt.Sprintf("✅ Fix proposal applied: %s\nContent: %s", proposal.FixTaskID, proposal.ProposalSubject),
			})
		} else if applyResult == "plans_missing" {
			return writeFixProposalJSON(w, fixProposalInjectorOutput{
				SystemMessage: "⚠️ Could not apply fix proposal. Plans.md not found.",
			})
		} else {
			return writeFixProposalJSON(w, fixProposalInjectorOutput{
				SystemMessage: fmt.Sprintf("⚠️ Failed to apply fix proposal. Source task %s not found in Plans.md.", proposal.SourceTaskID),
			})
		}
	}

	// Reject processing.
	if action == "reject" {
		_ = consumeFixProposal(proposalsFile, proposal.SourceTaskID)
		return writeFixProposalJSON(w, fixProposalInjectorOutput{
			SystemMessage: fmt.Sprintf("ℹ️ Fix proposal rejected: %s", proposal.FixTaskID),
		})
	}

	// No action → show reminder.
	reminder := buildFixProposalReminder(proposal, pendingCount)
	return writeFixProposalJSON(w, fixProposalInjectorOutput{SystemMessage: reminder})
}

// resolveProjectRoot resolves the project root path.
func (h *FixProposalInjectorHandler) resolveProjectRoot() string {
	if h.ProjectRoot != "" {
		return h.ProjectRoot
	}
	wd, _ := os.Getwd()
	return wd
}

// resolvePlansPath resolves the path to Plans.md.
// Uses PlansPath when explicitly set; otherwise resolves via the config's plansDirectory.
// Always returns a full path even if Plans.md does not exist
// (so that apply can return plans_missing).
func (h *FixProposalInjectorHandler) resolvePlansPath(projectRoot string) string {
	if h.PlansPath != "" {
		return h.PlansPath
	}
	// Get the path of an existing Plans.md.
	if p := resolvePlansPath(projectRoot); p != "" {
		return p
	}
	// Fall back to a default path that respects the configured plansDirectory.
	plansDir := readPlansDirectoryFromConfig(projectRoot)
	if plansDir != "" {
		return filepath.Join(projectRoot, plansDir, "Plans.md")
	}
	return filepath.Join(projectRoot, "Plans.md")
}

// parseFixProposalAction parses the action and target ID from a prompt line.
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
	case lower == "yes" || lower == "approve":
		action = "approve"
	case lower == "no" || lower == "reject":
		action = "reject"
	}
	return action, targetID
}

// loadPendingFixProposals loads fixProposal entries with status=pending from a JSONL file.
// The fixProposal type is defined in task_completed_escalation.go.
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

// selectFixProposal returns the fixProposal matching selector from proposals.
// Returns the first proposal when selector is empty.
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

// consumeFixProposal removes the line with the given source_task_id from the JSONL file.
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
			continue // Skip the entry to be removed.
		}
		remaining = append(remaining, p)
	}
	f.Close()
	if err := scanner.Err(); err != nil {
		return err
	}

	// Rewrite the file (JSONL rotation: trim from the tail when over 500 lines).
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

// applyFixProposalToPlans inserts the proposal into Plans.md immediately after the source_task_id row.
// Returns: "applied" / "already_present" / "plans_missing" / "source_not_found"
func applyFixProposalToPlans(plansPath string, proposal fixProposal) string {
	rawData, err := os.ReadFile(plansPath)
	if err != nil {
		return "plans_missing"
	}

	text := string(rawData)

	// Check whether fix_task_id already exists.
	fixPattern := regexp.MustCompile(`(?m)^\|\s*` + regexp.QuoteMeta(proposal.FixTaskID) + `\s*\|`)
	if fixPattern.MatchString(text) {
		return "already_present"
	}

	// Find the source_task_id row and insert immediately after it.
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

// buildFixProposalReminder builds the reminder message.
func buildFixProposalReminder(proposal fixProposal, pendingCount int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[FIX PROPOSAL] There are %d unprocessed fix task proposals\n", pendingCount))
	sb.WriteString(fmt.Sprintf("Target: %s — %s\n", proposal.FixTaskID, proposal.ProposalSubject))
	sb.WriteString(fmt.Sprintf("Failure category: %s\n", proposal.FailureCategory))
	sb.WriteString(fmt.Sprintf("DoD: %s\n", proposal.DoD))
	if proposal.RecommendedAction != "" {
		sb.WriteString(fmt.Sprintf("Recommended action: %s\n", proposal.RecommendedAction))
	}
	sb.WriteString(fmt.Sprintf("Approve: approve fix %s\n", proposal.SourceTaskID))
	sb.WriteString(fmt.Sprintf("Reject: reject fix %s", proposal.SourceTaskID))
	return sb.String()
}

// hasFixSymlinkComponent reports whether path contains a symlink component within the project root.
// isSymlink is defined in userprompt_track_command.go (notification_handler.go).
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

// writeFixProposalJSON serializes v as JSON and writes it to w.
func writeFixProposalJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
