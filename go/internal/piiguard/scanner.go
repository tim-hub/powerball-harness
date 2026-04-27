package piiguard

// Severity weights for risk score aggregation.  Mirrors upstream
// PrivacyScanner risk weights from src/scanner/engine.ts.
const (
	weightCritical = 100
	weightHigh     = 50
	weightMedium   = 20
	weightLow      = 10
	maxRiskScore   = 100

	redactionMaxRunes = 50
	redactionEllipsis = "…"
)

// Finding represents a single match of a Rule against scanned text.
type Finding struct {
	RuleID     string   `json:"rule_id"`
	Title      string   `json:"title"`
	Severity   Severity `json:"severity"`
	Category   Category `json:"category"`
	Match      string   `json:"match"`
	StartIndex int      `json:"start_index"`
	EndIndex   int      `json:"end_index"`
	// RedactedValue is Match truncated to 50 runes (+ ellipsis if longer).
	// Designed for safe display in user-visible output without leaking the secret.
	RedactedValue string `json:"redacted_value"`
}

// ScanResult is the aggregated output of (*Scanner).Scan.
type ScanResult struct {
	Findings  []Finding      `json:"findings"`
	RiskScore int            `json:"risk_score"` // 0-100, capped at 100
	Summary   map[string]int `json:"summary"`    // category → count
}

// Scanner runs a set of compiled Rules against arbitrary text.
// Safe for concurrent use because the underlying *regexp.Regexp instances
// are themselves goroutine-safe.
type Scanner struct {
	rules []Rule
}

// NewScanner constructs a Scanner from the given rules.
// Passing nil produces a scanner that returns empty results for every input.
func NewScanner(rules []Rule) *Scanner {
	return &Scanner{rules: rules}
}

// Rules returns the underlying rule set (read-only view).
func (s *Scanner) Rules() []Rule {
	return s.rules
}

// Scan applies every rule to text and returns the aggregated findings,
// weighted risk score (capped at 100), and a per-category count summary.
func (s *Scanner) Scan(text string) ScanResult {
	result := ScanResult{Summary: map[string]int{}}
	if text == "" || len(s.rules) == 0 {
		return result
	}

	for _, r := range s.rules {
		if r.Pattern == nil {
			continue
		}
		for _, idx := range r.Pattern.FindAllStringIndex(text, -1) {
			start, end := idx[0], idx[1]
			match := text[start:end]
			result.Findings = append(result.Findings, Finding{
				RuleID:        r.ID,
				Title:         r.Title,
				Severity:      r.Severity,
				Category:      r.Category,
				Match:         match,
				StartIndex:    start,
				EndIndex:      end,
				RedactedValue: redactMatch(match),
			})
			result.Summary[string(r.Category)]++
			result.RiskScore += severityWeight(r.Severity)
		}
	}

	if result.RiskScore > maxRiskScore {
		result.RiskScore = maxRiskScore
	}
	return result
}

// severityWeight returns the risk-score weight for a severity level.
func severityWeight(s Severity) int {
	switch s {
	case SeverityCritical:
		return weightCritical
	case SeverityHigh:
		return weightHigh
	case SeverityMedium:
		return weightMedium
	case SeverityLow:
		return weightLow
	default:
		return 0
	}
}

// redactMatch truncates a match to redactionMaxRunes runes and appends the
// ellipsis if the value was actually shortened.  Rune-aware so multi-byte
// characters are not split mid-codepoint.
func redactMatch(match string) string {
	runes := []rune(match)
	if len(runes) <= redactionMaxRunes {
		return match
	}
	return string(runes[:redactionMaxRunes]) + redactionEllipsis
}
