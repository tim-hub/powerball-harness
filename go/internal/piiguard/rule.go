// Package piiguard provides content-based PII and secret scanning for Harness hooks.
//
// It is the content-layer complement to go/internal/guardrail (which is operation-layer):
//   - guardrail: blocks dangerous tool *operations* (R01-R13: paths, commands, tampering)
//   - piiguard:  blocks dangerous tool *content*  (secrets, PII in prompts and I/O)
//
// Patterns ported from datumbrain/claude-privacy-guard (MIT License).
// See go/internal/piiguard/data/SOURCE.md for attribution details.
package piiguard

import "regexp"

// Severity indicates the risk level of a detected finding.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// Category classifies the type of sensitive data.
type Category string

const (
	// CategorySecret covers API keys, tokens, credentials, and private keys.
	CategorySecret Category = "secret"
	// CategoryPII covers personally identifiable information such as email addresses.
	CategoryPII Category = "pii"
)

// Rule pairs an identifier and metadata with a compiled regex for content scanning.
// Pattern is compiled once at package init via regexp.MustCompile — never nil.
type Rule struct {
	ID       string
	Title    string
	Severity Severity
	Category Category
	Pattern  *regexp.Regexp
	// Validator is an optional post-filter applied to each raw match string.
	// When non-nil, a match is included only when Validator returns true.
	// Use for character-class diversity checks that RE2 cannot express (no lookaheads).
	Validator func(match string) bool
}
