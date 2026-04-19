package session

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// MonitorHandler is the SessionStart hook handler (project state collection).
// Collects project state at session start and generates session.json and tooling-policy.json.
//
// Shell equivalent: scripts/session-monitor.sh
type MonitorHandler struct {
	// StateDir is the state directory path. If empty, it is inferred from cwd.
	StateDir string
	// PlansFile is the path to Plans.md. If empty, projectRoot/Plans.md is used.
	PlansFile string
	// now is a time injection function for testing. If nil, time.Now() is used.
	now func() time.Time
}

// monitorInput is the stdin JSON for the SessionStart hook.
type monitorInput struct {
	SessionID string `json:"session_id,omitempty"`
	CWD       string `json:"cwd,omitempty"`
}

// sessionStateJSON is the complete schema for session.json.
type sessionStateJSON struct {
	SessionID    string            `json:"session_id"`
	ParentID     interface{}       `json:"parent_session_id"`
	State        string            `json:"state"`
	StateVersion int               `json:"state_version"`
	StartedAt    string            `json:"started_at"`
	UpdatedAt    string            `json:"updated_at"`
	ResumeToken  string            `json:"resume_token"`
	EventSeq     int               `json:"event_seq"`
	LastEventID  string            `json:"last_event_id"`
	ForkCount    int               `json:"fork_count"`
	Orchestration orchestrationJSON `json:"orchestration"`
	CWD          string            `json:"cwd"`
	ProjectName  string            `json:"project_name"`
	PromptSeq    int               `json:"prompt_seq"`
	Git          gitStateJSON      `json:"git"`
	Plans        plansStateJSON    `json:"plans"`
	ChangesThisSession []interface{} `json:"changes_this_session"`
}

type orchestrationJSON struct {
	MaxStateRetries      int `json:"max_state_retries"`
	RetryBackoffSeconds  int `json:"retry_backoff_seconds"`
}

type gitStateJSON struct {
	Branch             string `json:"branch"`
	UncommittedChanges int    `json:"uncommitted_changes"`
	LastCommit         string `json:"last_commit"`
}

type plansStateJSON struct {
	Exists         bool   `json:"exists"`
	LastModified   int64  `json:"last_modified"`
	WIPTasks       int    `json:"wip_tasks"`
	TODOTasks      int    `json:"todo_tasks"`
	PendingTasks   int    `json:"pending_tasks"`
	CompletedTasks int    `json:"completed_tasks"`
}

// toolingPolicyJSON is the schema for tooling-policy.json (simplified).
// Only generates basic information to avoid heavy external command dependencies for LSP/MCP detection.
type toolingPolicyJSON struct {
	LSP     lspPolicyJSON     `json:"lsp"`
	Plugins pluginPolicyJSON  `json:"plugins"`
	MCP     mcpPolicyJSON     `json:"mcp"`
	Skills  skillsPolicyJSON  `json:"skills"`
}

type lspPolicyJSON struct {
	Available         bool              `json:"available"`
	Plugins           string            `json:"plugins"`
	AvailableByExt    map[string]bool   `json:"available_by_ext"`
	LastUsedPromptSeq int               `json:"last_used_prompt_seq"`
	LastUsedToolName  string            `json:"last_used_tool_name"`
	UsedSinceLastPrompt bool            `json:"used_since_last_prompt"`
}

type pluginPolicyJSON struct {
	Installed       *int    `json:"installed"`
	EnabledEstimate *int    `json:"enabled_estimate"`
	Source          string  `json:"source"`
}

type mcpPolicyJSON struct {
	Configured      *int     `json:"configured"`
	Disabled        *int     `json:"disabled"`
	EnabledEstimate *int     `json:"enabled_estimate"`
	Sources         []string `json:"sources"`
}

type skillsPolicyJSON struct {
	Index           []interface{} `json:"index"`
	DecisionRequired bool         `json:"decision_required"`
}

// Handle reads the SessionStart payload from stdin, generates session.json and
// tooling-policy.json, and writes a state summary to stdout.
func (h *MonitorHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	var inp monitorInput
	if len(data) > 0 {
		_ = json.Unmarshal(data, &inp)
	}

	projectRoot := resolveProjectRoot(inp.CWD)
	if projectRoot == "" {
		cwd, _ := os.Getwd()
		projectRoot = cwd
	}

	stateDir := h.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(projectRoot, ".claude", "state")
	}

	// Symlink check
	if isSymlink(stateDir) || isSymlink(filepath.Dir(stateDir)) {
		fmt.Fprintf(os.Stderr, "[session-monitor] Warning: symlink detected in state directory path, aborting\n")
		return nil
	}

	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return nil
	}

	plansFile := h.PlansFile
	if plansFile == "" {
		plansFile = filepath.Join(projectRoot, "Plans.md")
	}

	now := h.currentTime()
	nowStr := now.UTC().Format(time.RFC3339)

	// Collect project information
	projectName := filepath.Base(projectRoot)
	gitState := h.collectGitState(projectRoot)
	plansState := h.collectPlansState(plansFile)

	// Generate session.json (determine resume vs new session)
	sessionFile := filepath.Join(stateDir, "session.json")
	h.generateSessionFile(sessionFile, projectRoot, projectName, nowStr, gitState, plansState)

	// Generate tooling-policy.json
	policyFile := filepath.Join(stateDir, "tooling-policy.json")
	h.generateToolingPolicy(policyFile)

	// Write summary to stdout
	h.writeSummary(w, projectName, gitState, plansState)

	// Check for Plans.md drift and emit warning if thresholds exceeded
	h.CheckPlansDrift(w, plansState, plansFile, projectRoot)

	return nil
}

// collectGitState collects git information.
func (h *MonitorHandler) collectGitState(projectRoot string) gitStateJSON {
	gitDir := filepath.Join(projectRoot, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return gitStateJSON{
			Branch:             "(no git)",
			UncommittedChanges: 0,
			LastCommit:         "none",
		}
	}

	// Read branch name from HEAD
	branch := h.readGitBranch(projectRoot)
	return gitStateJSON{
		Branch:             branch,
		UncommittedChanges: 0, // Fixed at 0 to avoid expensive operations
		LastCommit:         h.readGitLastCommit(projectRoot),
	}
}

// readGitBranch reads the branch name from .git/HEAD.
func (h *MonitorHandler) readGitBranch(projectRoot string) string {
	headFile := filepath.Join(projectRoot, ".git", "HEAD")
	data, err := os.ReadFile(headFile)
	if err != nil {
		return "unknown"
	}
	line := strings.TrimSpace(string(data))
	// "ref: refs/heads/<branch>"
	if strings.HasPrefix(line, "ref: refs/heads/") {
		return strings.TrimPrefix(line, "ref: refs/heads/")
	}
	// Detached HEAD: SHA only
	if len(line) >= 7 {
		return line[:7]
	}
	return "unknown"
}

// readGitLastCommit reads the latest commit SHA from .git/refs/heads/<branch>.
func (h *MonitorHandler) readGitLastCommit(projectRoot string) string {
	headFile := filepath.Join(projectRoot, ".git", "HEAD")
	data, err := os.ReadFile(headFile)
	if err != nil {
		return "none"
	}
	line := strings.TrimSpace(string(data))
	if strings.HasPrefix(line, "ref: ") {
		ref := strings.TrimPrefix(line, "ref: ")
		refFile := filepath.Join(projectRoot, ".git", filepath.FromSlash(ref))
		refData, err := os.ReadFile(refFile)
		if err != nil {
			// Fall back to checking packed-refs
			return h.readPackedRef(projectRoot, ref)
		}
		sha := strings.TrimSpace(string(refData))
		if len(sha) >= 7 {
			return sha[:7]
		}
		return sha
	}
	// Detached HEAD
	if len(line) >= 7 {
		return line[:7]
	}
	return "none"
}

// readPackedRef searches .git/packed-refs for a ref.
func (h *MonitorHandler) readPackedRef(projectRoot, ref string) string {
	packedFile := filepath.Join(projectRoot, ".git", "packed-refs")
	data, err := os.ReadFile(packedFile)
	if err != nil {
		return "none"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasSuffix(line, " "+ref) || strings.HasSuffix(line, "\t"+ref) {
			parts := strings.Fields(line)
			if len(parts) >= 1 && len(parts[0]) >= 7 {
				return parts[0][:7]
			}
		}
	}
	return "none"
}

// collectPlansState collects the state of Plans.md.
func (h *MonitorHandler) collectPlansState(plansFile string) plansStateJSON {
	fi, err := os.Stat(plansFile)
	if err != nil {
		return plansStateJSON{Exists: false}
	}

	wipCount := countMatches(plansFile, "cc:WIP")
	todoCount := countMatches(plansFile, "cc:TODO")
	pendingCount := countMatches(plansFile, "pm:pending", "cursor:pending")
	completedCount := countMatches(plansFile, "cc:done")

	return plansStateJSON{
		Exists:         true,
		LastModified:   fi.ModTime().Unix(),
		WIPTasks:       wipCount,
		TODOTasks:      todoCount,
		PendingTasks:   pendingCount,
		CompletedTasks: completedCount,
	}
}

// generateSessionFile generates session.json (determining resume vs new session).
func (h *MonitorHandler) generateSessionFile(
	sessionFile, projectRoot, projectName, nowStr string,
	git gitStateJSON,
	plans plansStateJSON,
) {
	if isSymlink(sessionFile) {
		return
	}

	resumeMode := false
	var existing sessionStateJSON

	// Load existing session
	if data, err := os.ReadFile(sessionFile); err == nil {
		if json.Unmarshal(data, &existing) == nil {
			// Resume mode when ended_at is not set
			// Here: resume when EventSeq > 0 and State is an active variant
			if existing.SessionID != "" && existing.State != "stopped" && existing.State != "completed" && existing.State != "failed" {
				resumeMode = true
			}
		}
	}

	forkMode := os.Getenv("HARNESS_SESSION_FORK") == "true"
	if forkMode {
		resumeMode = false
	}

	var sess sessionStateJSON

	if resumeMode {
		// Update existing session
		existing.CWD = projectRoot
		existing.ProjectName = projectName
		existing.UpdatedAt = nowStr
		existing.Git = git
		existing.Plans = plans
		existing.StateVersion = 1
		sess = existing
	} else {
		// New session
		sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
		resumeToken := fmt.Sprintf("resume-%d", time.Now().UnixNano())
		parentID := interface{}(nil)
		if forkMode && existing.SessionID != "" {
			parentID = existing.SessionID
		}

		sess = sessionStateJSON{
			SessionID:    sessionID,
			ParentID:     parentID,
			State:        "initialized",
			StateVersion: 1,
			StartedAt:    nowStr,
			UpdatedAt:    nowStr,
			ResumeToken:  resumeToken,
			EventSeq:     0,
			LastEventID:  "",
			ForkCount:    0,
			Orchestration: orchestrationJSON{
				MaxStateRetries:     3,
				RetryBackoffSeconds: 10,
			},
			CWD:          projectRoot,
			ProjectName:  projectName,
			PromptSeq:    0,
			Git:          git,
			Plans:        plans,
			ChangesThisSession: []interface{}{},
		}
	}

	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return
	}

	_ = writeFileAtomic(sessionFile, append(data, '\n'), 0600)
}

// generateToolingPolicy generates tooling-policy.json.
// Avoids heavy external command dependencies (claude plugin list, MCP server search, etc.)
// and generates only a basic scaffold.
func (h *MonitorHandler) generateToolingPolicy(policyFile string) {
	if isSymlink(policyFile) {
		return
	}

	policy := toolingPolicyJSON{
		LSP: lspPolicyJSON{
			Available:         false,
			Plugins:           "",
			AvailableByExt:    map[string]bool{},
			LastUsedPromptSeq: 0,
			LastUsedToolName:  "",
			UsedSinceLastPrompt: false,
		},
		Plugins: pluginPolicyJSON{
			Installed:       nil,
			EnabledEstimate: nil,
			Source:          "",
		},
		MCP: mcpPolicyJSON{
			Configured:      nil,
			Disabled:        nil,
			EnabledEstimate: nil,
			Sources:         []string{},
		},
		Skills: skillsPolicyJSON{
			Index:            []interface{}{},
			DecisionRequired: false,
		},
	}

	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return
	}

	_ = writeFileAtomic(policyFile, append(data, '\n'), 0644)
}

// plansDriftConfig holds the thresholds for Plans.md drift detection.
type plansDriftConfig struct {
	WIPThreshold   int // emit warning when WIP task count >= this value
	StaleHours     int // emit warning when Plans.md has not been modified for >= this many hours
}

// defaultPlansDriftConfig returns the default drift detection thresholds.
func defaultPlansDriftConfig() plansDriftConfig {
	return plansDriftConfig{WIPThreshold: 5, StaleHours: 24}
}

// loadPlansDriftConfig reads the monitor.plans_drift section from the config file.
// If the file or section is absent the defaults are returned.
// projectRoot is used to locate harness/.claude-code-harness.config.yaml.
func loadPlansDriftConfig(projectRoot string) plansDriftConfig {
	cfg := defaultPlansDriftConfig()

	configPath := filepath.Clean(filepath.Join(projectRoot, "harness", ".claude-code-harness.config.yaml"))
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg
	}

	// Lightweight key-value scan: only care about wip_threshold and stale_hours
	// inside the monitor.plans_drift block.
	inMonitor := false
	inPlansDrift := false
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimRight(raw, "\r")
		trimmed := strings.TrimSpace(line)

		// Track indentation level by leading spaces
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Top-level "monitor:" key
		if indent == 0 {
			inMonitor = strings.TrimRight(trimmed, ":") == "monitor" && strings.HasSuffix(trimmed, ":")
			inPlansDrift = false
			continue
		}

		if !inMonitor {
			continue
		}

		// Second-level "plans_drift:" key
		if indent > 0 && !inPlansDrift {
			inPlansDrift = strings.TrimRight(trimmed, ":") == "plans_drift" && strings.HasSuffix(trimmed, ":")
			continue
		}

		if !inPlansDrift {
			continue
		}

		// Third-level key: value pairs
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "wip_threshold":
			if v := atoi(val); v > 0 {
				cfg.WIPThreshold = v
			}
		case "stale_hours":
			if v := atoi(val); v > 0 {
				cfg.StaleHours = v
			}
		}
	}

	return cfg
}

// CheckPlansDrift emits a warning line to w when Plans.md has too many WIP tasks
// or has not been modified recently enough.
// projectRoot is used to locate the config file.
func (h *MonitorHandler) CheckPlansDrift(w io.Writer, plans plansStateJSON, plansFile string, projectRoot string) {
	if !plans.Exists {
		return
	}

	cfg := loadPlansDriftConfig(projectRoot)

	wipCount := plans.WIPTasks
	now := h.currentTime()
	staleFor := int(now.Unix()-plans.LastModified) / 3600

	if wipCount >= cfg.WIPThreshold || staleFor >= cfg.StaleHours {
		fmt.Fprintf(w, "⚠️ plans drift: WIP=%d, stale_for=%dh\n", wipCount, staleFor)
	}
}

// writeSummary writes the session state summary to w.
func (h *MonitorHandler) writeSummary(w io.Writer, projectName string, git gitStateJSON, plans plansStateJSON) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Session started - Project state")
	fmt.Fprintln(w, strings.Repeat("─", 36))
	fmt.Fprintf(w, "Project: %s\n", projectName)
	fmt.Fprintf(w, "Branch: %s\n", git.Branch)

	if plans.Exists {
		total := plans.WIPTasks + plans.TODOTasks + plans.PendingTasks
		if total > 0 {
			fmt.Fprintf(w, "Plans.md: WIP %d / TODO %d\n", plans.WIPTasks, plans.TODOTasks+plans.PendingTasks)
		}
	}

	fmt.Fprintln(w, strings.Repeat("─", 36))
	fmt.Fprintln(w, "")
}

// currentTime returns the current time.
func (h *MonitorHandler) currentTime() time.Time {
	if h.now != nil {
		return h.now()
	}
	return time.Now()
}

// formatInt returns an int as a pointer (used for null values in tooling-policy).
func formatInt(v int) *int {
	return &v
}

// atoi converts a string to int.
func atoi(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}
