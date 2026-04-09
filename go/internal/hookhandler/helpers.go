package hookhandler

// helpers.go - hookhandler パッケージ共通ユーティリティ関数
//
// 複数のハンドラで重複していたローカル関数を1箇所に集約する。

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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
// HARNESS_PROJECT_ROOT → PROJECT_ROOT → カレントディレクトリ の優先順で解決する。
func resolveProjectRoot() string {
	if v := os.Getenv("HARNESS_PROJECT_ROOT"); v != "" {
		return v
	}
	if v := os.Getenv("PROJECT_ROOT"); v != "" {
		return v
	}
	cwd, _ := os.Getwd()
	return cwd
}
