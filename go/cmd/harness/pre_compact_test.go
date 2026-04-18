package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluatePreCompact_BlocksMatchingLoopSession(t *testing.T) {
	dir := t.TempDir()
	lockDir := filepath.Join(dir, ".claude", "state", "locks", "loop-session.lock.d")
	if err := os.MkdirAll(lockDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lockDir, "meta.json"), []byte(`{"session_id":"sess-worker"}`), 0600); err != nil {
		t.Fatal(err)
	}

	input := `{"session_id":"sess-worker","cwd":"` + dir + `","agent_type":"worker"}`
	var out bytes.Buffer
	exitCode, err := evaluatePreCompact(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}

	var decision preCompactDecision
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &decision); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if decision.Decision != "block" {
		t.Fatalf("expected decision=block, got %q", decision.Decision)
	}
}

func TestEvaluatePreCompact_AllowsReviewer(t *testing.T) {
	dir := t.TempDir()
	input := `{"session_id":"sess-reviewer","cwd":"` + dir + `","agent_type":"reviewer"}`
	var out bytes.Buffer
	exitCode, err := evaluatePreCompact(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output, got %q", out.String())
	}
}

func TestEvaluatePreCompact_BlocksDirtyPlans(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Harness Test")
	runGit(t, dir, "config", "user.email", "harness@example.com")

	plansPath := filepath.Join(dir, "Plans.md")
	if err := os.WriteFile(plansPath, []byte("initial\n"), 0600); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "Plans.md")
	runGit(t, dir, "commit", "-m", "test: add plans")

	if err := os.WriteFile(plansPath, []byte("dirty\n"), 0600); err != nil {
		t.Fatal(err)
	}

	input := `{"session_id":"sess-main","cwd":"` + dir + `"}`
	var out bytes.Buffer
	exitCode, err := evaluatePreCompact(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
	if !strings.Contains(out.String(), "Plans.md") {
		t.Fatalf("expected Plans.md warning, got %q", out.String())
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}

// TestResolvePreCompactRoot_WalksUpFromSubdir guards against the regression
// where CC launched from a repository subdirectory would search for
// .claude/state/locks/ and Plans.md inside that subdir instead of the
// repository root, causing PreCompact protection to silently no-op for
// monorepo or subpackage layouts.
func TestResolvePreCompactRoot_WalksUpFromSubdir(t *testing.T) {
	repoRoot := t.TempDir()
	runGit(t, repoRoot, "init", "-q")
	subdir := filepath.Join(repoRoot, "packages", "api")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}

	got := resolvePreCompactRoot(subdir)

	// macOS prepends /private to t.TempDir(); compare via filepath.EvalSymlinks
	wantResolved, _ := filepath.EvalSymlinks(repoRoot)
	gotResolved, _ := filepath.EvalSymlinks(got)
	if gotResolved != wantResolved {
		t.Errorf("resolvePreCompactRoot(subdir) = %q (resolved %q), want repo root %q (resolved %q)",
			got, gotResolved, repoRoot, wantResolved)
	}
}
