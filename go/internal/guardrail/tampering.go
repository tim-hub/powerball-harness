package guardrail

import (
	"regexp"
)

// ---------------------------------------------------------------------------
// Test tampering patterns (ported from tampering.ts)
// ---------------------------------------------------------------------------

type tamperingPattern struct {
	ID           string
	TaxonomyID   string // FT-TAMPER-NN stable ID; see .claude/rules/failure-taxonomy.md
	Description  string
	Pattern      *regexp.Regexp
	TestFileOnly bool
}

var tamperingPatterns = []tamperingPattern{
	{ID: "T01:it-skip", TaxonomyID: "FT-TAMPER-01", Description: "Test skipping via it.skip / describe.skip",
		Pattern: regexp.MustCompile(`(?:it|test|describe|context)\.skip\s*\(`), TestFileOnly: true},
	{ID: "T02:xit-xdescribe", TaxonomyID: "FT-TAMPER-02", Description: "Test disabling via xit / xdescribe",
		Pattern: regexp.MustCompile(`\b(?:xit|xtest|xdescribe)\s*\(`), TestFileOnly: true},
	{ID: "T03:pytest-skip", TaxonomyID: "FT-TAMPER-03", Description: "Test skipping via pytest.mark.skip",
		Pattern: regexp.MustCompile(`@pytest\.mark\.(?:skip|xfail)\b`), TestFileOnly: true},
	{ID: "T04:go-skip", TaxonomyID: "FT-TAMPER-04", Description: "Test skipping via t.Skip()",
		Pattern: regexp.MustCompile(`\bt\.Skip(?:f|Now)?\s*\(`), TestFileOnly: true},
	{ID: "T05:expect-removed", TaxonomyID: "FT-TAMPER-05", Description: "expect / assert may have been removed (commented out)",
		Pattern: regexp.MustCompile(`//\s*expect\s*\(`), TestFileOnly: true},
	{ID: "T06:assert-commented", TaxonomyID: "FT-TAMPER-06", Description: "assert call was commented out",
		Pattern: regexp.MustCompile(`//\s*assert(?:Equal|NotEqual|True|False|Nil|Error)?\s*\(`), TestFileOnly: true},
	{ID: "T07:todo-assert", TaxonomyID: "FT-TAMPER-07", Description: "Assertion replaced by a TODO comment",
		Pattern: regexp.MustCompile(`(?i)//\s*TODO.*assert|//\s*TODO.*expect`), TestFileOnly: true},
	{ID: "T08:eslint-disable", TaxonomyID: "FT-TAMPER-08", Description: "Lint rule disabled via eslint-disable",
		Pattern: regexp.MustCompile(`(?m)(?://\s*eslint-disable(?:-next-line|-line)?(?:\s+[^\n]+)?$|/\*\s*eslint-disable\b[^*]*\*/)`), TestFileOnly: false},
	{ID: "T09:ci-continue-on-error", TaxonomyID: "FT-TAMPER-09", Description: "CI failure ignored via continue-on-error: true",
		Pattern: regexp.MustCompile(`continue-on-error\s*:\s*true`), TestFileOnly: false},
	{ID: "T10:ci-if-always", TaxonomyID: "FT-TAMPER-10", Description: "CI step forced to run via if: always()",
		Pattern: regexp.MustCompile(`if\s*:\s*always\s*\(\s*\)`), TestFileOnly: false},
	{ID: "T11:hardcoded-answer", TaxonomyID: "FT-TAMPER-11", Description: "Hardcoded test expected values (dictionary return)",
		Pattern: regexp.MustCompile(`answers?_for_tests?\s*=\s*\{`), TestFileOnly: true},
	{ID: "T12:return-hardcoded", TaxonomyID: "FT-TAMPER-12", Description: "Pattern of directly returning test case values",
		Pattern: regexp.MustCompile(`(?i)return\s+(?:"[^"]*"|'[^']*'|\d+)\s*;\s*//.*(?:test|spec|expect)`), TestFileOnly: true},
}

var testFilePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\.test\.[jt]sx?$`),
	regexp.MustCompile(`\.spec\.[jt]sx?$`),
	regexp.MustCompile(`\.test\.py$`),
	regexp.MustCompile(`test_[^/]+\.py$`),
	regexp.MustCompile(`[^/]+_test\.py$`),
	regexp.MustCompile(`\.test\.go$`),
	regexp.MustCompile(`[^/]+_test\.go$`),
	regexp.MustCompile(`/__tests__/`),
	regexp.MustCompile(`/tests/`),
}

var configFilePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:^|/)\.eslintrc(?:\.[^/]+)?$`),
	regexp.MustCompile(`(?:^|/)eslint\.config\.[^/]+$`),
	regexp.MustCompile(`(?:^|/)\.prettierrc(?:\.[^/]+)?$`),
	regexp.MustCompile(`(?:^|/)prettier\.config\.[^/]+$`),
	regexp.MustCompile(`(?:^|/)tsconfig(?:\.[^/]+)?\.json$`),
	regexp.MustCompile(`(?:^|/)biome\.json$`),
	regexp.MustCompile(`(?:^|/)\.stylelintrc(?:\.[^/]+)?$`),
	regexp.MustCompile(`(?:^|/)(?:jest|vitest)\.config\.[^/]+$`),
	regexp.MustCompile(`\.github/workflows/[^/]+\.ya?ml$`),
	regexp.MustCompile(`(?:^|/)\.gitlab-ci\.ya?ml$`),
	regexp.MustCompile(`(?:^|/)Jenkinsfile$`),
}

func isTestFile(filePath string) bool {
	for _, p := range testFilePatterns {
		if p.MatchString(filePath) {
			return true
		}
	}
	return false
}

func isConfigFile(filePath string) bool {
	for _, p := range configFilePatterns {
		if p.MatchString(filePath) {
			return true
		}
	}
	return false
}

type tamperingWarning struct {
	PatternID   string
	TaxonomyID  string // FT-TAMPER-NN; stable ID for trace events and advisor history
	Description string
	MatchedText string
}

func detectTampering(text string, isTest bool) []tamperingWarning {
	var warnings []tamperingWarning
	for _, tp := range tamperingPatterns {
		if tp.TestFileOnly && !isTest {
			continue
		}
		loc := tp.Pattern.FindStringIndex(text)
		if loc != nil {
			matched := text[loc[0]:loc[1]]
			if len(matched) > 120 {
				matched = matched[:120]
			}
			warnings = append(warnings, tamperingWarning{
				PatternID:   tp.ID,
				TaxonomyID:  tp.TaxonomyID,
				Description: tp.Description,
				MatchedText: matched,
			})
		}
	}
	return warnings
}
