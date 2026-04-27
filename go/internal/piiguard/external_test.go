package piiguard

import (
	"bytes"
	"strings"
	"testing"
)

// TestLoadExternal_AllCompile is the primary DoD gate for task 83.2.
// Verifies the embedded catalog loads, has at least one rule, and stays under 90.
func TestLoadExternal_AllCompile(t *testing.T) {
	rules := LoadExternalCatalog(true)
	if len(rules) == 0 {
		t.Fatal("LoadExternalCatalog returned 0 rules; expected > 0")
	}
	if len(rules) > 90 {
		t.Errorf("LoadExternalCatalog returned %d rules; expected ≤ 90", len(rules))
	}

	for _, r := range rules {
		if r.ID == "" {
			t.Error("found rule with empty ID")
		}
		if !strings.HasPrefix(r.ID, "external-") {
			t.Errorf("rule %q ID missing 'external-' prefix", r.ID)
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

// TestLoadExternal_CodingOnlyFilter verifies the keyword filter drops at least one
// known-irrelevant pattern (e.g. driver license, routing number) when codingOnly=true.
func TestLoadExternal_CodingOnlyFilter(t *testing.T) {
	all := LoadExternalCatalog(false)
	coding := LoadExternalCatalog(true)

	if len(coding) >= len(all) {
		t.Errorf("codingOnly filter must drop ≥ 1 pattern; all=%d coding=%d", len(all), len(coding))
	}

	// California Drivers License is a known prose-only pattern that must be filtered out.
	for _, r := range coding {
		if strings.Contains(strings.ToLower(r.Title), "drivers license") {
			t.Errorf("codingOnly filter should have dropped %q (driver license), but it survived", r.Title)
		}
		if strings.Contains(strings.ToLower(r.Title), "routing number") {
			t.Errorf("codingOnly filter should have dropped %q (routing number), but it survived", r.Title)
		}
	}
}

// TestLoadExternal_MalformedRegex verifies that an entry with an unparseable regex
// is skipped (warning to stderr), not fatal.
func TestLoadExternal_MalformedRegex(t *testing.T) {
	// Mix one malformed entry with one valid entry.
	jsonBlob := []byte(`[
		{"name": "Bad Pattern", "description": "bad token regex", "regex": "([unclosed", "risk": 5, "category": "Confidential"},
		{"name": "Good Token", "description": "good token", "regex": "abc[0-9]+", "risk": 7, "category": "Confidential"}
	]`)

	var warnBuf bytes.Buffer
	rules, err := loadExternalFromBytes(jsonBlob, true, &warnBuf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("want 1 rule (bad one skipped), got %d", len(rules))
	}
	if !strings.Contains(warnBuf.String(), "Bad Pattern") {
		t.Errorf("expected stderr warning about Bad Pattern, got: %q", warnBuf.String())
	}
}

// TestLoadExternal_DuplicateNames verifies that duplicate slugs get numeric suffixes.
func TestLoadExternal_DuplicateNames(t *testing.T) {
	jsonBlob := []byte(`[
		{"name": "DuplicateName", "description": "token rule a", "regex": "aaa[0-9]+", "risk": 5, "category": "Confidential"},
		{"name": "DuplicateName", "description": "token rule b", "regex": "bbb[0-9]+", "risk": 5, "category": "Confidential"},
		{"name": "DuplicateName", "description": "token rule c", "regex": "ccc[0-9]+", "risk": 5, "category": "Confidential"}
	]`)

	var warnBuf bytes.Buffer
	rules, err := loadExternalFromBytes(jsonBlob, true, &warnBuf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("want 3 rules, got %d", len(rules))
	}
	ids := []string{rules[0].ID, rules[1].ID, rules[2].ID}
	wantIDs := []string{"external-duplicatename", "external-duplicatename-2", "external-duplicatename-3"}
	for i, want := range wantIDs {
		if ids[i] != want {
			t.Errorf("ID[%d]: want %q, got %q", i, want, ids[i])
		}
	}
}

// TestLoadExternal_EmptyEntries verifies entries with empty name or regex are skipped.
func TestLoadExternal_EmptyEntries(t *testing.T) {
	jsonBlob := []byte(`[
		{"name": "", "description": "no name", "regex": "abc", "risk": 5, "category": "Confidential"},
		{"name": "no regex", "description": "no regex", "regex": "", "risk": 5, "category": "Confidential"},
		{"name": "Good token", "description": "ok token", "regex": "abc[0-9]+", "risk": 5, "category": "Confidential"}
	]`)

	var warnBuf bytes.Buffer
	rules, _ := loadExternalFromBytes(jsonBlob, true, &warnBuf)
	if len(rules) != 1 {
		t.Errorf("want 1 rule (empties skipped), got %d", len(rules))
	}
}

// TestLoadExternal_MissingFile verifies a nonexistent path returns nil, nil
// (matches upstream semantics — not an error, just an empty load).
func TestLoadExternal_MissingFile(t *testing.T) {
	rules, err := LoadExternal("/nonexistent/path/to/regex.json", true)
	if err != nil {
		t.Errorf("missing file should not error, got: %v", err)
	}
	if rules != nil {
		t.Errorf("want nil rules for missing file, got %d", len(rules))
	}
}

// TestLoadExternal_MalformedJSON verifies a syntactically invalid JSON returns an error.
func TestLoadExternal_MalformedJSON(t *testing.T) {
	jsonBlob := []byte(`{not valid json`)

	var warnBuf bytes.Buffer
	rules, err := loadExternalFromBytes(jsonBlob, true, &warnBuf)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
	if rules != nil {
		t.Errorf("expected nil rules on JSON error, got %d", len(rules))
	}
}

// TestRiskToSeverity covers the upstream score thresholds.
func TestRiskToSeverity(t *testing.T) {
	cases := []struct {
		risk int
		want Severity
	}{
		{0, SeverityLow},
		{2, SeverityLow},
		{3, SeverityMedium},
		{5, SeverityMedium},
		{6, SeverityHigh},
		{7, SeverityHigh},
		{8, SeverityCritical},
		{10, SeverityCritical},
	}
	for _, tc := range cases {
		if got := riskToSeverity(tc.risk); got != tc.want {
			t.Errorf("riskToSeverity(%d): want %q, got %q", tc.risk, tc.want, got)
		}
	}
}

// TestSourceCategoryToCategory covers the case-insensitive PII detection.
func TestSourceCategoryToCategory(t *testing.T) {
	cases := []struct {
		input string
		want  Category
	}{
		{"pii", CategoryPII},
		{"PII", CategoryPII},
		{"Pii", CategoryPII},
		{" pii ", CategoryPII},
		{"Confidential", CategorySecret},
		{"", CategorySecret},
		{"unknown", CategorySecret},
	}
	for _, tc := range cases {
		if got := sourceCategoryToCategory(tc.input); got != tc.want {
			t.Errorf("sourceCategoryToCategory(%q): want %q, got %q", tc.input, tc.want, got)
		}
	}
}

// TestSlugify covers slug normalization.
func TestSlugify(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello-world"},
		{"  Multiple   Spaces  ", "multiple-spaces"},
		{"AWS API Key", "aws-api-key"},
		{"Special!@#$%Chars", "special-chars"},
		{strings.Repeat("a", 100), strings.Repeat("a", 80)},
	}
	for _, tc := range cases {
		if got := slugify(tc.input); got != tc.want {
			t.Errorf("slugify(%q): want %q, got %q", tc.input, tc.want, got)
		}
	}
}

// TestIsCodingSecretPattern checks the keyword filter on synthetic entries.
func TestIsCodingSecretPattern(t *testing.T) {
	cases := []struct {
		entry externalEntry
		want  bool
	}{
		{externalEntry{Name: "OAuth Token", Description: "an oauth bearer", Regex: "abc"}, true},
		{externalEntry{Name: "API Key Pattern", Description: "key match", Regex: "abc"}, true},
		{externalEntry{Name: "California Drivers License", Description: "licence", Regex: "abc"}, false},
		{externalEntry{Name: "Routing Number", Description: "bank", Regex: "abc"}, false},
		{externalEntry{Name: "PGP Block", Description: "pgp encrypted", Regex: "abc"}, true},
	}
	for _, tc := range cases {
		if got := isCodingSecretPattern(tc.entry); got != tc.want {
			t.Errorf("isCodingSecretPattern(%+v): want %v, got %v", tc.entry, tc.want, got)
		}
	}
}
