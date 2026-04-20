package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// daemonProbe is a package-level variable for test injection.
// Default: real TCP connect. Tests inject a stub.
var daemonProbe = func(addr string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// resolveHarnessBinary resolves the harness binary path using a safe lookup
// strategy that never trusts a path derived from the project root (which could
// allow a malicious committed binary to bypass guardrails).
//
// Resolution order:
//  1. os.Executable() → sibling "harness" binary in the same directory
//  2. CLAUDE_PLUGIN_ROOT env var → bin/harness relative to plugin root
//  3. PATH fallback via exec.LookPath
func resolveHarnessBinary() string {
	// 1. os.Executable() → look for harness relative to executable dir
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "harness")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	// 2. CLAUDE_PLUGIN_ROOT env var
	if root := os.Getenv("CLAUDE_PLUGIN_ROOT"); root != "" {
		candidate := filepath.Join(root, "bin", "harness")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	// 3. PATH fallback
	if p, err := exec.LookPath("harness"); err == nil {
		return p
	}
	return ""
}

// memHealthResult is the JSON output of the mem health subcommand.
type memHealthResult struct {
	Healthy bool   `json:"healthy"`
	Reason  string `json:"reason,omitempty"`
}

// runMem dispatches harness mem subcommands.
func runMem(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: harness mem <health>")
		os.Exit(1)
	}
	switch args[0] {
	case "health":
		runMemHealthCheck()
	default:
		fmt.Fprintf(os.Stderr, "Unknown mem subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

// runMemHealthCheck runs the health check and prints JSON to stdout.
// On unhealthy result it also writes a human-readable warning to stderr
// and exits with code 1.
func runMemHealthCheck() {
	result := checkMemHealth()
	data, _ := json.Marshal(result)
	fmt.Println(string(data))
	if !result.Healthy {
		fmt.Fprintf(os.Stderr, "harness-mem unhealthy: %s\n", result.Reason)
		os.Exit(1)
	}
}

// checkMemHealth performs a two-stage health check:
//  1. File integrity: ~/.claude-mem directory + settings.json OR supervisor.json validity
//  2. Daemon TCP probe: connect to HARNESS_MEM_HOST:HARNESS_MEM_PORT
//
// Absent ~/.claude-mem/ means harness-mem is not installed (opt-in), not a failure.
func checkMemHealth() memHealthResult {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return memHealthResult{Healthy: true, Reason: "not-configured"}
	}

	memDir := filepath.Join(homeDir, ".claude-mem")

	// Stage 1a: missing directory means harness-mem is not installed — healthy opt-out.
	if _, err := os.Stat(memDir); os.IsNotExist(err) {
		return memHealthResult{Healthy: true, Reason: "not-configured"}
	}

	// Stage 1b: accept settings.json OR supervisor.json (either valid JSON suffices).
	validJSON := func(path string) bool {
		data, err := os.ReadFile(path)
		return err == nil && json.Valid(data)
	}
	settingsOK := validJSON(filepath.Join(memDir, "settings.json"))
	supervisorOK := validJSON(filepath.Join(memDir, "supervisor.json"))
	if !settingsOK && !supervisorOK {
		return memHealthResult{Healthy: false, Reason: "corrupted"}
	}

	// Stage 2: daemon TCP probe
	host := os.Getenv("HARNESS_MEM_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("HARNESS_MEM_PORT")
	if port == "" {
		port = "37888"
	}
	addr := host + ":" + port

	if err := daemonProbe(addr, 500*time.Millisecond); err != nil {
		return memHealthResult{Healthy: false, Reason: "daemon-unreachable"}
	}

	return memHealthResult{Healthy: true}
}
