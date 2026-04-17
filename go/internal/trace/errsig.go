package trace

import (
	"regexp"
	"strings"
)

// Normalization regexes are ordered so that more-specific patterns run before
// less-specific ones. Running numeric-sequence stripping last prevents it
// from eating digits inside UUIDs or hex addresses before they match.
var (
	// hexAddr catches 0x-prefixed addresses (e.g. 0xDEADBEEF) commonly seen
	// in Go panics and C/C++ error output.
	hexAddr = regexp.MustCompile(`0x[0-9a-fA-F]+`)

	// uuidCanonical matches 8-4-4-4-12 UUID strings with or without dashes.
	uuidCanonical = regexp.MustCompile(`[0-9a-fA-F]{8}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{4}-?[0-9a-fA-F]{12}`)

	// tmpPath catches run-specific temp paths on macOS and Linux.
	tmpPath = regexp.MustCompile(`(?:/private)?/(?:tmp|var/folders)/[^\s]+`)

	// commitSha matches 7–64 char hex tokens: git short SHAs (7), full SHA-1
	// (40), and SHA-256 (64). Word boundaries prevent it from eating tokens
	// that happen to contain hex-looking substrings.
	commitSha = regexp.MustCompile(`\b[0-9a-fA-F]{7,64}\b`)

	// numericSeq matches runs of digits: line numbers, PIDs, timestamps,
	// port numbers, etc. Runs last so it doesn't eat digits from other
	// patterns before they can match.
	numericSeq = regexp.MustCompile(`[0-9]+`)

	// whitespace collapses any run of whitespace into one space.
	whitespace = regexp.MustCompile(`\s+`)
)

// maxSigLen is the final truncation applied to a normalized signature.
const maxSigLen = 200

// NormalizeErrorSignature transforms a raw error string into a stable
// signature: the same logical error across runs yields the same output.
// This enables duplicate suppression in the Advisor (which caches decisions
// keyed on (task_id, reason_code, error_signature)) and trend analysis across
// traces.
//
// Rules (matching .claude/memory/schemas/trace.v1.md line 167):
//  1. Lowercase the string
//  2. Strip hex addresses (0xDEADBEEF)
//  3. Strip UUIDs
//  4. Replace run-specific tmp paths with "<tmp>/"
//  5. Strip commit SHAs and other long hex tokens
//  6. Strip numeric sequences (line numbers, PIDs, etc.)
//  7. Collapse whitespace, trim, truncate to maxSigLen
//
// Empty input yields empty output.
func NormalizeErrorSignature(raw string) string {
	if raw == "" {
		return ""
	}
	s := strings.ToLower(raw)
	s = hexAddr.ReplaceAllString(s, "")
	s = uuidCanonical.ReplaceAllString(s, "")
	s = tmpPath.ReplaceAllString(s, "<tmp>/")
	s = commitSha.ReplaceAllString(s, "")
	s = numericSeq.ReplaceAllString(s, "")
	s = whitespace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	if len(s) > maxSigLen {
		s = s[:maxSigLen]
	}
	return s
}
