package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tim-hub/powerball-harness/go/pkg/config"
)

// outputDir is the plugin subfolder that receives generated files.
const outputDir = "harness"

// runSync implements the "harness sync" subcommand.
//
// It reads harness/harness.toml from the project root, then generates:
//   - harness/settings.json ← [agent] + [env] + [safety.permissions] + [safety.sandbox]
//
// harness/hooks/hooks.json is now the canonical hooks location (no longer
// synced from a separate source; it is edited directly in harness/hooks/).
//
// The project root is determined by the first argument (or cwd if omitted).
// Exit 0 on success, exit 1 on any error.
func runSync(args []string) {
	// Determine project root
	projectRoot, err := resolveProjectRoot(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "harness sync: %v\n", err)
		os.Exit(1)
	}

	// Parse harness/harness.toml
	tomlPath := filepath.Join(projectRoot, outputDir, "harness.toml")
	cfg, err := config.ParseFile(tomlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "harness sync: %v\n", err)
		os.Exit(1)
	}

	if err := generateSettingsJSON(projectRoot, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "harness sync: settings.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("harness sync: done")
}

// ---------------------------------------------------------------------------
// resolveProjectRoot
// ---------------------------------------------------------------------------

// resolveProjectRoot returns the project root directory.
// If args contains one element it is treated as the root; otherwise the
// current working directory is used.
func resolveProjectRoot(args []string) (string, error) {
	if len(args) > 0 {
		abs, err := filepath.Abs(args[0])
		if err != nil {
			return "", fmt.Errorf("invalid project root %q: %w", args[0], err)
		}
		return abs, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine working directory: %w", err)
	}
	return cwd, nil
}

// ---------------------------------------------------------------------------
// settings.json
// ---------------------------------------------------------------------------

// settingsJSON mirrors the schema of harness/settings.json.
// Only non-empty / non-nil fields are included in the output so that
// a minimal harness.toml produces a minimal settings.json.
type settingsJSON struct {
	Schema      string            `json:"$schema,omitempty"`
	Agent       string            `json:"agent,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	Permissions *permissionsField `json:"permissions,omitempty"`
	Sandbox     *sandboxField     `json:"sandbox,omitempty"`
}

type permissionsField struct {
	Deny []string `json:"deny,omitempty"`
	Ask  []string `json:"ask,omitempty"`
}

type sandboxField struct {
	FailIfUnavailable bool                  `json:"failIfUnavailable"`
	Filesystem        *sandboxFilesystemField `json:"filesystem,omitempty"`
}

type sandboxFilesystemField struct {
	DenyRead  []string `json:"denyRead,omitempty"`
	AllowRead []string `json:"allowRead,omitempty"`
}

func generateSettingsJSON(projectRoot string, cfg *config.Config) error {
	s := settingsJSON{
		Schema: "https://json.schemastore.org/claude-code-settings.json",
	}

	// [agent]
	if cfg.Agent.Default != "" {
		s.Agent = cfg.Agent.Default
	}

	// [env]
	if len(cfg.Env) > 0 {
		s.Env = cfg.Env
	}

	// [safety.permissions]
	p := &permissionsField{
		Deny: cfg.Safety.Permissions.Deny,
		Ask:  cfg.Safety.Permissions.Ask,
	}
	if len(p.Deny) > 0 || len(p.Ask) > 0 {
		s.Permissions = p
	}

	// [safety.sandbox]
	sb := cfg.Safety.Sandbox
	if sb.FailIfUnavailable || len(sb.Filesystem.DenyRead) > 0 || len(sb.Filesystem.AllowRead) > 0 {
		sf := &sandboxField{
			FailIfUnavailable: sb.FailIfUnavailable,
		}
		if len(sb.Filesystem.DenyRead) > 0 || len(sb.Filesystem.AllowRead) > 0 {
			sf.Filesystem = &sandboxFilesystemField{
				DenyRead:  sb.Filesystem.DenyRead,
				AllowRead: sb.Filesystem.AllowRead,
			}
		}
		s.Sandbox = sf
	}

	data, err := marshalPretty(s)
	if err != nil {
		return err
	}

	dest := filepath.Join(projectRoot, outputDir, "settings.json")
	if err := writeFile(dest, data); err != nil {
		return err
	}

	fmt.Printf("  wrote %s\n", rel(projectRoot, dest))
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// marshalPretty marshals v to indented JSON with a trailing newline.
func marshalPretty(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("JSON marshal: %w", err)
	}
	return buf.Bytes(), nil
}

// writeFile writes data to path, creating parent directories as needed.
// It refuses to write to paths outside the OS temp directory or the file's
// own parent (safety guard against path-traversal in tests).
func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// rel returns path relative to base, falling back to the absolute path.
func rel(base, path string) string {
	r, err := filepath.Rel(base, path)
	if err != nil {
		return path
	}
	return r
}
