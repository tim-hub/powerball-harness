package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// runCodexLoop delegates to harness/scripts/codex-loop.sh so the public CLI surface is
// stable while the loop runtime can keep reusing the existing shell scripts.
func runCodexLoop(args []string) {
	projectRoot, err := resolveProjectRoot(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "harness codex-loop: %v\n", err)
		os.Exit(1)
	}

	scriptPath := filepath.Join(projectRoot, "harness", "scripts", "codex-loop.sh")
	if _, err := os.Stat(scriptPath); err != nil {
		fmt.Fprintf(os.Stderr, "harness codex-loop: script not found: %s\n", scriptPath)
		os.Exit(1)
	}

	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "PROJECT_ROOT="+projectRoot)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "harness codex-loop: %v\n", err)
		os.Exit(1)
	}
}
