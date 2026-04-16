package guardrail

import (
	"fmt"
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

// ---------------------------------------------------------------------------
// Task 62.2: EvalSymlinks cache tests
// ---------------------------------------------------------------------------

func TestEvalSymlinksCache(t *testing.T) {
	// Reset cache to known state before test
	evalSymlinksCache.mu.Lock()
	evalSymlinksCache.data = make(map[string]string, resolvedPathCacheMax)
	evalSymlinksCache.keys = make([]string, 0, resolvedPathCacheMax)
	evalSymlinksCache.mu.Unlock()

	tmp := t.TempDir()
	target := filepath.Join(tmp, "normal.txt")
	if err := os.WriteFile(target, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create normal.txt: %v", err)
	}
	link := filepath.Join(tmp, "link-normal")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// First call: cache miss → calls EvalSymlinks → populates cache
	isProtectedPath(link)

	evalSymlinksCache.mu.Lock()
	cacheLen := len(evalSymlinksCache.data)
	_, cached := evalSymlinksCache.data[link]
	evalSymlinksCache.mu.Unlock()

	if cacheLen != 1 {
		t.Errorf("expected 1 cache entry after first call, got %d", cacheLen)
	}
	if !cached {
		t.Errorf("expected link path to be cached after first call")
	}

	// Second call: cache hit → should not grow cache
	isProtectedPath(link)

	evalSymlinksCache.mu.Lock()
	cacheLenAfter := len(evalSymlinksCache.data)
	evalSymlinksCache.mu.Unlock()

	if cacheLenAfter != 1 {
		t.Errorf("expected 1 cache entry after second call (cache hit), got %d", cacheLenAfter)
	}
}

func TestEvalSymlinksCache_Eviction(t *testing.T) {
	// Reset cache to known state before test
	evalSymlinksCache.mu.Lock()
	evalSymlinksCache.data = make(map[string]string, resolvedPathCacheMax)
	evalSymlinksCache.keys = make([]string, 0, resolvedPathCacheMax)
	evalSymlinksCache.mu.Unlock()

	tmp := t.TempDir()

	// Fill cache to exactly 256 entries using set() directly
	firstKey := fmt.Sprintf("key-%d", 0)
	for i := 0; i < resolvedPathCacheMax; i++ {
		evalSymlinksCache.set(fmt.Sprintf("key-%d", i), fmt.Sprintf("val-%d", i))
	}

	evalSymlinksCache.mu.Lock()
	lenBefore := len(evalSymlinksCache.data)
	evalSymlinksCache.mu.Unlock()

	if lenBefore != resolvedPathCacheMax {
		t.Fatalf("expected %d entries before eviction, got %d", resolvedPathCacheMax, lenBefore)
	}

	// Add one more via a real file symlink resolution (triggers set() internally)
	target := filepath.Join(tmp, "extra.txt")
	if err := os.WriteFile(target, []byte("x"), 0644); err != nil {
		t.Fatalf("failed to create extra.txt: %v", err)
	}
	link := filepath.Join(tmp, "link-extra")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}
	isProtectedPath(link)

	evalSymlinksCache.mu.Lock()
	lenAfter := len(evalSymlinksCache.data)
	_, firstStillPresent := evalSymlinksCache.data[firstKey]
	_, newPresent := evalSymlinksCache.data[link]
	evalSymlinksCache.mu.Unlock()

	if lenAfter != resolvedPathCacheMax {
		t.Errorf("cache should remain at %d after eviction, got %d", resolvedPathCacheMax, lenAfter)
	}
	if firstStillPresent {
		t.Errorf("oldest entry (key-0) should have been evicted")
	}
	if !newPresent {
		t.Errorf("new entry (link) should be present after eviction")
	}
}

// ---------------------------------------------------------------------------
// Task 62.4: normalizeCommand fast-path tests
// ---------------------------------------------------------------------------

func TestNormalizeCommandFastPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"git push", "git push", "fast-path: already normalized"},
		{"git  push", "git push", "slow-path: double space"},
		{"git\tpush", "git push", "slow-path: tab"},
		{"  git push  ", "git push", "fast-path with trim: leading/trailing spaces"},
		{"git\r\npush", "git push", "slow-path: carriage return + newline"},
		{"git push --force", "git push --force", "fast-path: multi-word already normalized"},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := normalizeCommand(tc.input)
			if got != tc.expected {
				t.Errorf("normalizeCommand(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
