package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunMemHealth_NotConfigured(t *testing.T) {
	// daemonProbe should not be called when ~/.claude-mem is absent
	orig := daemonProbe
	defer func() { daemonProbe = orig }()
	daemonProbe = func(addr string, d time.Duration) error {
		t.Error("daemonProbe should not be called when not configured")
		return nil
	}

	// Point HOME at a temp dir that has no .claude-mem subdirectory.
	// This guarantees the not-configured path is taken on any machine.
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	result := checkMemHealth()
	if !result.Healthy {
		t.Errorf("expected healthy (not-configured is opt-out, not an error), got unhealthy")
	}
	if result.Reason != "not-configured" {
		t.Errorf("expected reason=not-configured, got %q", result.Reason)
	}
}

func TestRunMemHealth_Corrupted(t *testing.T) {
	// daemonProbe should not be called when file integrity check fails
	orig := daemonProbe
	defer func() { daemonProbe = orig }()
	daemonProbe = func(addr string, d time.Duration) error {
		t.Error("daemonProbe should not be called with corrupted config files")
		return nil
	}

	dir := t.TempDir()
	memDir := filepath.Join(dir, ".claude-mem")
	if err := os.MkdirAll(memDir, 0700); err != nil {
		t.Fatal(err)
	}
	// Write invalid JSON to settings.json; supervisor.json is also absent.
	// Both settings.json (invalid) and supervisor.json (absent) fail → corrupted.
	if err := os.WriteFile(filepath.Join(memDir, "settings.json"), []byte("not-json{{{"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)

	result := checkMemHealth()
	if result.Healthy {
		t.Error("expected unhealthy for corrupted config files")
	}
	if result.Reason != "corrupted" {
		t.Errorf("expected reason=corrupted, got %q", result.Reason)
	}
}

func TestRunMemHealth_SupervisorJSONFallback(t *testing.T) {
	// When settings.json is absent but supervisor.json contains valid JSON,
	// the integrity check passes and the daemon probe is consulted.
	orig := daemonProbe
	defer func() { daemonProbe = orig }()
	daemonProbe = func(addr string, d time.Duration) error {
		return nil // simulate reachable daemon
	}

	dir := t.TempDir()
	memDir := filepath.Join(dir, ".claude-mem")
	if err := os.MkdirAll(memDir, 0700); err != nil {
		t.Fatal(err)
	}
	// Do NOT write settings.json — only supervisor.json with valid JSON.
	if err := os.WriteFile(filepath.Join(memDir, "supervisor.json"), []byte(`{"version":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)

	result := checkMemHealth()
	if !result.Healthy {
		t.Errorf("expected healthy when supervisor.json is valid (settings.json absent), got reason=%q", result.Reason)
	}
}

func TestRunMemHealth_DaemonUnreachable(t *testing.T) {
	orig := daemonProbe
	defer func() { daemonProbe = orig }()
	daemonProbe = func(addr string, d time.Duration) error {
		return fmt.Errorf("connection refused")
	}

	dir := t.TempDir()
	memDir := filepath.Join(dir, ".claude-mem")
	if err := os.MkdirAll(memDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(memDir, "settings.json"), []byte(`{"version":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)

	result := checkMemHealth()
	if result.Healthy {
		t.Error("expected unhealthy when daemon unreachable")
	}
	if result.Reason != "daemon-unreachable" {
		t.Errorf("expected reason=daemon-unreachable, got %q", result.Reason)
	}
}

func TestRunMemHealth_Healthy(t *testing.T) {
	// Spin up a real TCP listener to simulate daemon
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, port, _ := net.SplitHostPort(ln.Addr().String())

	orig := daemonProbe
	defer func() { daemonProbe = orig }()
	realProbe := func(addr string, d time.Duration) error {
		conn, err := net.DialTimeout("tcp", addr, d)
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}
	daemonProbe = realProbe

	dir := t.TempDir()
	memDir := filepath.Join(dir, ".claude-mem")
	if err := os.MkdirAll(memDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(memDir, "settings.json"), []byte(`{"version":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)
	t.Setenv("HARNESS_MEM_HOST", "127.0.0.1")
	t.Setenv("HARNESS_MEM_PORT", port)

	result := checkMemHealth()
	if !result.Healthy {
		t.Errorf("expected healthy, got reason=%q", result.Reason)
	}
}
