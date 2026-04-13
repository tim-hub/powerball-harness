package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//
type UsageTrackerHandler struct {
	ProjectRoot string
}

type usageTrackerInput struct {
	ToolName  string          `json:"tool_name"`
	ToolInput json.RawMessage `json:"tool_input"`
	CWD       string          `json:"cwd"`
}

type skillToolInput struct {
	Skill string `json:"skill"`
}

type slashCommandInput struct {
	Command string `json:"command"`
	Name    string `json:"name"`
}

type taskToolInput struct {
	SubagentType string `json:"subagent_type"`
}

type usageEntry struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Digest    string `json:"digest,omitempty"`
	Timestamp string `json:"timestamp"`
}

type usageTrackerResponse struct {
	Continue bool `json:"continue"`
}

const (
	usageStatsFile    = "usage-stats.jsonl"
	usageMaxSizeBytes = 100 * 1024 // 100KB
)

func (h *UsageTrackerHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	if len(data) > 0 {
		var inp usageTrackerInput
		if err := json.Unmarshal(data, &inp); err == nil && inp.ToolName != "" {
			projectRoot := h.resolveProjectRoot(inp.CWD)
			h.track(inp, projectRoot)
		}
	}

	return writeUsageJSON(w, usageTrackerResponse{Continue: true})
}

func (h *UsageTrackerHandler) resolveProjectRoot(cwd string) string {
	if cwd != "" {
		if root, err := gitRepoRoot(cwd); err == nil {
			return root
		}
		return cwd
	}
	if h.ProjectRoot != "" {
		return h.ProjectRoot
	}
	wd, _ := os.Getwd()
	return wd
}

func gitRepoRoot(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (h *UsageTrackerHandler) track(inp usageTrackerInput, projectRoot string) {
	var entry *usageEntry

	switch inp.ToolName {
	case "Skill":
		entry = h.trackSkill(inp, projectRoot)
	case "SlashCommand":
		entry = h.trackSlashCommand(inp, projectRoot)
	case "Task":
		entry = h.trackTask(inp)
	}

	if entry == nil {
		return
	}

	stateDir := filepath.Join(projectRoot, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return
	}
	statsFile := filepath.Join(stateDir, usageStatsFile)
	h.appendEntry(statsFile, entry)
}

func (h *UsageTrackerHandler) trackSkill(inp usageTrackerInput, projectRoot string) *usageEntry {
	var toolIn skillToolInput
	if err := json.Unmarshal(inp.ToolInput, &toolIn); err != nil || toolIn.Skill == "" {
		return nil
	}

	// "claude-code-harness:impl" → "impl"
	baseName := extractBaseName(toolIn.Skill, ":")

	if baseName == "sync-ssot-from-memory" || baseName == "memory" ||
		strings.Contains(toolIn.Skill, "sync-ssot-from-memory") ||
		strings.Contains(toolIn.Skill, ":memory") {
		h.touchSSOTFlag(projectRoot)
	}

	return &usageEntry{
		Type:      "skill",
		Name:      baseName,
		Digest:    digest(inp.ToolInput),
		Timestamp: nowISO(),
	}
}

func (h *UsageTrackerHandler) trackSlashCommand(inp usageTrackerInput, projectRoot string) *usageEntry {
	var toolIn slashCommandInput
	if err := json.Unmarshal(inp.ToolInput, &toolIn); err != nil {
		return nil
	}

	cmdName := toolIn.Command
	if cmdName == "" {
		cmdName = toolIn.Name
	}
	if cmdName == "" {
		return nil
	}

	baseName := strings.TrimPrefix(cmdName, "/")

	if strings.Contains(baseName, "sync-ssot-from-memory") || baseName == "memory" {
		h.touchSSOTFlag(projectRoot)
	}

	return &usageEntry{
		Type:      "command",
		Name:      baseName,
		Digest:    digest(inp.ToolInput),
		Timestamp: nowISO(),
	}
}

func (h *UsageTrackerHandler) trackTask(inp usageTrackerInput) *usageEntry {
	var toolIn taskToolInput
	if err := json.Unmarshal(inp.ToolInput, &toolIn); err != nil || toolIn.SubagentType == "" {
		return nil
	}

	return &usageEntry{
		Type:      "agent",
		Name:      toolIn.SubagentType,
		Digest:    digest(inp.ToolInput),
		Timestamp: nowISO(),
	}
}

func (h *UsageTrackerHandler) touchSSOTFlag(projectRoot string) {
	stateDir := filepath.Join(projectRoot, ".claude", "state")
	_ = os.MkdirAll(stateDir, 0700)
	flag := filepath.Join(stateDir, ".ssot-synced-this-session")
	_ = os.WriteFile(flag, []byte(""), 0600)
}

func (h *UsageTrackerHandler) appendEntry(statsFile string, entry *usageEntry) {
	if fi, err := os.Stat(statsFile); err == nil && fi.Size() > usageMaxSizeBytes {
		bakFile := statsFile + ".bak"
		_ = os.Rename(statsFile, bakFile)
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(statsFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s\n", line)
}

func extractBaseName(s, sep string) string {
	parts := strings.Split(s, sep)
	return parts[len(parts)-1]
}

func digest(raw json.RawMessage) string {
	s := string(raw)
	if len(s) > 100 {
		return s[:100]
	}
	return s
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func writeUsageJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
