package guardrail

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tim-hub/powerball-harness/go/pkg/config"
	"github.com/tim-hub/powerball-harness/go/pkg/hookproto"
)

// resolveProtectedPathAskList reads [[safety.guardrail.protectedPathAskList]] from
// the project-local harness.toml only. pluginRoot is intentionally ignored so that
// break-glass entries are always scoped to the repository being edited.
func resolveProtectedPathAskList(_ hookproto.HookInput, projectRoot string) []hookproto.ProtectedPathAskEntry {
	return readProtectedPathAskListFromHarnessTOML(harnessTOMLPath(projectRoot, ""))
}

func readProtectedPathAskListFromHarnessTOML(path string) []hookproto.ProtectedPathAskEntry {
	cfg, err := config.ParseFile(path)
	if err != nil || cfg == nil {
		return nil
	}
	entries := make([]hookproto.ProtectedPathAskEntry, 0, len(cfg.Safety.Guardrail.ProtectedPathAskList))
	for _, e := range cfg.Safety.Guardrail.ProtectedPathAskList {
		entries = append(entries, hookproto.ProtectedPathAskEntry{
			Path:   e.Path,
			Reason: e.Reason,
		})
	}
	return entries
}

// r03EnvShellWriteTargetRe extracts file targets from shell redirection and tee.
// Matches: > file, >> file, tee file, tee -a file.
// sed -i and similar in-place operators are intentionally NOT matched.
var r03EnvShellWriteTargetRe = regexp.MustCompile(
	`(?:>>?|tee(?:\s+-a)?)\s+(\S+)`,
)

// r03BreakGlassDenyPaths are hard-deny paths that must never be unlocked via ask-list,
// even if the operator adds them to [[safety.guardrail.protectedPathAskList]].
var r03BreakGlassDenyPaths = []*regexp.Regexp{
	regexp.MustCompile(`^\.git/`),
	regexp.MustCompile(`/\.git/`),
	regexp.MustCompile(`(?:^|/)secrets/`),
	regexp.MustCompile(`\.pem$`),
	regexp.MustCompile(`\.key$`),
	regexp.MustCompile(`authorized_keys`),
	regexp.MustCompile(`(?:^|/)\.(?:zshrc|bashrc|bash_profile|profile)$`),
	regexp.MustCompile(`(?:^|/)\.husky(?:/|$)`),
	regexp.MustCompile(`(?:^|/)\.claude/hooks/`),
}

// isHardDenyTarget reports whether target matches a path that must never be
// softened to an Ask via the break-glass ask-list.
func isHardDenyTarget(target string) bool {
	for _, p := range r03BreakGlassDenyPaths {
		if p.MatchString(target) {
			return true
		}
	}
	return false
}

// canonicalProtectedPathAskPath normalises a path from the ask-list entry.
// Relative paths are kept as-is (matched against the raw target token).
// Returns ("", false) for unsafe patterns (relative traversal, absolute outside project).
func canonicalProtectedPathAskPath(entryPath, projectRoot string) (string, bool) {
	if entryPath == "" {
		return "", false
	}
	if strings.HasPrefix(entryPath, "..") || strings.Contains(entryPath, "/../") {
		return "", false
	}
	if filepath.IsAbs(entryPath) {
		cleaned := filepath.Clean(entryPath)
		root := filepath.Clean(projectRoot)
		if !strings.HasPrefix(cleaned, root+string(filepath.Separator)) && cleaned != root {
			return "", false
		}
		// Convert to relative for comparison with extracted shell tokens
		rel, err := filepath.Rel(root, cleaned)
		if err != nil {
			return "", false
		}
		return rel, true
	}
	return entryPath, true
}

// r03EnvBreakGlassCheckBashCommand checks whether command writes to a path listed
// in ctx.ProtectedPathAskList. Returns a HookResult if a match is found, nil otherwise.
// Hard-deny targets are checked across ALL write targets before any Ask is returned,
// so a mixed command (safe target + hard-deny target) is always denied.
func r03EnvBreakGlassCheckBashCommand(ctx hookproto.RuleContext, command string) *hookproto.HookResult {
	if len(ctx.ProtectedPathAskList) == 0 {
		return nil
	}

	matches := r03EnvShellWriteTargetRe.FindAllStringSubmatch(command, -1)
	if len(matches) == 0 {
		return nil
	}

	// Collect non-empty targets.
	var targets []string
	for _, m := range matches {
		if t := strings.TrimSpace(m[1]); t != "" {
			targets = append(targets, t)
		}
	}

	// Pass 1: hard-deny check across all targets — must run before any Ask.
	for _, target := range targets {
		if isHardDenyTarget(target) {
			return &hookproto.HookResult{
				Decision: hookproto.DecisionDeny,
				Reason:   "Shell writes to protected files are prohibited.",
			}
		}
	}

	// Pass 2: ask-list match.
	for _, target := range targets {
		for _, entry := range ctx.ProtectedPathAskList {
			if entry.Reason == "" || strings.TrimSpace(entry.Reason) == "" {
				return &hookproto.HookResult{
					Decision: hookproto.DecisionDeny,
					Reason:   "Shell writes to protected files are prohibited.",
				}
			}

			canonical, ok := canonicalProtectedPathAskPath(entry.Path, ctx.ProjectRoot)
			if !ok {
				continue
			}

			if target == canonical || filepath.Clean(target) == filepath.Clean(canonical) {
				return &hookproto.HookResult{
					Decision: hookproto.DecisionAsk,
					Reason: "R03: Writing to " + target + " requires confirmation.\n" +
						"Configured in harness.toml [[safety.guardrail.protectedPathAskList]].\n" +
						"Reason: " + strings.TrimSpace(entry.Reason),
				}
			}
		}
	}

	return nil
}
