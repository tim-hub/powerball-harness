package piiguard

import (
	"regexp"
	"strings"
	"testing"
)

// TestScanner_Empty verifies empty inputs produce zero findings + zero risk.
func TestScanner_Empty(t *testing.T) {
	s := NewScanner(BuiltinRules)
	res := s.Scan("")
	if len(res.Findings) != 0 {
		t.Errorf("want 0 findings on empty input, got %d", len(res.Findings))
	}
	if res.RiskScore != 0 {
		t.Errorf("want 0 risk score on empty input, got %d", res.RiskScore)
	}
	if res.Summary == nil {
		t.Error("Summary should be initialised even on empty input")
	}
}

// TestScanner_NilRules verifies a scanner constructed from nil rules
// returns empty results without panicking.
func TestScanner_NilRules(t *testing.T) {
	s := NewScanner(nil)
	res := s.Scan("anything goes here including " + "AKIA" + "IOSFODNN7EXAMPLE")
	if len(res.Findings) != 0 {
		t.Errorf("nil rules should produce 0 findings, got %d", len(res.Findings))
	}
	if res.RiskScore != 0 {
		t.Errorf("nil rules should produce 0 risk, got %d", res.RiskScore)
	}
}

// TestScanner_NoMatches verifies clean text produces no findings.
func TestScanner_NoMatches(t *testing.T) {
	s := NewScanner(BuiltinRules)
	res := s.Scan("This is plain text without any sensitive data.")
	if len(res.Findings) != 0 {
		t.Errorf("want 0 findings on clean input, got %d (%+v)", len(res.Findings), res.Findings)
	}
	if res.RiskScore != 0 {
		t.Errorf("want 0 risk score on clean input, got %d", res.RiskScore)
	}
}

// TestScanner_TableDriven covers the primary DoD requirement: ≥ 5 cases of
// (input, expected finding count, expected risk score).
func TestScanner_TableDriven(t *testing.T) {
	type tcase struct {
		name           string
		input          string
		minFindings    int
		expectScoreCap bool // true → expect score == 100
		expectCategory string
	}

	cases := []tcase{
		{
			name:           "single email (medium PII)",
			input:          "Contact me at user@example.com please.",
			minFindings:    1,
			expectScoreCap: false,
			expectCategory: "pii",
		},
		{
			name:           "single AWS key (critical secret)",
			input:          "config: " + "AKIA" + "IOSFODNN7EXAMPLE end",
			minFindings:    1,
			expectScoreCap: true, // 1 critical = 100 = capped
			expectCategory: "secret",
		},
		{
			name:           "GitHub token (critical secret)",
			input:          "GH_TOKEN=ghp_" + "abcdefghijklmnopqrstuvwxyz12345678901234",
			minFindings:    1,
			expectScoreCap: true,
			expectCategory: "secret",
		},
		{
			name:           "Anthropic key (critical secret)",
			input:          "ANTHROPIC=sk-ant-" + "api03-abcdefghijklmnopqrstuvwxyz",
			minFindings:    1,
			expectScoreCap: true,
			expectCategory: "secret",
		},
		{
			name:           "private key block (critical secret)",
			input:          "-----BEGIN PRIVATE" + " KEY-----\nMIIEvQIBADAN\n-----END PRIVATE" + " KEY-----",
			minFindings:    1,
			expectScoreCap: true,
			expectCategory: "secret",
		},
		{
			name:           "two emails (medium x 2)",
			input:          "From a@a.io and from b@b.io please reply.",
			minFindings:    2,
			expectScoreCap: false, // 2 medium = 40, not capped
			expectCategory: "pii",
		},
	}

	s := NewScanner(BuiltinRules)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := s.Scan(tc.input)
			if len(res.Findings) < tc.minFindings {
				t.Errorf("want ≥ %d findings, got %d (%+v)", tc.minFindings, len(res.Findings), res.Findings)
			}
			if tc.expectScoreCap && res.RiskScore != maxRiskScore {
				t.Errorf("want risk = %d (capped), got %d", maxRiskScore, res.RiskScore)
			}
			if !tc.expectScoreCap && res.RiskScore == 0 {
				t.Errorf("want non-zero risk, got 0")
			}
			if tc.expectCategory != "" && res.Summary[tc.expectCategory] == 0 {
				t.Errorf("want %q in Summary, got %+v", tc.expectCategory, res.Summary)
			}
		})
	}
}

// TestScanner_RiskScoreCap verifies the score never exceeds 100 even with
// many critical hits.
func TestScanner_RiskScoreCap(t *testing.T) {
	// Synthetic rule that matches every "X" in the input.
	rules := []Rule{{
		ID:       "test-cap",
		Title:    "Test Cap",
		Severity: SeverityCritical,
		Category: CategorySecret,
		Pattern:  regexp.MustCompile(`X`),
	}}
	s := NewScanner(rules)
	res := s.Scan("X X X X X X X X X X") // 10 critical hits = 1000 raw → 100 capped
	if res.RiskScore != maxRiskScore {
		t.Errorf("want risk = %d (capped), got %d", maxRiskScore, res.RiskScore)
	}
	if len(res.Findings) != 10 {
		t.Errorf("want 10 findings, got %d", len(res.Findings))
	}
	if res.Summary["secret"] != 10 {
		t.Errorf("want secret count 10, got %d", res.Summary["secret"])
	}
}

// TestScanner_FindingFields verifies each Finding carries correct positional
// and metadata data.
func TestScanner_FindingFields(t *testing.T) {
	rules := []Rule{{
		ID:       "needle",
		Title:    "Needle",
		Severity: SeverityHigh,
		Category: CategorySecret,
		Pattern:  regexp.MustCompile(`needle`),
	}}
	s := NewScanner(rules)
	input := "find the needle in this haystack"
	res := s.Scan(input)
	if len(res.Findings) != 1 {
		t.Fatalf("want 1 finding, got %d", len(res.Findings))
	}
	f := res.Findings[0]
	if f.RuleID != "needle" || f.Title != "Needle" {
		t.Errorf("metadata mismatch: %+v", f)
	}
	if f.Match != "needle" {
		t.Errorf("Match: want %q, got %q", "needle", f.Match)
	}
	if f.StartIndex != 9 {
		t.Errorf("StartIndex: want 9, got %d", f.StartIndex)
	}
	if f.EndIndex != 15 {
		t.Errorf("EndIndex: want 15, got %d", f.EndIndex)
	}
	if input[f.StartIndex:f.EndIndex] != "needle" {
		t.Errorf("indices don't bound the match correctly")
	}
	if res.RiskScore != 50 {
		t.Errorf("want risk 50 (one high), got %d", res.RiskScore)
	}
}

// TestScanner_RedactedValueTruncation verifies long matches get truncated
// with the ellipsis suffix.
func TestScanner_RedactedValueTruncation(t *testing.T) {
	long := strings.Repeat("a", 100)
	if got := redactMatch(long); got != strings.Repeat("a", 50)+"…" {
		t.Errorf("want 50 chars + ellipsis, got %q (len=%d)", got, len(got))
	}

	short := "abc"
	if got := redactMatch(short); got != "abc" {
		t.Errorf("short match should pass through untouched, got %q", got)
	}

	// Exactly at the 50-rune threshold — no ellipsis.
	exact := strings.Repeat("b", 50)
	if got := redactMatch(exact); got != exact {
		t.Errorf("50-char match should pass through, got %q (len=%d runes=%d)", got, len(got), len([]rune(got)))
	}
}

// TestScanner_RedactedValueRuneAware verifies the truncation respects rune
// boundaries (non-ASCII secrets do not get split mid-codepoint).
func TestScanner_RedactedValueRuneAware(t *testing.T) {
	// 60 emoji characters — each is 4 bytes in UTF-8.
	emoji := strings.Repeat("🎯", 60)
	got := redactMatch(emoji)
	// Want: 50 emoji + ellipsis.
	wantPrefix := strings.Repeat("🎯", 50)
	if !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("rune-aware truncation lost emoji boundaries")
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("missing ellipsis suffix")
	}
}

// TestScanner_RealCatalog smoke-tests with the full builtin + external catalog.
func TestScanner_RealCatalog(t *testing.T) {
	rules := append([]Rule{}, BuiltinRules...)
	rules = append(rules, LoadExternalCatalog(true)...)
	s := NewScanner(rules)

	// Plant a few synthetic credentials and confirm the scanner finds them.
	input := "Email: user@example.com, AWS: " + "AKIA" + "IOSFODNN7EXAMPLE"
	res := s.Scan(input)
	if len(res.Findings) < 2 {
		t.Errorf("expected ≥ 2 findings on planted input, got %d", len(res.Findings))
	}
	if res.RiskScore == 0 {
		t.Error("expected non-zero risk on planted input")
	}
	if res.Summary["secret"] == 0 || res.Summary["pii"] == 0 {
		t.Errorf("expected both secret and pii in Summary, got %+v", res.Summary)
	}
}

// TestSeverityWeight covers all severity weights including unknown.
func TestSeverityWeight(t *testing.T) {
	cases := []struct {
		sev  Severity
		want int
	}{
		{SeverityCritical, 100},
		{SeverityHigh, 50},
		{SeverityMedium, 20},
		{SeverityLow, 10},
		{Severity("unknown"), 0},
		{Severity(""), 0},
	}
	for _, tc := range cases {
		if got := severityWeight(tc.sev); got != tc.want {
			t.Errorf("severityWeight(%q): want %d, got %d", tc.sev, tc.want, got)
		}
	}
}
