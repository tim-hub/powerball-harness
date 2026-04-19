package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunMemHealth_NotInitialized(t *testing.T) {
	// daemonProbe should not be reached when dir is missing
	orig := daemonProbe
	defer func() { daemonProbe = orig }()
	daemonProbe = func(addr string, d time.Duration) error {
		t.Error("daemonProbe should not be called when not initialized")
		return nil
	}

	// Point HOME at a temp dir that has no .claude-mem subdirectory.
	// This guarantees the not-initialized path is taken on any machine.
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	result := checkMemHealth()
	if result.Healthy {
		t.Errorf("expected unhealthy for missing ~/.claude-mem, got healthy")
	}
	if result.Reason != "not-initialized" {
		t.Errorf("expected reason=not-initialized, got %q", result.Reason)
	}
}

func TestRunMemHealth_CorruptedSettings(t *testing.T) {
	orig := daemonProbe
	defer func() { daemonProbe = orig }()
	daemonProbe = func(addr string, d time.Duration) error {
		t.Error("daemonProbe should not be called with corrupted settings")
		return nil
	}

	dir := t.TempDir()
	memDir := filepath.Join(dir, ".claude-mem")
	if err := os.MkdirAll(memDir, 0700); err != nil {
		t.Fatal(err)
	}
	// Write invalid JSON
	if err := os.WriteFile(filepath.Join(memDir, "settings.json"), []byte("not-json{{{"), 0600); err != nil {
		t.Fatal(err)
	}
	// We need to test checkMemHealth with a custom home. The function uses os.UserHomeDir().
	// Patch via HOME env var.
	t.Setenv("HOME", dir)

	result := checkMemHealth()
	if result.Healthy {
		t.Error("expected unhealthy for corrupted settings")
	}
	if result.Reason != "corrupted-settings" {
		t.Errorf("expected reason=corrupted-settings, got %q", result.Reason)
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
