// Package config provides parsing and validation for harness.toml,
// the single user-editable configuration file for Claude Code Harness.
//
// Supported sections:
//
//	[project]  — plugin metadata (name, version, description, author, homepage)
//	[agent]    — default CC agent
//	[env]      — environment variables injected into CC sessions
//	[safety]   — permissions and sandbox settings
//	[telemetry]— harness-internal settings (not reflected to CC files)
//
// Unsupported keys (userConfig, channels) are explicitly rejected.
package config

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

// ---------------------------------------------------------------------------
// Config structs
// ---------------------------------------------------------------------------

// TDD enforce level constants.
const (
	TDDEnforceLevelOff     = "off"
	TDDEnforceLevelCentral = "central"
	TDDEnforceLevelMax     = "max"
)

// TDDEnforceConfig maps to [tdd.enforce] in harness.toml.
type TDDEnforceConfig struct {
	Enabled                    bool   `toml:"enabled"`
	Level                      string `toml:"level"` // off, central, max
	HookEnabled                bool   `toml:"hook_enabled"`
	DefaultMaxRedLogAgeMinutes int    `toml:"default_max_red_log_age_minutes"`
	BypassAuditRequired        bool   `toml:"bypass_audit_required"`
}

// TDDConfig maps to [tdd] in harness.toml.
type TDDConfig struct {
	AdoptTodoListFirst      bool             `toml:"adopt_todo_list_first"`
	AdoptTriangulation      string           `toml:"adopt_triangulation"`
	AdoptFakeImplementation bool             `toml:"adopt_fake_implementation"`
	Enforce                 TDDEnforceConfig `toml:"enforce"`
}

// Config is the top-level harness.toml structure.
type Config struct {
	Project   ProjectConfig     `toml:"project"`
	Agent     AgentConfig       `toml:"agent"`
	Env       map[string]string `toml:"env"`
	Safety    SafetyConfig      `toml:"safety"`
	Telemetry TelemetryConfig   `toml:"telemetry"`
	TDD       TDDConfig         `toml:"tdd"`
}

// ProjectConfig maps to [project] in harness.toml.
// These fields are reflected to .claude-plugin/plugin.json.
type ProjectConfig struct {
	Name         string      `toml:"name"`
	Version      string      `toml:"version"`
	Description  string      `toml:"description"`
	Author       interface{} `toml:"author"`
	Homepage     string      `toml:"homepage"`
	Repository   string      `toml:"repository"`
	License      string      `toml:"license"`
	Keywords     []string    `toml:"keywords"`
	OutputStyles string      `toml:"outputStyles"`
}

// AuthorName returns the author name regardless of format (string or object).
func (c *ProjectConfig) AuthorName() string {
	switch v := c.Author.(type) {
	case string:
		return v
	case map[string]interface{}:
		if name, ok := v["name"].(string); ok {
			return name
		}
	}
	return ""
}

// AuthorURL returns the author URL if the author is an object form.
func (c *ProjectConfig) AuthorURL() string {
	if m, ok := c.Author.(map[string]interface{}); ok {
		if url, ok := m["url"].(string); ok {
			return url
		}
	}
	return ""
}

// AgentConfig maps to [agent] in harness.toml.
// The Default field is reflected to settings.json as the "agent" key.
type AgentConfig struct {
	Default string `toml:"default"`
}

// SafetyConfig maps to [safety] in harness.toml.
type SafetyConfig struct {
	Permissions PermissionsConfig `toml:"permissions"`
	Sandbox     SandboxConfig     `toml:"sandbox"`
}

// PermissionsConfig maps to [safety.permissions].
// Deny/Ask are reflected to settings.json. Harness-only policy fields are read
// by the Go guardrail engine at runtime.
type PermissionsConfig struct {
	Deny                []string `toml:"deny"`
	Ask                 []string `toml:"ask"`
	ProtectedBranchPush string   `toml:"protectedBranchPush"`
}

// SandboxConfig maps to [safety.sandbox].
// Reflected to settings.json as the sandbox key.
type SandboxConfig struct {
	FailIfUnavailable bool                  `toml:"failIfUnavailable"`
	Filesystem        SandboxFilesystemConfig `toml:"filesystem"`
}

// SandboxFilesystemConfig maps to [safety.sandbox.filesystem].
type SandboxFilesystemConfig struct {
	DenyRead  []string `toml:"denyRead"`
	AllowRead []string `toml:"allowRead"`
}

// TelemetryConfig maps to [telemetry] in harness.toml.
// These are harness-internal settings and are NOT reflected to CC files.
type TelemetryConfig struct {
	OtelEndpoint string `toml:"otel_endpoint"`
	WebhookURL   string `toml:"webhook_url"`
}

// ---------------------------------------------------------------------------
// Unsupported key detection
// ---------------------------------------------------------------------------

// rejectedKeys lists top-level TOML keys that harness.toml must not contain.
// These either do not exist in CC or are reserved for future incompatible use.
var rejectedKeys = []string{
	"userConfig",
	"channels",
}

// ---------------------------------------------------------------------------
// Parse
// ---------------------------------------------------------------------------

// ParseFile reads harness.toml from the given path and returns a validated Config.
// Returns an error if:
//   - the file cannot be read or parsed
//   - any unsupported key (userConfig, channels) is present
//   - [tdd].enforce.level is not one of: off, central, max
func ParseFile(path string) (*Config, error) {
	var cfg Config

	meta, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, fmt.Errorf("harness.toml: parse error: %w", err)
	}

	if err := validateKeys(meta); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ParseBytes parses harness.toml content from a byte slice.
// Useful for testing without filesystem access.
func ParseBytes(data []byte) (*Config, error) {
	var cfg Config

	meta, err := toml.Decode(string(data), &cfg)
	if err != nil {
		return nil, fmt.Errorf("harness.toml: parse error: %w", err)
	}

	if err := validateKeys(meta); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)

	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// applyDefaults fills in zero-value TDD config fields with safe defaults.
func applyDefaults(cfg *Config) {
	if cfg.TDD.Enforce.Level == "" {
		cfg.TDD.Enforce.Level = TDDEnforceLevelOff
	}
	if cfg.TDD.Enforce.DefaultMaxRedLogAgeMinutes == 0 {
		cfg.TDD.Enforce.DefaultMaxRedLogAgeMinutes = 60
	}
	if !cfg.TDD.Enforce.BypassAuditRequired {
		cfg.TDD.Enforce.BypassAuditRequired = true
	}
}

// validateConfig checks semantic constraints beyond TOML key validity.
func validateConfig(cfg *Config) error {
	level := cfg.TDD.Enforce.Level
	switch level {
	case TDDEnforceLevelOff, TDDEnforceLevelCentral, TDDEnforceLevelMax:
		// valid
	default:
		return fmt.Errorf(
			"harness.toml: tdd.enforce.level %q is invalid; must be one of: off, central, max",
			level,
		)
	}
	return nil
}

// validateKeys checks that no unsupported top-level keys are present.
// Unsupported keys are explicitly rejected so users notice misconfiguration early.
func validateKeys(meta toml.MetaData) error {
	undecoded := meta.Undecoded()
	for _, key := range undecoded {
		topLevel := key[0] // e.g., "userConfig" or "channels"
		for _, rejected := range rejectedKeys {
			if strings.EqualFold(topLevel, rejected) {
				return fmt.Errorf(
					"harness.toml: unsupported key %q — this key does not exist in Claude Code; remove it from harness.toml",
					topLevel,
				)
			}
		}
	}
	return nil
}
