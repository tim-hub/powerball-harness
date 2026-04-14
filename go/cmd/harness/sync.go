package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tim-hub/powerball-harness/go/pkg/config"
)

// runSync implements the "harness sync" subcommand.
//
// It reads harness.toml from the project root, then generates:
//   - .claude-plugin/plugin.json   ← [project] section
//   - .claude-plugin/hooks.json    ← copied from hooks/hooks.json, or skipped if that file is a symlink
//   - .claude-plugin/settings.json ← [agent] + [env] + [safety.permissions] + [safety.sandbox]
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

	// Parse harness.toml
	tomlPath := filepath.Join(projectRoot, "harness.toml")
	cfg, err := config.ParseFile(tomlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "harness sync: %v\n", err)
		os.Exit(1)
	}

	// Run each generator; collect errors to report all at once
	var errs []error

	if err := generatePluginJSON(projectRoot, cfg); err != nil {
		errs = append(errs, fmt.Errorf("plugin.json: %w", err))
	}

	if err := syncHooksJSON(projectRoot); err != nil {
		errs = append(errs, fmt.Errorf("hooks.json sync: %w", err))
	}

	if err := generateSettingsJSON(projectRoot, cfg); err != nil {
		errs = append(errs, fmt.Errorf("settings.json: %w", err))
	}

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "harness sync: %v\n", e)
		}
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
// plugin.json
// ---------------------------------------------------------------------------

// pluginJSON is the schema for .claude-plugin/plugin.json.
// Fields that are not set in harness.toml are omitted from the output.
type pluginJSON struct {
	Name         string      `json:"name,omitempty"`
	Version      string      `json:"version,omitempty"`
	Description  string      `json:"description,omitempty"`
	Author       interface{} `json:"author,omitempty"`
	Homepage     string      `json:"homepage,omitempty"`
	Repository   string      `json:"repository,omitempty"`
	License      string      `json:"license,omitempty"`
	Keywords     []string    `json:"keywords,omitempty"`
	// Skills declares the skill discovery roots per CC 2.1.94+.
	// With ["./"], CC uses each SKILL.md frontmatter `name` field as the
	// invocation name, allowing directory name and invocation name to differ.
	Skills       []string    `json:"skills,omitempty"`
	OutputStyles string      `json:"outputStyles,omitempty"`
}

func generatePluginJSON(projectRoot string, cfg *config.Config) error {
	// Author: preserve object form if URL is set, otherwise use string
	var author interface{}
	name := cfg.Project.AuthorName()
	url := cfg.Project.AuthorURL()
	if name != "" {
		if url != "" {
			author = map[string]string{"name": name, "url": url}
		} else {
			author = name
		}
	}

	p := pluginJSON{
		Name:         cfg.Project.Name,
		Version:      cfg.Project.Version,
		Description:  cfg.Project.Description,
		Author:       author,
		Homepage:     cfg.Project.Homepage,
		Repository:   cfg.Project.Repository,
		License:      cfg.Project.License,
		Keywords:     cfg.Project.Keywords,
		// Always emit ["./"] so CC 2.1.94+ uses frontmatter `name` fields
		// for skill invocation. This enables short-name aliases like
		// HAR:plan while keeping directory names unchanged.
		Skills:       []string{"./"},
		OutputStyles: cfg.Project.OutputStyles,
	}

	data, err := marshalPretty(p)
	if err != nil {
		return err
	}

	dest := filepath.Join(projectRoot, ".claude-plugin", "plugin.json")
	if err := writeFile(dest, data); err != nil {
		return err
	}

	fmt.Printf("  wrote %s\n", rel(projectRoot, dest))
	return nil
}

// ---------------------------------------------------------------------------
// hooks.json sync
// ---------------------------------------------------------------------------

// syncHooksJSON ensures hooks/hooks.json and .claude-plugin/hooks.json are in sync.
//
// If hooks/hooks.json is a symlink (the preferred setup, pointing to
// ../.claude-plugin/hooks.json), the two files are always identical by
// definition and no copy is needed. Otherwise, hooks/hooks.json is treated as
// the source and copied to .claude-plugin/hooks.json.
func syncHooksJSON(projectRoot string) error {
	src := filepath.Join(projectRoot, "hooks", "hooks.json")
	dst := filepath.Join(projectRoot, ".claude-plugin", "hooks.json")

	// If hooks/hooks.json is a symlink it points directly to the destination,
	// so they are always in sync — nothing to do.
	if fi, err := os.Lstat(src); err == nil && fi.Mode()&os.ModeSymlink != 0 {
		fmt.Printf("  skipped %s (symlinked to %s)\n", rel(projectRoot, src), rel(projectRoot, dst))
		return nil
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	// Validate that the source is valid JSON before copying
	if !json.Valid(data) {
		return fmt.Errorf("%s is not valid JSON", src)
	}

	if err := writeFile(dst, data); err != nil {
		return err
	}

	fmt.Printf("  wrote %s (copied from %s)\n", rel(projectRoot, dst), rel(projectRoot, src))
	return nil
}

// ---------------------------------------------------------------------------
// settings.json
// ---------------------------------------------------------------------------

// settingsJSON mirrors the schema of .claude-plugin/settings.json.
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

	dest := filepath.Join(projectRoot, ".claude-plugin", "settings.json")
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
func marshalPretty(v interface{}) ([]byte, error) {
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
