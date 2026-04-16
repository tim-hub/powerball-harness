package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveHarnessInstallRootFromEnv(t *testing.T) {
	t.Setenv("HARNESS_INSTALL_ROOT", "/tmp/custom-harness-root")
	t.Setenv("CLAUDE_PLUGIN_ROOT", "/tmp/ignored-plugin-root")

	root, err := resolveHarnessInstallRoot("/tmp/project-root")
	if err != nil {
		t.Fatalf("resolveHarnessInstallRoot returned error: %v", err)
	}

	if got, want := root, "/tmp/custom-harness-root"; got != want {
		t.Fatalf("resolveHarnessInstallRoot = %q, want %q", got, want)
	}
}

func TestResolveHarnessInstallRootFromExecutableLayout(t *testing.T) {
	t.Setenv("HARNESS_INSTALL_ROOT", "")
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")

	tmpDir := t.TempDir()
	installRoot := filepath.Join(tmpDir, "install-root")
	scriptsDir := filepath.Join(installRoot, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll scripts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "codex-loop.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("WriteFile codex-loop.sh: %v", err)
	}

	binaryPath := filepath.Join(installRoot, "bin", "harness-darwin-arm64")
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
		t.Fatalf("MkdirAll bin: %v", err)
	}
	if err := os.WriteFile(binaryPath, []byte("binary"), 0o755); err != nil {
		t.Fatalf("WriteFile binary: %v", err)
	}

	realExecutablePath := osExecutablePath
	defer func() { osExecutablePath = realExecutablePath }()
	osExecutablePath = func() (string, error) {
		return binaryPath, nil
	}

	root, err := resolveHarnessInstallRoot("/tmp/project-root")
	if err != nil {
		t.Fatalf("resolveHarnessInstallRoot returned error: %v", err)
	}

	resolvedInstallRoot, err := filepath.EvalSymlinks(installRoot)
	if err != nil {
		t.Fatalf("EvalSymlinks installRoot: %v", err)
	}

	if got, want := root, resolvedInstallRoot; got != want {
		t.Fatalf("resolveHarnessInstallRoot = %q, want %q", got, want)
	}
}

func TestResolveHarnessInstallRootFallsBackToProjectRoot(t *testing.T) {
	t.Setenv("HARNESS_INSTALL_ROOT", "")
	t.Setenv("CLAUDE_PLUGIN_ROOT", "")

	projectRoot := t.TempDir()
	scriptsDir := filepath.Join(projectRoot, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll scripts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(scriptsDir, "codex-loop.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("WriteFile codex-loop.sh: %v", err)
	}

	realExecutablePath := osExecutablePath
	defer func() { osExecutablePath = realExecutablePath }()
	osExecutablePath = func() (string, error) {
		return "", os.ErrNotExist
	}

	root, err := resolveHarnessInstallRoot(projectRoot)
	if err != nil {
		t.Fatalf("resolveHarnessInstallRoot returned error: %v", err)
	}

	if got, want := root, projectRoot; got != want {
		t.Fatalf("resolveHarnessInstallRoot = %q, want %q", got, want)
	}
}
