package guardrail

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// isProtectedPath — static pattern tests
// ---------------------------------------------------------------------------

func TestIsProtectedPath_Env(t *testing.T) {
	if !isProtectedPath(".env") {
		t.Error(".env should be protected")
	}
}

func TestIsProtectedPath_EnvVariant(t *testing.T) {
	if !isProtectedPath(".env.local") {
		t.Error(".env.local should be protected")
	}
}

func TestIsProtectedPath_GitDir(t *testing.T) {
	if !isProtectedPath(".git/config") {
		t.Error(".git/config should be protected")
	}
}

func TestIsProtectedPath_IdRsa(t *testing.T) {
	if !isProtectedPath("/home/user/.ssh/id_rsa") {
		t.Error("id_rsa should be protected")
	}
}

func TestIsProtectedPath_NormalFile(t *testing.T) {
	if isProtectedPath("/project/src/main.go") {
		t.Error("normal source file should NOT be protected")
	}
}

// ---------------------------------------------------------------------------
// Task 38.1.1: .husky protection (CC 2.1.90)
// ---------------------------------------------------------------------------

func TestIsProtectedPath_HuskyPreCommit(t *testing.T) {
	if !isProtectedPath("/project/.husky/pre-commit") {
		t.Error(".husky/pre-commit should be protected")
	}
}

func TestIsProtectedPath_HuskyNested(t *testing.T) {
	if !isProtectedPath("/project/.husky/hooks/commit-msg") {
		t.Error(".husky/hooks/commit-msg should be protected")
	}
}

func TestIsProtectedPath_HuskyRoot(t *testing.T) {
	// Just the .husky directory itself
	if !isProtectedPath(".husky/") {
		t.Error(".husky/ should be protected")
	}
}

// ---------------------------------------------------------------------------
// Task 38.1.1: symlink resolution tests (CC 2.1.89)
// ---------------------------------------------------------------------------

func TestIsProtectedPath_SymlinkToEnv(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, ".env")
	if err := os.WriteFile(target, []byte("SECRET=1"), 0600); err != nil {
		t.Fatalf("failed to create .env: %v", err)
	}
	link := filepath.Join(tmp, "link-env")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	if !isProtectedPath(link) {
		t.Errorf("symlink to .env should be protected; link=%s target=%s", link, target)
	}
}

func TestIsProtectedPath_NestedSymlink(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, ".env")
	if err := os.WriteFile(target, []byte("x"), 0600); err != nil {
		t.Fatalf("failed to create .env: %v", err)
	}
	link2 := filepath.Join(tmp, "link2")
	if err := os.Symlink(target, link2); err != nil {
		t.Fatalf("failed to create link2: %v", err)
	}
	link1 := filepath.Join(tmp, "link1")
	if err := os.Symlink(link2, link1); err != nil {
		t.Fatalf("failed to create link1: %v", err)
	}

	if !isProtectedPath(link1) {
		t.Errorf("nested symlink should resolve to .env and be protected")
	}
}

func TestIsProtectedPath_SymlinkLoop(t *testing.T) {
	tmp := t.TempDir()
	a := filepath.Join(tmp, "a")
	b := filepath.Join(tmp, "b")
	// Create a → b → a loop
	if err := os.Symlink(b, a); err != nil {
		t.Fatalf("failed to create symlink a: %v", err)
	}
	if err := os.Symlink(a, b); err != nil {
		t.Fatalf("failed to create symlink b: %v", err)
	}

	// Fail-safe: symlink loop should be denied (true)
	if !isProtectedPath(a) {
		t.Errorf("symlink loop should fail-safe to protected (deny)")
	}
}

func TestIsProtectedPath_SymlinkToNormalFile(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "normal.txt")
	if err := os.WriteFile(target, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create normal.txt: %v", err)
	}
	link := filepath.Join(tmp, "link-normal")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Symlink to a normal file should NOT be protected
	if isProtectedPath(link) {
		t.Errorf("symlink to normal file should NOT be protected")
	}
}

func TestIsProtectedPath_NonExistentPath(t *testing.T) {
	// Non-existent path that doesn't match patterns should NOT be protected
	if isProtectedPath("/nonexistent/totally/random/path.go") {
		t.Error("non-existent non-protected path should NOT be protected")
	}
}
