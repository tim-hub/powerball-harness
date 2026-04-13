package hookhandler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePlansPath_Default(t *testing.T) {
	dir := t.TempDir()
	plansPath := filepath.Join(dir, "Plans.md")
	if err := os.WriteFile(plansPath, []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := resolvePlansPath(dir)
	if got != plansPath {
		t.Errorf("resolvePlansPath() = %q, want %q", got, plansPath)
	}
}

func TestResolvePlansPath_FileNotExist(t *testing.T) {
	dir := t.TempDir()

	got := resolvePlansPath(dir)
	if got != "" {
		t.Errorf("resolvePlansPath() = %q, want empty string when Plans.md not found", got)
	}
}

func TestResolvePlansPath_WithConfig(t *testing.T) {
	dir := t.TempDir()

	configContent := "plansDirectory: docs\n"
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	docsDir := filepath.Join(dir, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	plansPath := filepath.Join(docsDir, "Plans.md")
	if err := os.WriteFile(plansPath, []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := resolvePlansPath(dir)
	if got != plansPath {
		t.Errorf("resolvePlansPath() = %q, want %q", got, plansPath)
	}
}

func TestResolvePlansPath_WithConfig_FileNotExist(t *testing.T) {
	dir := t.TempDir()

	configContent := "plansDirectory: docs\n"
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}

	got := resolvePlansPath(dir)
	if got != "" {
		t.Errorf("resolvePlansPath() = %q, want empty string when Plans.md not found in custom dir", got)
	}
}

func TestResolvePlansPath_CaseVariants(t *testing.T) {
	variants := []string{"plans.md", "PLANS.md", "PLANS.MD"}
	for _, name := range variants {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			plansPath := filepath.Join(dir, name)
			if err := os.WriteFile(plansPath, []byte("# Plans\n"), 0o644); err != nil {
				t.Fatal(err)
			}

			got := resolvePlansPath(dir)
			if got == "" {
				t.Errorf("resolvePlansPath() returned empty for variant %s", name)
				return
			}
			if _, err := os.Stat(got); err != nil {
				t.Errorf("resolvePlansPath() = %q, but file does not exist: %v", got, err)
			}
		})
	}
}

func TestReadPlansDirectoryFromConfig_NormalValue(t *testing.T) {
	dir := t.TempDir()
	configContent := "plansDirectory: docs\n"
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	got := readPlansDirectoryFromConfig(dir)
	if got != "docs" {
		t.Errorf("readPlansDirectoryFromConfig() = %q, want %q", got, "docs")
	}
}

func TestReadPlansDirectoryFromConfig_QuotedValue(t *testing.T) {
	dir := t.TempDir()
	configContent := `plansDirectory: "my-plans"` + "\n"
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	got := readPlansDirectoryFromConfig(dir)
	if got != "my-plans" {
		t.Errorf("readPlansDirectoryFromConfig() = %q, want %q", got, "my-plans")
	}
}

func TestReadPlansDirectoryFromConfig_NoConfig(t *testing.T) {
	dir := t.TempDir()

	got := readPlansDirectoryFromConfig(dir)
	if got != "" {
		t.Errorf("readPlansDirectoryFromConfig() = %q, want empty when no config file", got)
	}
}

func TestReadPlansDirectoryFromConfig_AbsolutePathRejected(t *testing.T) {
	dir := t.TempDir()
	configContent := "plansDirectory: /etc/plans\n"
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	got := readPlansDirectoryFromConfig(dir)
	if got != "" {
		t.Errorf("readPlansDirectoryFromConfig() = %q, want empty for absolute path (security)", got)
	}
}

func TestReadPlansDirectoryFromConfig_ParentRefRejected(t *testing.T) {
	dir := t.TempDir()
	configContent := "plansDirectory: ../outside\n"
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	got := readPlansDirectoryFromConfig(dir)
	if got != "" {
		t.Errorf("readPlansDirectoryFromConfig() = %q, want empty for parent dir reference (security)", got)
	}
}

func TestResolveProjectRoot_EnvVarPriority(t *testing.T) {
	t.Setenv("HARNESS_PROJECT_ROOT", "/custom/root")
	got := resolveProjectRoot()
	if got != "/custom/root" {
		t.Errorf("resolveProjectRoot() = %q, want /custom/root (HARNESS_PROJECT_ROOT)", got)
	}
}

func TestResolveProjectRoot_ProjectRootFallback(t *testing.T) {
	t.Setenv("HARNESS_PROJECT_ROOT", "")
	t.Setenv("PROJECT_ROOT", "/project/root")
	got := resolveProjectRoot()
	if got != "/project/root" {
		t.Errorf("resolveProjectRoot() = %q, want /project/root (PROJECT_ROOT)", got)
	}
}

func TestResolveProjectRoot_GitFallback(t *testing.T) {
	t.Setenv("HARNESS_PROJECT_ROOT", "")
	t.Setenv("PROJECT_ROOT", "")

	got := resolveProjectRoot()
	if got == "" {
		t.Error("resolveProjectRoot() returned empty string; expected a non-empty path")
	}
	if _, err := os.Stat(got); err != nil {
		t.Errorf("resolveProjectRoot() = %q, but path does not exist: %v", got, err)
	}
	if !strings.HasPrefix(got, "/") {
		t.Errorf("resolveProjectRoot() = %q, want absolute path", got)
	}
}
