package guardrail

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------------
// Protected path detection
// ---------------------------------------------------------------------------

// allowedPathPatterns lists paths that look like protected names but are safe
// to write (e.g. templates and examples that must not contain real secrets).
var allowedPathPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:^|/)\.env\.example$`),
	regexp.MustCompile(`(?:^|/)\.env\.template$`),
	regexp.MustCompile(`(?:^|/)\.env\.sample$`),
	// Next.js shared env files: committed to VCS, no secrets.
	// *.local variants (e.g. .env.development.local) remain protected.
	regexp.MustCompile(`(?:^|/)\.env\.development$`),
	regexp.MustCompile(`(?:^|/)\.env\.test$`),
}

var protectedPathPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\.git/`),
	regexp.MustCompile(`/\.git/`),
	regexp.MustCompile(`^\.env$`),
	regexp.MustCompile(`/\.env$`),
	regexp.MustCompile(`\.env\.`),
	regexp.MustCompile(`id_rsa`),
	regexp.MustCompile(`id_ed25519`),
	regexp.MustCompile(`id_ecdsa`),
	regexp.MustCompile(`id_dsa`),
	regexp.MustCompile(`\.pem$`),
	regexp.MustCompile(`\.key$`),
	regexp.MustCompile(`\.p12$`),
	regexp.MustCompile(`\.pfx$`),
	regexp.MustCompile(`authorized_keys`),
	regexp.MustCompile(`known_hosts`),
	// CC 2.1.90: protect git hooks directory (.husky/)
	regexp.MustCompile(`(?:^|/)\.husky(?:/|$)`),
}

// matchesProtectedPatterns reports whether filePath matches any protected pattern.
func matchesProtectedPatterns(filePath string) bool {
	for _, p := range protectedPathPatterns {
		if p.MatchString(filePath) {
			return true
		}
	}
	return false
}

// resolvedPathCache is a bounded cache for filepath.EvalSymlinks results.
// Max 256 entries; when full, the oldest entry is evicted (FIFO).
// Only successful resolutions are cached; errors are never stored.
type resolvedPathCache struct {
	mu   sync.Mutex
	data map[string]string
	keys []string // insertion-order for FIFO eviction
}

const resolvedPathCacheMax = 256

var evalSymlinksCache = &resolvedPathCache{
	data: make(map[string]string, resolvedPathCacheMax),
	keys: make([]string, 0, resolvedPathCacheMax),
}

func (c *resolvedPathCache) get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *resolvedPathCache) set(key, val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.data[key]; exists {
		return // already cached
	}
	if len(c.data) >= resolvedPathCacheMax {
		oldest := c.keys[0]
		c.keys = c.keys[1:]
		delete(c.data, oldest)
	}
	c.data[key] = val
	c.keys = append(c.keys, key)
}

// isProtectedPath checks whether filePath refers to a protected resource.
// It first matches against protectedPathPatterns directly, then resolves
// symlinks via filepath.EvalSymlinks and re-checks the resolved path.
// If EvalSymlinks returns an error (symlink loop, broken link, etc.),
// the function returns true (fail-safe: deny the operation).
// Successful EvalSymlinks resolutions are cached (bounded to 256 entries, FIFO eviction).
// isAllowedPath reports whether filePath is explicitly allowed despite matching
// a protected pattern (e.g. .env.example is safe to write).
func isAllowedPath(filePath string) bool {
	for _, p := range allowedPathPatterns {
		if p.MatchString(filePath) {
			return true
		}
	}
	return false
}

func isProtectedPath(filePath string) bool {
	// Allowlist takes precedence: safe templates/examples are never protected.
	if isAllowedPath(filePath) {
		return false
	}

	// Direct match
	if matchesProtectedPatterns(filePath) {
		return true
	}

	// Check cache before calling EvalSymlinks
	if realPath, ok := evalSymlinksCache.get(filePath); ok {
		return matchesProtectedPatterns(realPath)
	}

	// Resolve symlinks and check the real path (CC 2.1.89: symlink target resolution)
	realPath, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		// Fail-safe: symlink loop, broken link, or other error → deny
		// Exception: if the path simply doesn't exist, it's not protected.
		// We distinguish by checking os.Lstat; if the path doesn't exist at all
		// (not even as a dangling symlink), treat as not protected.
		if _, statErr := os.Lstat(filePath); os.IsNotExist(statErr) {
			return false
		}
		// Any other error (loop, permission denied, etc.) → fail-safe deny (never cache errors)
		return true
	}

	evalSymlinksCache.set(filePath, realPath)
	return matchesProtectedPatterns(realPath)
}

// ---------------------------------------------------------------------------
// Project root check
// ---------------------------------------------------------------------------

func isUnderProjectRoot(filePath, projectRoot string) bool {
	// Resolve relative paths relative to projectRoot
	resolved := filePath
	if !filepath.IsAbs(filePath) {
		resolved = filepath.Join(projectRoot, filePath)
	}
	cleaned := filepath.Clean(resolved)
	root := filepath.Clean(projectRoot)
	if !strings.HasSuffix(root, string(filepath.Separator)) {
		root += string(filepath.Separator)
	}
	return strings.HasPrefix(cleaned, root) || cleaned == root
}

// ---------------------------------------------------------------------------
// Whitespace normalization (CC 2.1.98: wildcard pattern defense-in-depth)
// ---------------------------------------------------------------------------

// wsNormPattern matches one or more whitespace characters (spaces, tabs, etc.)
var wsNormPattern = regexp.MustCompile(`\s+`)

// normalizeCommand collapses consecutive whitespace characters (spaces, tabs,
// and other whitespace) into a single space and trims leading/trailing whitespace.
// This is used as a defense-in-depth measure before wildcard pattern matching,
// so that "git  push  --force" and "git\tpush\t--force" are treated identically
// to "git push --force".
func normalizeCommand(cmd string) string {
	// Fast-path: no tabs, newlines, or consecutive spaces — skip regex entirely
	if !strings.ContainsAny(cmd, "\t\r\n") && !strings.Contains(cmd, "  ") {
		return strings.TrimSpace(cmd) // cheap trim only; may itself return cmd with no alloc if no leading/trailing space
	}
	return strings.TrimSpace(wsNormPattern.ReplaceAllString(cmd, " "))
}

// ---------------------------------------------------------------------------
// Dangerous rm -rf detection
// ---------------------------------------------------------------------------

var rmRecursivePattern = regexp.MustCompile(`\brm\s+--recursive\b`)

// rmRfManual detects rm with both -r and -f flags (in any order/combination).
// Go regexp doesn't support lookahead (?=...) so we check manually.
var rmWithFlags = regexp.MustCompile(`\brm\s+(.+)`)

func hasDangerousRmRf(command string) bool {
	// Normalize whitespace before matching (CC 2.1.98: defense-in-depth)
	command = normalizeCommand(command)
	if rmRecursivePattern.MatchString(command) {
		return true
	}
	// Check for -rf, -fr, -r -f, etc. in rm arguments
	m := rmWithFlags.FindStringSubmatch(command)
	if m == nil {
		return false
	}
	args := m[1]
	// Scan tokens for flag groups containing both r and f
	hasR := false
	hasF := false
	for _, token := range strings.Fields(args) {
		if !strings.HasPrefix(token, "-") || strings.HasPrefix(token, "--") {
			continue // skip non-short-flags and long flags
		}
		flags := token[1:] // strip leading -
		for _, c := range flags {
			if c == 'r' {
				hasR = true
			}
			if c == 'f' {
				hasF = true
			}
		}
	}
	return hasR && hasF
}

// ---------------------------------------------------------------------------
// git push --force detection
// ---------------------------------------------------------------------------

var (
	forcePushPattern  = regexp.MustCompile(`\bgit\s+push\b.*--force(?:-with-lease)?\b`)
	forcePushShort    = regexp.MustCompile(`\bgit\s+push\b.*-f\b`)
)

func hasForcePush(command string) bool {
	// Normalize whitespace before matching (CC 2.1.98: defense-in-depth)
	command = normalizeCommand(command)
	return forcePushPattern.MatchString(command) || forcePushShort.MatchString(command)
}

// ---------------------------------------------------------------------------
// sudo detection
// ---------------------------------------------------------------------------

var sudoPattern = regexp.MustCompile(`(?:^|\s)sudo\s`)

func hasSudo(command string) bool {
	command = normalizeCommand(command)
	return sudoPattern.MatchString(command)
}

// ---------------------------------------------------------------------------
// --no-verify / --no-gpg-sign detection
// ---------------------------------------------------------------------------

var (
	noVerifyPattern  = regexp.MustCompile(`(?:^|\s)--no-verify(?:\s|$)`)
	noGpgSignPattern = regexp.MustCompile(`(?:^|\s)--no-gpg-sign(?:\s|$)`)
)

func hasDangerousGitBypassFlag(command string) bool {
	command = normalizeCommand(command)
	return noVerifyPattern.MatchString(command) || noGpgSignPattern.MatchString(command)
}

// ---------------------------------------------------------------------------
// Protected branch reset --hard detection
// ---------------------------------------------------------------------------

var protectedBranchRefPattern = regexp.MustCompile(
	`^(?:origin/|upstream/)?(?:refs/heads/)?(?:main|master)(?:[~^]\d+)?$`,
)

func normalizeGitToken(token string) string {
	return strings.Trim(token, "'\"")
}

func hasProtectedBranchResetHard(command string) bool {
	command = normalizeCommand(command)
	tokens := strings.Fields(command)
	resetIndex := -1
	hasHard := false
	for i, t := range tokens {
		normalized := normalizeGitToken(t)
		if normalized == "reset" {
			resetIndex = i
		}
		if normalized == "--hard" {
			hasHard = true
		}
	}
	if resetIndex == -1 || !hasHard {
		return false
	}
	for _, t := range tokens[resetIndex+1:] {
		normalized := normalizeGitToken(t)
		if strings.HasPrefix(normalized, "-") {
			continue
		}
		if protectedBranchRefPattern.MatchString(normalized) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Direct push to protected branch detection
// ---------------------------------------------------------------------------

var gitPushPattern = regexp.MustCompile(`\bgit\s+push\b`)

func hasDirectPushToProtectedBranch(command string) bool {
	command = normalizeCommand(command)
	if !gitPushPattern.MatchString(command) {
		return false
	}
	tokens := strings.Fields(command)
	pushIndex := -1
	for i, t := range tokens {
		if t == "push" {
			pushIndex = i
			break
		}
	}
	if pushIndex == -1 {
		return false
	}

	// Collect non-flag args after "push"
	var args []string
	for _, t := range tokens[pushIndex+1:] {
		if !strings.HasPrefix(t, "-") {
			args = append(args, t)
		}
	}
	if len(args) == 0 {
		return false
	}

	for _, arg := range args {
		normalized := normalizeGitToken(arg)
		if protectedBranchRefPattern.MatchString(normalized) {
			return true
		}
		// Check refspec (src:dst)
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) == 2 {
			if protectedBranchRefPattern.MatchString(normalizeGitToken(parts[1])) {
				return true
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Protected review path detection (warn-only)
// ---------------------------------------------------------------------------

var protectedReviewPathPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:^|/)package\.json$`),
	regexp.MustCompile(`(?:^|/)Dockerfile$`),
	regexp.MustCompile(`(?:^|/)docker-compose\.yml$`),
	regexp.MustCompile(`(?:^|/)\.github/workflows/[^/]+$`),
	regexp.MustCompile(`(?:^|/)schema\.prisma$`),
	regexp.MustCompile(`(?:^|/)wrangler\.toml$`),
	regexp.MustCompile(`(?:^|/)index\.html$`),
}

func isProtectedReviewPath(filePath string) bool {
	for _, p := range protectedReviewPathPatterns {
		if p.MatchString(filePath) {
			return true
		}
	}
	return false
}
