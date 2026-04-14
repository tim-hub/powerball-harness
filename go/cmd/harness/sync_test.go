package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// setupProjectDir creates a temporary project root with:
//   - harness.toml (provided content)
//   - hooks/hooks.json (minimal valid JSON, enough to exercise the copy path)
func setupProjectDir(t *testing.T, tomlContent string) string {
	t.Helper()

	dir := t.TempDir()

	// Write harness.toml
	if err := os.WriteFile(filepath.Join(dir, "harness.toml"), []byte(tomlContent), 0o644); err != nil {
		t.Fatalf("write harness.toml: %v", err)
	}

	// Write hooks/hooks.json — minimal but valid JSON
	hooksDir := filepath.Join(dir, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	minimalHooks := `{"description":"test hooks","hooks":{"PreToolUse":[]}}`
	if err := os.WriteFile(filepath.Join(hooksDir, "hooks.json"), []byte(minimalHooks), 0o644); err != nil {
		t.Fatalf("write hooks/hooks.json: %v", err)
	}

	return dir
}

// readJSON reads and unmarshals a JSON file into a map.
func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var v map[string]any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}

	return v
}

// ---------------------------------------------------------------------------
// Full sync: all sections set
// ---------------------------------------------------------------------------

var fullTOML = `
[project]
name = "claude-code-harness"
version = "3.17.0"
description = "Claude harness"
author = "tim-hub"
homepage = "https://github.com/tim-hub/powerball-harness"

[agent]
default = "security-reviewer"

[env]
CLAUDE_CODE_SUBPROCESS_ENV_SCRUB = "1"

[safety.permissions]
deny = [
  "Bash(sudo:*)",
  "Bash(rm -rf:*)",
  "mcp__codex__*",
  "Read(./.env)",
]
ask = [
  "Bash(rm -r:*)",
  "Bash(git push -f:*)",
]

[safety.sandbox]
failIfUnavailable = true

[safety.sandbox.filesystem]
denyRead = [".env", "secrets/**", "**/*.pem"]
allowRead = [".env.example", "docs/**"]

[telemetry]
otel_endpoint = ""
webhook_url = ""
`

func TestSync_GeneratesSettingsJSON(t *testing.T) {
	dir := setupProjectDir(t, fullTOML)
	runSync([]string{dir})

	v := readJSON(t, filepath.Join(dir, "harness", "settings.json"))

	// $schema
	if v["$schema"] != "https://json.schemastore.org/claude-code-settings.json" {
		t.Errorf("settings.json $schema = %v", v["$schema"])
	}

	// agent
	if v["agent"] != "security-reviewer" {
		t.Errorf("settings.json agent = %v, want security-reviewer", v["agent"])
	}

	// env
	envRaw, ok := v["env"].(map[string]any)
	if !ok {
		t.Fatalf("settings.json env is not an object: %T", v["env"])
	}
	if envRaw["CLAUDE_CODE_SUBPROCESS_ENV_SCRUB"] != "1" {
		t.Errorf("settings.json env.CLAUDE_CODE_SUBPROCESS_ENV_SCRUB = %v, want 1", envRaw["CLAUDE_CODE_SUBPROCESS_ENV_SCRUB"])
	}

	// permissions
	permRaw, ok := v["permissions"].(map[string]any)
	if !ok {
		t.Fatalf("settings.json permissions is not an object: %T", v["permissions"])
	}
	denyRaw, ok := permRaw["deny"].([]any)
	if !ok {
		t.Fatalf("settings.json permissions.deny is not an array")
	}
	if len(denyRaw) != 4 {
		t.Errorf("permissions.deny len = %d, want 4", len(denyRaw))
	}
	if denyRaw[0] != "Bash(sudo:*)" {
		t.Errorf("permissions.deny[0] = %v, want Bash(sudo:*)", denyRaw[0])
	}

	askRaw, ok := permRaw["ask"].([]any)
	if !ok {
		t.Fatalf("settings.json permissions.ask is not an array")
	}
	if len(askRaw) != 2 {
		t.Errorf("permissions.ask len = %d, want 2", len(askRaw))
	}

	// sandbox
	sbRaw, ok := v["sandbox"].(map[string]any)
	if !ok {
		t.Fatalf("settings.json sandbox is not an object: %T", v["sandbox"])
	}
	if sbRaw["failIfUnavailable"] != true {
		t.Errorf("sandbox.failIfUnavailable = %v, want true", sbRaw["failIfUnavailable"])
	}
	fsRaw, ok := sbRaw["filesystem"].(map[string]any)
	if !ok {
		t.Fatalf("sandbox.filesystem is not an object")
	}
	denyReadRaw, ok := fsRaw["denyRead"].([]any)
	if !ok {
		t.Fatalf("sandbox.filesystem.denyRead is not an array")
	}
	if len(denyReadRaw) != 3 {
		t.Errorf("sandbox.filesystem.denyRead len = %d, want 3", len(denyReadRaw))
	}
}

func TestSync_CopiesHooksJSON(t *testing.T) {
	dir := setupProjectDir(t, fullTOML)
	runSync([]string{dir})

	// Source and generated file must exist and have identical content
	srcData, err := os.ReadFile(filepath.Join(dir, "hooks", "hooks.json"))
	if err != nil {
		t.Fatalf("read hooks/hooks.json: %v", err)
	}
	dstData, err := os.ReadFile(filepath.Join(dir, "harness", "hooks", "hooks.json"))
	if err != nil {
		t.Fatalf("read harness/hooks/hooks.json: %v", err)
	}

	if string(srcData) != string(dstData) {
		t.Errorf("hooks.json files differ:\nsrc: %s\ndst: %s", srcData, dstData)
	}
}

// ---------------------------------------------------------------------------
// Telemetry must NOT appear in settings.json
// ---------------------------------------------------------------------------

func TestSync_TelemetryNotInSettings(t *testing.T) {
	dir := setupProjectDir(t, fullTOML)
	runSync([]string{dir})

	v := readJSON(t, filepath.Join(dir, "harness", "settings.json"))

	if _, ok := v["telemetry"]; ok {
		t.Error("settings.json must not contain telemetry key")
	}
	if _, ok := v["otel_endpoint"]; ok {
		t.Error("settings.json must not contain otel_endpoint key")
	}
	if _, ok := v["webhook_url"]; ok {
		t.Error("settings.json must not contain webhook_url key")
	}
}

// ---------------------------------------------------------------------------
// Minimal TOML: only [project].name — most keys should be absent
// ---------------------------------------------------------------------------

func TestSync_MinimalTOML(t *testing.T) {
	dir := setupProjectDir(t, `
[project]
name = "minimal"
`)
	runSync([]string{dir})

	sv := readJSON(t, filepath.Join(dir, "harness", "settings.json"))
	// agent must be absent
	if _, ok := sv["agent"]; ok {
		t.Error("settings.json must not have agent when not set")
	}
	// env must be absent
	if _, ok := sv["env"]; ok {
		t.Error("settings.json must not have env when not set")
	}
	// permissions must be absent
	if _, ok := sv["permissions"]; ok {
		t.Error("settings.json must not have permissions when not set")
	}
	// sandbox must be absent
	if _, ok := sv["sandbox"]; ok {
		t.Error("settings.json must not have sandbox when not set")
	}
}

// ---------------------------------------------------------------------------
// Missing harness.toml should produce error (exit via os.Exit — tested indirectly)
// ---------------------------------------------------------------------------

func TestSync_ResolveProjectRoot_CurrentDir(t *testing.T) {
	root, err := resolveProjectRoot(nil)
	if err != nil {
		t.Fatalf("resolveProjectRoot with nil args: %v", err)
	}
	if root == "" {
		t.Error("expected non-empty project root from cwd")
	}
}

func TestSync_ResolveProjectRoot_ExplicitPath(t *testing.T) {
	dir := t.TempDir()
	root, err := resolveProjectRoot([]string{dir})
	if err != nil {
		t.Fatalf("resolveProjectRoot with explicit path: %v", err)
	}
	if root != dir {
		t.Errorf("root = %q, want %q", root, dir)
	}
}

// ---------------------------------------------------------------------------
// sandbox with failIfUnavailable=false and no filesystem — omit sandbox key
// ---------------------------------------------------------------------------

func TestSync_SandboxFalse_NoFilesystem_Omitted(t *testing.T) {
	dir := setupProjectDir(t, `
[project]
name = "test"

[safety.sandbox]
failIfUnavailable = false
`)
	runSync([]string{dir})

	sv := readJSON(t, filepath.Join(dir, "harness", "settings.json"))
	if _, ok := sv["sandbox"]; ok {
		t.Error("settings.json should not have sandbox when failIfUnavailable=false and no filesystem rules")
	}
}

// ---------------------------------------------------------------------------
// sandbox with failIfUnavailable=true — sandbox key present even without filesystem
// ---------------------------------------------------------------------------

func TestSync_SandboxTrue_NoFilesystem(t *testing.T) {
	dir := setupProjectDir(t, `
[project]
name = "test"

[safety.sandbox]
failIfUnavailable = true
`)
	runSync([]string{dir})

	sv := readJSON(t, filepath.Join(dir, "harness", "settings.json"))
	sbRaw, ok := sv["sandbox"].(map[string]any)
	if !ok {
		t.Fatalf("settings.json sandbox should be present when failIfUnavailable=true")
	}
	if sbRaw["failIfUnavailable"] != true {
		t.Errorf("sandbox.failIfUnavailable = %v, want true", sbRaw["failIfUnavailable"])
	}
	if _, ok := sbRaw["filesystem"]; ok {
		t.Error("sandbox.filesystem should not appear when no filesystem rules are set")
	}
}
