package piiguard

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// defaultExternalCatalog is the embedded external regex catalog,
// ported from datumbrain/claude-privacy-guard/data/regex_list_1.json (MIT).
// See data/SOURCE.md for attribution and update instructions.
//
//go:embed data/pii-regex.json
var defaultExternalCatalog []byte

// codingSecretKeywords mirrors upstream CODING_SECRET_KEYWORDS — used to keep only
// patterns relevant to source code contexts when codingOnly=true. Drops prose-only
// patterns like driver license formats and routing numbers.
var codingSecretKeywords = []string{
	"api key",
	"apikey",
	"access key",
	"token",
	"secret",
	"password",
	"passwd",
	"private key",
	"credential",
	"bearer",
	"jwt",
	"oauth",
	"auth",
	"ssh",
	"pgp",
}

// disabledExternalRuleIDs lists external catalog rules that produce too many false positives
// to be useful in an agentic coding context.  Rules are identified by their slugified ID.
var disabledExternalRuleIDs = map[string]bool{
	// Matches plain AWS field-name keywords in docs/comments, not actual credentials.
	"external-aws-credentials-context": true,
	// Matches OAuth doc URLs and prose, not actual secrets.
	"external-microsoft-office-365-oauth-context": true,
	// Matches the tool name in any security-discussion text.
	"external-john-the-ripper": true,
}

// externalEntry mirrors the upstream regex_list_1.json entry shape.
type externalEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Regex       string `json:"regex"`
	Risk        int    `json:"risk"`
	Category    string `json:"category"`
}

// LoadExternalCatalog returns the embedded external catalog as compiled rules.
// codingOnly=true filters to coding-relevant patterns (matches upstream default).
//
// Patterns that fail to compile under Go's RE2 engine (e.g. PCRE lookaheads
// or backreferences) are silently skipped with a warning written to stderr.
func LoadExternalCatalog(codingOnly bool) []Rule {
	rules, _ := loadExternalFromBytes(defaultExternalCatalog, codingOnly, os.Stderr)
	return rules
}

// LoadExternal reads an external catalog JSON from the given filesystem path
// and returns compiled rules. Returns nil + nil if the file does not exist
// (matches upstream loadExternalRulesFromJson semantics).
func LoadExternal(path string, codingOnly bool) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("piiguard: read external catalog %q: %w", path, err)
	}
	return loadExternalFromBytes(data, codingOnly, os.Stderr)
}

// loadExternalFromBytes is the shared implementation; warnLog is exposed
// as a parameter so tests can capture warnings.
func loadExternalFromBytes(data []byte, codingOnly bool, warnLog io.Writer) ([]Rule, error) {
	if len(data) == 0 {
		return nil, nil
	}

	var entries []externalEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("piiguard: unmarshal external catalog: %w", err)
	}

	rules := make([]Rule, 0, len(entries))
	usedIDs := make(map[string]bool, len(entries))

	for _, entry := range entries {
		if entry.Name == "" || entry.Regex == "" {
			continue
		}
		if codingOnly && !isCodingSecretPattern(entry) {
			continue
		}

		slug := "external-" + slugify(entry.Name)
		if disabledExternalRuleIDs[slug] {
			fmt.Fprintf(warnLog, "piiguard: skipping disabled rule %q (high false-positive rate in agentic context)\n", slug)
			continue
		}

		pattern, err := regexp.Compile(entry.Regex)
		if err != nil {
			fmt.Fprintf(warnLog, "piiguard: skipping %q (regex compile failed: %v)\n", entry.Name, err)
			continue
		}

		id := uniqueID(usedIDs, slug)
		usedIDs[id] = true

		rules = append(rules, Rule{
			ID:       id,
			Title:    entry.Name,
			Severity: riskToSeverity(entry.Risk),
			Category: sourceCategoryToCategory(entry.Category),
			Pattern:  pattern,
		})
	}

	return rules, nil
}

// riskToSeverity maps an upstream risk score (1-10) to a Severity.
// Mirrors upstream riskToSeverity in detectors.ts.
func riskToSeverity(risk int) Severity {
	switch {
	case risk >= 8:
		return SeverityCritical
	case risk >= 6:
		return SeverityHigh
	case risk >= 3:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// sourceCategoryToCategory maps an upstream JSON category string to our Category.
// Only "pii" (case-insensitive) maps to CategoryPII; all other values map to CategorySecret.
func sourceCategoryToCategory(sourceCategory string) Category {
	if strings.EqualFold(strings.TrimSpace(sourceCategory), "pii") {
		return CategoryPII
	}
	return CategorySecret
}

var (
	slugifyNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
	slugifyTrim     = regexp.MustCompile(`(?:^-+)|(?:-+$)`)
)

// slugify converts a name to a URL-safe slug, capped at 80 chars.
// Mirrors upstream slugify in detectors.ts.
func slugify(value string) string {
	s := strings.ToLower(value)
	s = slugifyNonAlnum.ReplaceAllString(s, "-")
	s = slugifyTrim.ReplaceAllString(s, "")
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}

// isCodingSecretPattern returns true if the entry mentions any coding-secret keyword
// in its name, description, or pattern. Search is case-insensitive.
func isCodingSecretPattern(entry externalEntry) bool {
	text := strings.ToLower(entry.Name + " " + entry.Description + " " + entry.Regex)
	for _, kw := range codingSecretKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

// uniqueID disambiguates duplicate slugs by appending a numeric suffix.
func uniqueID(used map[string]bool, base string) string {
	if !used[base] {
		return base
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !used[candidate] {
			return candidate
		}
	}
}
