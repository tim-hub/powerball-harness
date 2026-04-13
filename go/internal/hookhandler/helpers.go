package hookhandler

// helpers.go - common utility functions for the hookhandler package.
//
// Consolidates local functions that were duplicated across multiple handlers.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// fileExists reports whether the file at path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isSymlink reports whether path is a symbolic link (returns false if the path does not exist).
func isSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// rotateJSONL truncates a JSONL file to keepLines when it exceeds maxLines.
// Returns nil (no error) when the file does not exist.
// Refuses to write to symbolic links and returns an error.
func rotateJSONL(path string, maxLines, keepLines int) error {
	if isSymlink(path) || isSymlink(path+".tmp") {
		return fmt.Errorf("symlinked file refused for rotation")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil // File does not exist — ignore.
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) <= maxLines {
		return nil
	}

	// Keep the last keepLines lines.
	start := len(lines) - keepLines
	if start < 0 {
		start = 0
	}
	trimmed := strings.Join(lines[start:], "\n") + "\n"

	tmpPath := path + ".tmp"
	if writeErr := os.WriteFile(tmpPath, []byte(trimmed), 0o644); writeErr != nil {
		return fmt.Errorf("write tmp file: %w", writeErr)
	}
	return os.Rename(tmpPath, path)
}

// firstNonEmpty returns the first non-empty string from the given values.
// Returns "" when all values are empty.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// writeJSON serializes an arbitrary value as JSON and writes it to w.
func writeJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

// resolveProjectRoot returns the project root directory.
//
// Resolution order:
//  1. HARNESS_PROJECT_ROOT environment variable
//  2. PROJECT_ROOT environment variable
//  3. git rev-parse --show-toplevel (supports monorepo subdirectories)
//  4. Current working directory (fallback)
//
// Equivalent to detect_project_root() in the bash versions of path-utils.sh / config-utils.sh.
func resolveProjectRoot() string {
	if v := os.Getenv("HARNESS_PROJECT_ROOT"); v != "" {
		return v
	}
	if v := os.Getenv("PROJECT_ROOT"); v != "" {
		return v
	}
	// Detect the repository root via git rev-parse --show-toplevel.
	// Ensures .claude/ is found even when running inside a monorepo subdirectory.
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err == nil {
		if root := strings.TrimSpace(stdout.String()); root != "" {
			return root
		}
	}
	cwd, _ := os.Getwd()
	return cwd
}

// harnessConfigFileName is the default name for the configuration file.
const harnessConfigFileName = ".claude-code-harness.config.yaml"

// readPlansDirectoryFromConfig returns the plansDirectory value from the configuration file
// under projectRoot. Returns an empty string when the setting is absent or unreadable.
//
// To avoid importing a YAML parser, falls back to scanning for lines of the form
// "plansDirectory: <value>" using bufio.Scanner.
//
// Security: the following values are rejected and fall back to the default (empty string):
//   - Absolute paths (starting with /)
//   - Parent directory references (containing ..)
func readPlansDirectoryFromConfig(projectRoot string) string {
	configPath := filepath.Join(projectRoot, harnessConfigFileName)
	f, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Look for a line starting with "plansDirectory:".
		const key = "plansDirectory:"
		if !strings.HasPrefix(line, key) {
			continue
		}
		value := strings.TrimSpace(line[len(key):])
		// Strip surrounding quotes (single or double).
		value = strings.Trim(value, `"'`)
		value = strings.TrimSpace(value)

		if value == "" {
			return ""
		}
		// Security: reject absolute paths.
		if filepath.IsAbs(value) {
			return ""
		}
		// Security: reject parent directory references.
		if strings.Contains(value, "..") {
			return ""
		}
		return value
	}
	return ""
}

// resolvePlansPath returns the full path to Plans.md under projectRoot.
//
// Resolution logic:
//  1. Read plansDirectory from the config file (.claude-code-harness.config.yaml)
//  2. If set, return filepath.Join(projectRoot, plansDirectory, "Plans.md")
//  3. Otherwise return filepath.Join(projectRoot, "Plans.md")
//  4. Returns an empty string when the file does not exist
//
// Equivalent to get_plans_file_path() in the bash version.
func resolvePlansPath(projectRoot string) string {
	// Read plansDirectory from config.
	plansDir := readPlansDirectoryFromConfig(projectRoot)

	candidates := []string{"Plans.md", "plans.md", "PLANS.md", "PLANS.MD"}

	var baseDir string
	if plansDir != "" {
		baseDir = filepath.Join(projectRoot, plansDir)
	} else {
		baseDir = projectRoot
	}

	for _, name := range candidates {
		full := filepath.Join(baseDir, name)
		if _, err := os.Stat(full); err == nil {
			return full
		}
	}

	// File not found — return empty string (equivalent to plans_file_exists() in the bash version).
	return ""
}
