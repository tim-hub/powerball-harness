package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var osExecutablePath = os.Executable

// runCodexLoop delegates to scripts/codex-loop.sh so the public CLI surface is
// stable while the loop runtime can keep reusing the existing shell scripts.
func runCodexLoop(args []string) {
	projectRoot, err := resolveProjectRoot(nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "harness codex-loop: %v\n", err)
		os.Exit(1)
	}

	installRoot, err := resolveHarnessInstallRoot(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "harness codex-loop: %v\n", err)
		os.Exit(1)
	}

	scriptPath := filepath.Join(installRoot, "scripts", "codex-loop.sh")
	if _, err := os.Stat(scriptPath); err != nil {
		fmt.Fprintf(os.Stderr, "harness codex-loop: script not found: %s\n", scriptPath)
		os.Exit(1)
	}

	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Dir = projectRoot
	cmd.Env = append(
		os.Environ(),
		"PROJECT_ROOT="+projectRoot,
		"HARNESS_INSTALL_ROOT="+installRoot,
	)
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

func resolveHarnessInstallRoot(projectRoot string) (string, error) {
	if root := os.Getenv("HARNESS_INSTALL_ROOT"); root != "" {
		return filepath.Abs(root)
	}
	if root := os.Getenv("CLAUDE_PLUGIN_ROOT"); root != "" {
		return filepath.Abs(root)
	}

	executablePath, err := osExecutablePath()
	if err == nil {
		if resolvedPath, resolveErr := filepath.EvalSymlinks(executablePath); resolveErr == nil {
			executablePath = resolvedPath
		}

		binDir := filepath.Dir(executablePath)
		candidate := filepath.Dir(binDir)
		if candidate != "" {
			if _, statErr := os.Stat(filepath.Join(candidate, "scripts", "codex-loop.sh")); statErr == nil {
				return candidate, nil
			}
		}
	}

	if _, err := os.Stat(filepath.Join(projectRoot, "scripts", "codex-loop.sh")); err == nil {
		return projectRoot, nil
	}

	return "", fmt.Errorf("cannot resolve harness install root from executable or project root")
}
