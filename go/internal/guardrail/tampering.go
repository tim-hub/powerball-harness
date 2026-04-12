package guardrail

import (
	"regexp"
)

// ---------------------------------------------------------------------------
// Test tampering patterns (ported from tampering.ts)
// ---------------------------------------------------------------------------

type tamperingPattern struct {
	ID           string
	Description  string
	Pattern      *regexp.Regexp
	TestFileOnly bool
}

var tamperingPatterns = []tamperingPattern{
	{ID: "T01:it-skip", Description: "it.skip / describe.skip によるテストスキップ",
		Pattern: regexp.MustCompile(`(?:it|test|describe|context)\.skip\s*\(`), TestFileOnly: true},
	{ID: "T02:xit-xdescribe", Description: "xit / xdescribe によるテスト無効化",
		Pattern: regexp.MustCompile(`\b(?:xit|xtest|xdescribe)\s*\(`), TestFileOnly: true},
	{ID: "T03:pytest-skip", Description: "pytest.mark.skip によるテストスキップ",
		Pattern: regexp.MustCompile(`@pytest\.mark\.(?:skip|xfail)\b`), TestFileOnly: true},
	{ID: "T04:go-skip", Description: "t.Skip() によるテストスキップ",
		Pattern: regexp.MustCompile(`\bt\.Skip(?:f|Now)?\s*\(`), TestFileOnly: true},
	{ID: "T05:expect-removed", Description: "expect / assert が削除された可能性（コメントアウト）",
		Pattern: regexp.MustCompile(`//\s*expect\s*\(`), TestFileOnly: true},
	{ID: "T06:assert-commented", Description: "assert 呼び出しがコメントアウトされた",
		Pattern: regexp.MustCompile(`//\s*assert(?:Equal|NotEqual|True|False|Nil|Error)?\s*\(`), TestFileOnly: true},
	{ID: "T07:todo-assert", Description: "TODO コメントによってアサーションが置き換えられた",
		Pattern: regexp.MustCompile(`(?i)//\s*TODO.*assert|//\s*TODO.*expect`), TestFileOnly: true},
	{ID: "T08:eslint-disable", Description: "eslint-disable による lint ルール無効化",
		Pattern: regexp.MustCompile(`(?m)(?://\s*eslint-disable(?:-next-line|-line)?(?:\s+[^\n]+)?$|/\*\s*eslint-disable\b[^*]*\*/)`), TestFileOnly: false},
	{ID: "T09:ci-continue-on-error", Description: "continue-on-error: true による CI 失敗無視",
		Pattern: regexp.MustCompile(`continue-on-error\s*:\s*true`), TestFileOnly: false},
	{ID: "T10:ci-if-always", Description: "if: always() による CI ステップ強制実行",
		Pattern: regexp.MustCompile(`if\s*:\s*always\s*\(\s*\)`), TestFileOnly: false},
	{ID: "T11:hardcoded-answer", Description: "テスト期待値のハードコード（辞書返し）",
		Pattern: regexp.MustCompile(`answers?_for_tests?\s*=\s*\{`), TestFileOnly: true},
	{ID: "T12:return-hardcoded", Description: "テストケース値を直接 return するパターン",
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
				Description: tp.Description,
				MatchedText: matched,
			})
		}
	}
	return warnings
}
