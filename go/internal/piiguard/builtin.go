package piiguard

import "regexp"

// BuiltinRules is the set of built-in PII and secret detection rules.
// All patterns are compiled once at package init via regexp.MustCompile.
// Ported 1:1 from datumbrain/claude-privacy-guard src/scanner/detectors.ts (MIT).
//
// Severity mapping from upstream risk scores:
//   critical → cloud credentials, private keys, payment keys, code-hosting tokens
//   high     → session/auth tokens, platform tokens
//   medium   → PII (email)
var BuiltinRules = []Rule{
	{
		ID:       "email-address",
		Title:    "Email Address",
		Severity: SeverityMedium,
		Category: CategoryPII,
		// Upstream bug fix: [A-Z|a-z] → [A-Za-z] (| is literal in char class, not alternation).
		Pattern: regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`),
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
		Pattern:  regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9\-._~+/]+=*`),
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
