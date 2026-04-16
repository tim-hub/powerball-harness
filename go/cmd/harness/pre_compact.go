package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type preCompactInput struct {
	SessionID string `json:"session_id,omitempty"`
	CWD       string `json:"cwd,omitempty"`
	AgentType string `json:"agent_type,omitempty"`
	Role      string `json:"role,omitempty"`
}

type preCompactDecision struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

func runPreCompact(_ []string) {
	exitCode, err := evaluatePreCompact(os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pre-compact error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func evaluatePreCompact(r io.Reader, w io.Writer) (int, error) {
	input, err := readPreCompactInput(r)
	if err != nil {
		return 0, nil
	}

	projectRoot := resolvePreCompactRoot(input.CWD)
	role := normalizePreCompactRole(input.Role, input.AgentType)
	sessionID := strings.TrimSpace(firstNonEmpty(input.SessionID, os.Getenv("CLAUDE_SESSION_ID")))

	if role == "reviewer" || role == "advisor" {
		return 0, nil
	}

	if shouldBlockLongRunningSession(projectRoot, sessionID) {
		return writePreCompactBlock(w, "long-running worker session is active; compact would interrupt owned work"), nil
	}

	plansPath := filepath.Join(projectRoot, "Plans.md")
	if isPlansDirty(projectRoot, plansPath) {
		return writePreCompactBlock(w, "Plans.md has uncommitted edits; save or checkpoint before compacting"), nil
	}

	return 0, nil
}

func readPreCompactInput(r io.Reader) (preCompactInput, error) {
	var input preCompactInput
	data, err := io.ReadAll(r)
	if err != nil {
		return input, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return input, nil
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return preCompactInput{}, err
	}
	return input, nil
}

func resolvePreCompactRoot(cwd string) string {
	if strings.TrimSpace(cwd) != "" {
		return cwd
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

func normalizePreCompactRole(values ...string) string {
	candidates := append([]string{}, values...)
	candidates = append(candidates,
		os.Getenv("HARNESS_SESSION_ROLE"),
		os.Getenv("HARNESS_ACTIVE_ROLE"),
		os.Getenv("CLAUDE_AGENT_ROLE"),
		os.Getenv("CLAUDE_AGENT_TYPE"),
	)

	for _, raw := range candidates {
		value := strings.ToLower(strings.TrimSpace(raw))
		switch {
		case strings.Contains(value, "reviewer"):
			return "reviewer"
		case strings.Contains(value, "advisor"):
			return "advisor"
		case strings.Contains(value, "worker"):
			return "worker"
		}
	}
	return ""
}

func shouldBlockLongRunningSession(projectRoot, sessionID string) bool {
	if sessionID == "" {
		return false
	}

	metaPath := filepath.Join(projectRoot, ".claude", "state", "locks", "loop-session.lock.d", "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return false
	}

	var meta struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return false
	}

	return strings.TrimSpace(meta.SessionID) == sessionID
}

func isPlansDirty(projectRoot, plansPath string) bool {
	if _, err := os.Stat(plansPath); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(projectRoot, ".git")); err != nil {
		return false
	}

	relPath, err := filepath.Rel(projectRoot, plansPath)
	if err != nil {
		relPath = "Plans.md"
	}

	cmd := exec.Command("git", "status", "--short", "--", relPath)
	cmd.Dir = projectRoot
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(out)) != ""
}

func writePreCompactBlock(w io.Writer, reason string) int {
	resp := preCompactDecision{
		Decision: "block",
		Reason:   reason,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(w, "{\"decision\":\"block\"}\n")
		return 2
	}
	fmt.Fprintf(w, "%s\n", data)
	return 2
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
