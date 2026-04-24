package hookhandler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestResolvePlansPath_Default は設定なし・Plans.md 存在でデフォルトパスを返すことを確認する。
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

// TestResolvePlansPath_FileNotExist は Plans.md が存在しない場合に空文字を返すことを確認する。
func TestResolvePlansPath_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	// Plans.md を作成しない

	got := resolvePlansPath(dir)
	if got != "" {
		t.Errorf("resolvePlansPath() = %q, want empty string when Plans.md not found", got)
	}
}

// TestResolvePlansPath_WithConfig は plansDirectory 設定があるとき
// サブディレクトリの Plans.md パスを返すことを確認する。
func TestResolvePlansPath_WithConfig(t *testing.T) {
	dir := t.TempDir()

	// 設定ファイルを作成
	configContent := "plansDirectory: docs\n"
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte(configContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// docs/Plans.md を作成
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

// TestResolvePlansPath_WithConfig_FileNotExist は設定あり・ファイル非存在で空文字を返すことを確認する。
func TestResolvePlansPath_WithConfig_FileNotExist(t *testing.T) {
	dir := t.TempDir()

	// 設定ファイルを作成（docs ディレクトリは作るが Plans.md は作らない）
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

// TestResolvePlansPath_CaseVariants は大文字小文字のバリエーションを検出することを確認する。
// macOS の APFS は大文字小文字を区別しないため、検出されたパスが存在するファイルを
// 指していることを確認する（正確なファイル名ではなく存在確認）。
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
			// 返ってきたパスが存在していること（大文字小文字不区別のFSでも動作確認）
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

// TestReadPlansDirectoryFromConfig_NormalValue は通常の値が正しく読めることを確認する。
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

// TestReadPlansDirectoryFromConfig_QuotedValue はクォートされた値が正しく読めることを確認する。
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

// TestReadPlansDirectoryFromConfig_NoConfig は設定ファイルなしで空文字を返すことを確認する。
func TestReadPlansDirectoryFromConfig_NoConfig(t *testing.T) {
	dir := t.TempDir()

	got := readPlansDirectoryFromConfig(dir)
	if got != "" {
		t.Errorf("readPlansDirectoryFromConfig() = %q, want empty when no config file", got)
	}
}

// TestReadPlansDirectoryFromConfig_AbsolutePathRejected は絶対パスがセキュリティ上拒否されることを確認する。
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

// TestReadPlansDirectoryFromConfig_ParentRefRejected は .. を含む値がセキュリティ上拒否されることを確認する。
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

// TestResolveProjectRoot_EnvVarPriority は環境変数が最優先であることを確認する。
func TestResolveProjectRoot_EnvVarPriority(t *testing.T) {
	t.Setenv("HARNESS_PROJECT_ROOT", "/custom/root")
	got := resolveProjectRoot()
	if got != "/custom/root" {
		t.Errorf("resolveProjectRoot() = %q, want /custom/root (HARNESS_PROJECT_ROOT)", got)
	}
}

// TestResolveProjectRoot_ProjectRootFallback は PROJECT_ROOT のフォールバックを確認する。
func TestResolveProjectRoot_ProjectRootFallback(t *testing.T) {
	t.Setenv("HARNESS_PROJECT_ROOT", "")
	t.Setenv("PROJECT_ROOT", "/project/root")
	got := resolveProjectRoot()
	if got != "/project/root" {
		t.Errorf("resolveProjectRoot() = %q, want /project/root (PROJECT_ROOT)", got)
	}
}

// TestResolveProjectRoot_GitFallback は git rev-parse の結果が使われることを確認する。
// HARNESS_PROJECT_ROOT と PROJECT_ROOT が未設定の場合、git toplevel が使用される。
// このテストは git リポジトリ内で実行される前提（CI 環境も含む）。
func TestResolveProjectRoot_GitFallback(t *testing.T) {
	t.Setenv("HARNESS_PROJECT_ROOT", "")
	t.Setenv("PROJECT_ROOT", "")

	got := resolveProjectRoot()
	// git リポジトリ内であれば空文字にならないこと
	if got == "" {
		t.Error("resolveProjectRoot() returned empty string; expected a non-empty path")
	}
	// 返ってきたパスが存在すること
	if _, err := os.Stat(got); err != nil {
		t.Errorf("resolveProjectRoot() = %q, but path does not exist: %v", got, err)
	}
	// git toplevel またはカレントディレクトリのどちらかであること（スラッシュで始まる）
	if !strings.HasPrefix(got, "/") {
		t.Errorf("resolveProjectRoot() = %q, want absolute path", got)
	}
}

func TestResolveHarnessLocale_DefaultEnNoConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_CODE_HARNESS_LANG", "")

	got := resolveHarnessLocale(dir)
	if got != "en" {
		t.Errorf("resolveHarnessLocale() = %q, want en", got)
	}
}

func TestResolveHarnessLocale_EnvJapanese(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_CODE_HARNESS_LANG", "ja")

	got := resolveHarnessLocale(dir)
	if got != "ja" {
		t.Errorf("resolveHarnessLocale() = %q, want ja", got)
	}
}

func TestResolveHarnessLocale_ConfigPriorityOverEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_CODE_HARNESS_LANG", "en")
	config := "i18n:\n  language: ja\n"
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	got := resolveHarnessLocale(dir)
	if got != "ja" {
		t.Errorf("resolveHarnessLocale() = %q, want ja from config", got)
	}
}

func TestResolveHarnessLocale_InvalidValuesNormalizeToEn(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_CODE_HARNESS_LANG", "ja")
	config := "i18n:\n  language: fr\n"
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	got := resolveHarnessLocale(dir)
	if got != "en" {
		t.Errorf("resolveHarnessLocale() = %q, want invalid config normalized to en", got)
	}
}

func TestResolveHarnessLocale_ExplicitPriority(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_CODE_HARNESS_LANG", "ja")
	config := "i18n:\n  language: ja\n"
	if err := os.WriteFile(filepath.Join(dir, harnessConfigFileName), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}

	got := resolveHarnessLocale(dir, "en")
	if got != "en" {
		t.Errorf("resolveHarnessLocale(explicit en) = %q, want en", got)
	}
}
