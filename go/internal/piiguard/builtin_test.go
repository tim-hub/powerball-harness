package piiguard

import "testing"

// TestBuiltinRules_AllCompile verifies rule count, non-nil patterns, and basic metadata.
// This is the primary DoD gate for task 83.1.
func TestBuiltinRules_AllCompile(t *testing.T) {
	const minRules = 13
	if got := len(BuiltinRules); got < minRules {
		t.Errorf("want at least %d built-in rules, got %d", minRules, got)
	}

	for _, r := range BuiltinRules {
		if r.ID == "" {
			t.Error("found rule with empty ID")
		}
		if r.Title == "" {
			t.Errorf("rule %q has empty Title", r.ID)
		}
		if r.Pattern == nil {
			t.Errorf("rule %q has nil Pattern", r.ID)
		}
		if r.Severity == "" {
			t.Errorf("rule %q has empty Severity", r.ID)
		}
		if r.Category == "" {
			t.Errorf("rule %q has empty Category", r.ID)
		}
	}
}

// TestBuiltinRules_PatternFixtures verifies each pattern matches a synthetic positive
// sample and rejects a negative sample.  All values are fabricated — not real credentials.
//
// Sensitive fixture strings are built via concatenation so static secret scanners do not
// see the full pattern as contiguous bytes in this source file.  At runtime, Go's + operator
// assembles the correct match string before the regex is applied.
func TestBuiltinRules_PatternFixtures(t *testing.T) {
	type fixture struct {
		ruleID   string
		positive string // must match
		negative string // must NOT match
	}

	fixtures := []fixture{
		{
			ruleID:   "email-address",
			positive: "Please contact user@example.com for help",
			negative: "no at-sign here",
		},
		{
			// Split both eyJ prefixes so the JWT pattern is not contiguous in source.
			ruleID:   "jwt-token",
			positive: "eyJ" + "hbGciOiJIUzI1NiJ9." + "eyJ" + "zdWIiOiJ0ZXN0In0.xyzABC123def",
			negative: "not-a-jwt-token",
		},
		{
			// Split Bearer from its token value.
			ruleID:   "bearer-token",
			positive: "Authorization: Bearer " + "abc123def456ghi789jkl012mno345",
			negative: "Authorization: Basic abc123def456",
		},
		{
			// Split after the AKIA prefix.
			ruleID:   "aws-api-key",
			positive: "export AWS_KEY=" + "AKIA" + "IOSFODNN7EXAMPLE",
			negative: "export FOO=NOTAWSKEY12345678901234",
		},
		{
			// Split across string literals so the scanner never sees a complete key.
			// Real keys are sk-proj-<48 chars> — require {40,} to avoid sk-learn, etc.
			ruleID:   "openai-api-key",
			positive: "OPENAI_KEY=sk-" + "proj-abcdefghijklmnopqrstuvwxyz0123456789ABCDEF01",
			negative: "no key here at all",
		},
		{
			// Split after sk-ant- prefix.
			ruleID:   "anthropic-api-key",
			positive: "ANTHROPIC_KEY=sk-ant-" + "api03-abcdefghijklmnopqrstuvwxyz",
			negative: "sk-notanthropic1234",
		},
		{
			// Split after sk-or- prefix.
			ruleID:   "openrouter-api-key",
			positive: "OR_KEY=sk-or-" + "v1-abcdefghijklmnopqrstuvwxyz1234567",
			negative: "sk-or-v1-tooshort",
		},
		{
			// Split after AIza prefix.
			ruleID:   "google-ai-api-key",
			positive: "GOOGLE_KEY=" + "AIza" + "SyA1234567890ABCDEfghijklmnopqrstUV",
			negative: "GOOGLE_KEY=notgooglekeyformat",
		},
		{
			// Split after gsk_ prefix.
			ruleID:   "groq-api-key",
			positive: "GROQ_KEY=gsk_" + "abcdefghijklmnopqrstuvwxyz1234567890abcdefgh",
			negative: "gsk_tooshort",
		},
		{
			// Split after pplx- prefix.
			ruleID:   "perplexity-api-key",
			positive: "PPX_KEY=pplx-" + "abcdefghijklmnopqrstuvwxyz12345",
			negative: "pplx-tooshort",
		},
		{
			// Split after hf_ prefix.
			ruleID:   "huggingface-api-token",
			positive: "HF_TOKEN=hf_" + "abcdefghijklmnopqrstuvwxyz1234567890",
			negative: "hf_tooshort",
		},
		{
			// Split between sk_ and live_ so sk_live_ is not contiguous in source.
			ruleID:   "stripe-api-key",
			positive: "STRIPE_KEY=sk_" + "live_abcdefghijklmnopqrstuvwxyz123456",
			negative: "sk_live_short",
		},
		{
			// Split after ghp_ prefix.
			ruleID:   "github-token",
			positive: "GH_TOKEN=ghp_" + "abcdefghijklmnopqrstuvwxyz12345678901234",
			negative: "gh_not_a_real_token_format",
		},
		{
			// Split so "api_key = \"..." is not contiguous in source.
			ruleID:   "generic-code-secret-assignment",
			positive: "api_key" + ` = "abc1234567890secretvalue12345"`,
			negative: `username = "john"`,
		},
		{
			// Split "PRIVATE KEY" so the PEM header is not contiguous in source.
			ruleID:   "private-key",
			positive: "-----BEGIN PRIVATE" + " KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAAS\n-----END PRIVATE" + " KEY-----",
			negative: "no private key material here",
		},
	}

	// Build lookup index by rule ID.
	ruleByID := make(map[string]*Rule, len(BuiltinRules))
	for i := range BuiltinRules {
		ruleByID[BuiltinRules[i].ID] = &BuiltinRules[i]
	}

	for _, tc := range fixtures {
		t.Run(tc.ruleID, func(t *testing.T) {
			r, ok := ruleByID[tc.ruleID]
			if !ok {
				t.Fatalf("rule %q not found in BuiltinRules", tc.ruleID)
			}
			if !r.Pattern.MatchString(tc.positive) {
				t.Errorf("rule %q did not match positive sample: %q", tc.ruleID, tc.positive)
			}
			if r.Pattern.MatchString(tc.negative) {
				t.Errorf("rule %q unexpectedly matched negative sample: %q", tc.ruleID, tc.negative)
			}
		})
	}
}

// TestBuiltinRules_SeverityCategories verifies severity and category assignments
// match upstream intent: cloud credentials should be critical, PII should be pii.
func TestBuiltinRules_SeverityCategories(t *testing.T) {
	criticalSecrets := map[string]bool{
		"aws-api-key":         true,
		"anthropic-api-key":   true,
		"openai-api-key":      true,
		"openrouter-api-key":  true,
		"google-ai-api-key":   true,
		"groq-api-key":        true,
		"perplexity-api-key":  true,
		"stripe-api-key":      true,
		"github-token":        true,
		"private-key":         true,
	}

	for _, r := range BuiltinRules {
		if criticalSecrets[r.ID] {
			if r.Severity != SeverityCritical {
				t.Errorf("rule %q: want critical severity, got %q", r.ID, r.Severity)
			}
			if r.Category != CategorySecret {
				t.Errorf("rule %q: want secret category, got %q", r.ID, r.Category)
			}
		}
	}

	// email is the only PII rule; all others are secret.
	for _, r := range BuiltinRules {
		if r.ID == "email-address" {
			if r.Category != CategoryPII {
				t.Errorf("email-address rule: want pii category, got %q", r.Category)
			}
		} else if r.Category != CategorySecret {
			t.Errorf("rule %q: want secret category, got %q", r.ID, r.Category)
		}
	}
}

// TestBearerToken_FalsePositives verifies that the tightened bearer-token rule
// rejects short tokens (below the 20-char floor) and all-lowercase tokens
// (rejected by bearerTokenValidator).  Inputs use concatenation so the raw source
// does not contain a contiguous trigger string.
func TestBearerToken_FalsePositives(t *testing.T) {
	scanner := NewScanner(BuiltinRules)

	negatives := []struct {
		name  string
		input string
	}{
		{
			name:  "below-floor-1char",
			input: "Authorization: Bearer " + "x",
		},
		{
			name:  "below-floor-3chars",
			input: "Authorization: Bearer " + "abc",
		},
		{
			// 25 lowercase-only chars: pattern matches ({20,}) but Validator rejects.
			name:  "all-lowercase-25chars",
			input: "Authorization: Bearer " + "abcdefghijklmnopqrstuvwxy",
		},
	}

	for _, tc := range negatives {
		t.Run(tc.name, func(t *testing.T) {
			res := scanner.Scan(tc.input)
			for _, f := range res.Findings {
				if f.RuleID == "bearer-token" {
					t.Errorf("bearer-token rule false-positive on %q: got finding %q", tc.name, f.RedactedValue)
				}
			}
		})
	}
}
