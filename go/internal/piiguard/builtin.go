package piiguard

import (
	"regexp"
	"strings"
)

// BuiltinRules is the set of built-in PII and secret detection rules.
// All patterns are compiled once at package init via regexp.MustCompile.
// Ported 1:1 from datumbrain/claude-privacy-guard src/scanner/detectors.ts (MIT).
//
// Severity mapping from upstream risk scores:
//   critical → cloud credentials, private keys, payment keys, code-hosting tokens
//   high     → session/auth tokens, platform tokens
//   medium   → PII (email)
// emailAllowlistDomains lists exact email domains (matched case-insensitively) that are
// known-safe public or test addresses and must not trigger the email-address PII rule.
// RFC 2606 / RFC 6761 reserved domains and GitHub's noreply domain are included.
// Use emailAllowlistSuffixes for TLD-level suffix patterns.
var emailAllowlistDomains = []string{
	"users.noreply.github.com", // GitHub privacy-preserving commit/PR address
	"localhost",                // bare localhost, non-routable
	"example.com",             // RFC 2606 reserved
	"example.org",             // RFC 2606 reserved
	"example.net",             // RFC 2606 reserved
}

// emailAllowlistSuffixes lists domain suffixes (matched case-insensitively via HasSuffix)
// that indicate test, local-network, or RFC-reserved addresses.
// A suffix entry ".lan" matches any domain ending in ".lan" (e.g. machine.lan).
var emailAllowlistSuffixes = []string{
	".test",       // RFC 6761 special-use TLD
	".example",    // RFC 6761 special-use TLD
	".invalid",    // RFC 6761 — guaranteed never resolves
	".localhost",  // RFC 6761 / RFC 2606 loopback TLD
	".local",      // RFC 6762 mDNS / Bonjour
	".lan",        // common local-network convention (user-requested)
	".localdomain", // common Linux default hostname suffix
	".internal",   // common internal-services convention
	".home.arpa",  // RFC 8375 home-network standard
}

// emailAllowlistValidator returns false (skip the finding) when the matched email's
// domain part is a known-safe domain or matches a known-safe suffix.
// Exact-domain check uses EqualFold; suffix check uses HasSuffix on lowercased domain.
// The exact check uses the FULL domain (everything after the last @) — not HasSuffix —
// to prevent evil-subdomain.users.noreply.github.com.attacker.example from passing.
func emailAllowlistValidator(match string) bool {
	i := strings.LastIndex(match, "@")
	if i < 0 {
		return true
	}
	// "git" is the reserved SSH service account for git hosting (GitHub, GitLab, Bitbucket…).
	// git@host is always a remote URL, never a personal email address.
	if strings.EqualFold(match[:i], "git") {
		return false
	}
	domain := match[i+1:]
	for _, d := range emailAllowlistDomains {
		if strings.EqualFold(domain, d) {
			return false
		}
	}
	lower := strings.ToLower(domain)
	for _, s := range emailAllowlistSuffixes {
		if strings.HasSuffix(lower, s) {
			return false
		}
	}
	return true
}

var BuiltinRules = []Rule{
	{
		ID:       "email-address",
		Title:    "Email Address",
		Severity: SeverityMedium,
		Category: CategoryPII,
		// Upstream bug fix: [A-Z|a-z] → [A-Za-z] (| is literal in char class, not alternation).
		Pattern:   regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`),
		Validator: emailAllowlistValidator,
	},
	{
		ID:       "jwt-token",
		Title:    "JWT Token",
		Severity: SeverityHigh,
		Category: CategorySecret,
		Pattern:  regexp.MustCompile(`\beyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`),
	},
	{
		ID:       "bearer-token",
		Title:    "Bearer Token",
		Severity: SeverityHigh,
		Category: CategorySecret,
		Pattern:   regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9\-._~+/]{20,}=*`),
		Validator: bearerTokenValidator,
	},
	{
		ID:       "aws-api-key",
		Title:    "AWS API Key",
		Severity: SeverityCritical,
		Category: CategorySecret,
		// Covers all AWS key prefixes: AKIA (user), A3T/AGPA/AIDA/AROA/AIPA/ANPA/ANVA (role), ASIA (STS).
		Pattern: regexp.MustCompile(`\b(AKIA|A3T|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}\b`),
	},
	{
		ID:       "openai-api-key",
		Title:    "OpenAI API Key",
		Severity: SeverityCritical,
		Category: CategorySecret,
		// Real keys: sk-<48 chars> (legacy) or sk-proj-<48+ chars> (project scoped).
		// Minimum 40 chars after the prefix avoids matching sk-learn, sk-ant-*, etc.
		Pattern: regexp.MustCompile(`\bsk-(?:proj-[A-Za-z0-9\-_]{40,}|[A-Za-z0-9]{40,})\b`),
	},
	{
		ID:       "anthropic-api-key",
		Title:    "Anthropic API Key",
		Severity: SeverityCritical,
		Category: CategorySecret,
		Pattern:  regexp.MustCompile(`\bsk-ant-[A-Za-z0-9\-_]{20,}\b`),
	},
	{
		ID:       "openrouter-api-key",
		Title:    "OpenRouter API Key",
		Severity: SeverityCritical,
		Category: CategorySecret,
		Pattern:  regexp.MustCompile(`\bsk-or-v1-[A-Za-z0-9]{20,}\b`),
	},
	{
		ID:       "google-ai-api-key",
		Title:    "Google AI API Key",
		Severity: SeverityCritical,
		Category: CategorySecret,
		Pattern:  regexp.MustCompile(`\bAIza[0-9A-Za-z\-_]{35}\b`),
	},
	{
		ID:       "groq-api-key",
		Title:    "Groq API Key",
		Severity: SeverityCritical,
		Category: CategorySecret,
		Pattern:  regexp.MustCompile(`\bgsk_[A-Za-z0-9]{20,}\b`),
	},
	{
		ID:       "perplexity-api-key",
		Title:    "Perplexity API Key",
		Severity: SeverityCritical,
		Category: CategorySecret,
		Pattern:  regexp.MustCompile(`\bpplx-[A-Za-z0-9]{20,}\b`),
	},
	{
		ID:       "huggingface-api-token",
		Title:    "Hugging Face API Token",
		Severity: SeverityHigh,
		Category: CategorySecret,
		Pattern:  regexp.MustCompile(`\bhf_[A-Za-z0-9]{30,}\b`),
	},
	{
		ID:       "stripe-api-key",
		Title:    "Stripe API Key",
		Severity: SeverityCritical,
		Category: CategorySecret,
		Pattern:  regexp.MustCompile(`\bsk_(live|test)_[0-9a-zA-Z]{24,}\b`),
	},
	{
		ID:       "github-token",
		Title:    "GitHub Token",
		Severity: SeverityCritical,
		Category: CategorySecret,
		// Covers classic PATs (ghp_), OAuth (gho_), user (ghu_), app (ghs_, ghr_), and fine-grained PATs.
		Pattern: regexp.MustCompile(`\b(?:ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9]{20,255}\b|\bgithub_pat_[A-Za-z0-9_]{20,255}\b`),
	},
	{
		// genericSecretPattern contains a backtick in its quote character class,
		// so this rule uses a double-quoted string literal instead of a raw literal.
		// The character class ['"` + "`" + `] matches single-quote, double-quote, or backtick.
		ID:       "generic-code-secret-assignment",
		Title:    "Generic Code Secret Assignment",
		Severity: SeverityHigh,
		Category: CategorySecret,
		Pattern:  regexp.MustCompile("(?i)\\b(?:api[_-]?key|secret|token|access[_-]?token|auth[_-]?token)\\b\\s*[:=]\\s*['\"`][A-Za-z0-9_\\-\\/+=]{16,}['\"`]"),
	},
	{
		ID:       "private-key",
		Title:    "Private Key",
		Severity: SeverityCritical,
		Category: CategorySecret,
		// Uses [\s\S]*? (lazy, matches newlines) because PEM blocks span multiple lines.
		// Go RE2 supports lazy quantifiers and [\s\S] for any-char-including-newline.
		Pattern: regexp.MustCompile(`-----BEGIN (?:RSA |OPENSSH |EC )?PRIVATE KEY-----[\s\S]*?-----END (?:RSA |OPENSSH |EC )?PRIVATE KEY-----`),
	},
}

// bearerTokenValidator rejects matches where the captured token value is entirely
// lowercase-ASCII alpha — plain English words, not credentials.
// Real tokens contain at least one digit, uppercase letter, or non-alpha character.
func bearerTokenValidator(match string) bool {
	i := strings.IndexAny(match, " \t")
	if i < 0 {
		return true
	}
	token := strings.TrimLeft(match[i:], " \t")
	token = strings.TrimRight(token, "=")
	if len(token) == 0 {
		return true
	}
	for _, c := range token {
		if c < 'a' || c > 'z' {
			return true
		}
	}
	return false
}
