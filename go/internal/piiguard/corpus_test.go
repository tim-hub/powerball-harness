package piiguard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCorpus_Positive verifies every fixture in testdata/positive/ produces
// at least one finding when scanned with the full builtin + external catalog.
// This is the regression gate for the rule set: a future change that drops a
// rule (or breaks a pattern) will fail here loudly.
func TestCorpus_Positive(t *testing.T) {
	rules := append([]Rule{}, BuiltinRules...)
	rules = append(rules, LoadExternalCatalog(true)...)
	s := NewScanner(rules)

	files, err := filepath.Glob("testdata/positive/*.txt")
	if err != nil {
		t.Fatalf("glob testdata/positive: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("testdata/positive/ contains no fixtures — corpus is empty")
	}

	for _, path := range files {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			res := s.Scan(string(data))
			if len(res.Findings) == 0 {
				t.Errorf("expected ≥ 1 finding in %s, got 0", name)
			}
			if res.RiskScore == 0 {
				t.Errorf("expected non-zero risk score in %s", name)
			}
		})
	}
}

// TestCorpus_Negative verifies every fixture in testdata/negative/ produces
// zero findings.  Catches regex over-matching that would create false-positive
// blocks in production.
func TestCorpus_Negative(t *testing.T) {
	rules := append([]Rule{}, BuiltinRules...)
	rules = append(rules, LoadExternalCatalog(true)...)
	s := NewScanner(rules)

	files, err := filepath.Glob("testdata/negative/*.txt")
	if err != nil {
		t.Fatalf("glob testdata/negative: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("testdata/negative/ contains no fixtures")
	}

	for _, path := range files {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			res := s.Scan(string(data))
			if len(res.Findings) > 0 {
				var titles []string
				for _, f := range res.Findings {
					titles = append(titles, f.Title)
				}
				t.Errorf("expected 0 findings in %s, got %d (rules: %s)",
					name, len(res.Findings), strings.Join(titles, ", "))
			}
		})
	}
}
