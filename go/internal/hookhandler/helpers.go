package hookhandler

// helpers.go - hookhandler パッケージ共通ユーティリティ関数
//
// 複数のハンドラで重複していたローカル関数を1箇所に集約する。

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

// fileExists はファイルが存在するかを確認する。
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isSymlink はパスがシンボリックリンクかどうかを返す（存在しない場合は false）。
func isSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// rotateJSONL は JSONL ファイルが maxLines を超えた場合に keepLines 行に切り詰める。
// ファイルが存在しない場合は nil を返す（エラーなし）。
// シンボリックリンクへの書き込みは拒否してエラーを返す。
func rotateJSONL(path string, maxLines, keepLines int) error {
	if isSymlink(path) || isSymlink(path+".tmp") {
		return fmt.Errorf("symlinked file refused for rotation")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil // ファイルが存在しない場合は無視
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) <= maxLines {
		return nil
	}

	// 末尾 keepLines 行を残す
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

// firstNonEmpty は引数の中で最初の空でない文字列を返す。
// いずれも空の場合は "" を返す。
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// writeJSON は任意の値を JSON として w に書き込む。
func writeJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

// resolveProjectRoot はプロジェクトルートディレクトリを返す。
//
// 解決優先順:
//  1. HARNESS_PROJECT_ROOT 環境変数
//  2. PROJECT_ROOT 環境変数
//  3. git rev-parse --show-toplevel（monorepo の subdir 対応）
//  4. カレントディレクトリ（フォールバック）
//
// bash 版 path-utils.sh / config-utils.sh の detect_project_root() に相当。
func resolveProjectRoot() string {
	if v := os.Getenv("HARNESS_PROJECT_ROOT"); v != "" {
		return v
	}
	if v := os.Getenv("PROJECT_ROOT"); v != "" {
		return v
	}
	// git rev-parse --show-toplevel でリポジトリルートを検出する。
	// monorepo のサブディレクトリで実行された場合でも .claude/ が見つかるよう対応。
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

// harnessConfigFileName は設定ファイルのデフォルト名。
const harnessConfigFileName = ".claude-code-harness.config.yaml"

// readPlansDirectoryFromConfig は projectRoot 配下の設定ファイルから
// plansDirectory の値を返す。設定がない・読めない場合は空文字を返す。
//
// YAML パーサーをインポートしないため、以下の順でフォールバックする:
//  1. "plansDirectory: <value>" 形式の行を bufio.Scanner でスキャン
//
// セキュリティ: 以下の値は安全のためデフォルト（空文字）にフォールバックする:
//   - 絶対パス（/ 始まり）
//   - 親ディレクトリ参照（.. を含む）
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
		// "plansDirectory:" で始まる行を探す
		const key = "plansDirectory:"
		if !strings.HasPrefix(line, key) {
			continue
		}
		value := strings.TrimSpace(line[len(key):])
		// クォートを除去（シングル・ダブル）
		value = strings.Trim(value, `"'`)
		value = strings.TrimSpace(value)

		if value == "" {
			return ""
		}
		// セキュリティ: 絶対パスを拒否
		if filepath.IsAbs(value) {
			return ""
		}
		// セキュリティ: 親ディレクトリ参照を拒否
		if strings.Contains(value, "..") {
			return ""
		}
		return value
	}
	return ""
}

// resolvePlansPath は projectRoot 配下で Plans.md のフルパスを返す。
//
// 解決ロジック:
//  1. 設定ファイル (.claude-code-harness.config.yaml) の plansDirectory を読む
//  2. 設定があれば filepath.Join(projectRoot, plansDirectory, "Plans.md") を返す
//  3. なければ filepath.Join(projectRoot, "Plans.md") を返す
//  4. ファイルが存在しない場合は空文字を返す
//
// bash 版の get_plans_file_path() に相当する。
func resolvePlansPath(projectRoot string) string {
	// 設定から plansDirectory を取得
	plansDir := readPlansDirectoryFromConfig(projectRoot)

	// 候補ファイル名（bash 版と同じ大文字小文字バリエーション）
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

	// 存在しない場合は空文字を返す（bash 版の plans_file_exists() 相当）
	return ""
}
