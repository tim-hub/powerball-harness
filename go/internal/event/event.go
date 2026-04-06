// Package event は worker-runtime 側のフックハンドラを実装する。
//
// 各ハンドラは stdin から CC フック JSON を受け取り、
// 必要な処理を行い、stdout に結果を返す。
// shell スクリプト (hook-handlers/*.sh) と同じ I/O プロトコルを維持する。
//
// SPEC.md §12 パッケージ境界: これらは worker-runtime 側のハンドラであり、
// internal/guard/ (hook-fastpath) とは分離する。
package event

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ---------------------------------------------------------------------------
// 共通型
// ---------------------------------------------------------------------------

// Input は CC フックから stdin 経由で受け取る JSON ペイロード。
// tool_name のような必須フィールドなし（guard パッケージとは異なる）。
type Input struct {
	SessionID    string `json:"session_id,omitempty"`
	HookEvent    string `json:"hook_event_name,omitempty"`
	ToolName     string `json:"tool_name,omitempty"`
	AgentID      string `json:"agent_id,omitempty"`
	AgentType    string `json:"agent_type,omitempty"`
	Error        string `json:"error,omitempty"`
	Message      string `json:"message,omitempty"`
	CWD          string `json:"cwd,omitempty"`
	PluginRoot   string `json:"plugin_root,omitempty"`

	// PermissionDenied 用
	Tool         string `json:"tool,omitempty"`
	DeniedReason string `json:"denied_reason,omitempty"`
	Reason       string `json:"reason,omitempty"`

	// Notification 用
	NotificationType string `json:"notification_type,omitempty"`
	Type             string `json:"type,omitempty"`
	Matcher          string `json:"matcher,omitempty"`
}

// ApproveResponse は基本的な approve レスポンス。
type ApproveResponse struct {
	Decision          string `json:"decision"`
	Reason            string `json:"reason,omitempty"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

// SystemMessageResponse は systemMessage を含むレスポンス。
type SystemMessageResponse struct {
	SystemMessage string `json:"systemMessage,omitempty"`
}

// RetryResponse は retry フラグを含むレスポンス。
type RetryResponse struct {
	Retry         bool   `json:"retry"`
	SystemMessage string `json:"systemMessage,omitempty"`
}

// ---------------------------------------------------------------------------
// 共通ユーティリティ
// ---------------------------------------------------------------------------

// ReadInput は r から JSON を読み取り Input に変換する。
// 空の入力はエラーを返す。
func ReadInput(r io.Reader) (Input, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Input{}, fmt.Errorf("reading stdin: %w", err)
	}
	if len(data) == 0 {
		return Input{}, fmt.Errorf("empty input")
	}

	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		return Input{}, fmt.Errorf("parsing JSON: %w", err)
	}
	return input, nil
}

// WriteJSON は v を JSON として w に書き出す（末尾に改行付き）。
func WriteJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

// Now は現在時刻を ISO 8601 UTC 形式で返す。
func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ResolveStateDir は PROJECT_ROOT からステートディレクトリのパスを返す。
// CLAUDE_PLUGIN_DATA が設定されている場合はプロジェクト別ハッシュでスコープする。
func ResolveStateDir(projectRoot string) string {
	pluginData := os.Getenv("CLAUDE_PLUGIN_DATA")
	if pluginData != "" {
		// プロジェクトルートのハッシュ（先頭12文字）でスコープ
		h := simpleHash(projectRoot)
		return filepath.Join(pluginData, "projects", h)
	}
	return filepath.Join(projectRoot, ".claude", "state")
}

// simpleHash はプロジェクトルートパスから 12 文字の単純なハッシュを生成する。
// shasum に依存しないため pure Go で実装する。
func simpleHash(s string) string {
	// FNV-like ハッシュ（セキュリティ不要、識別子として使用）
	var h uint64 = 14695981039346656037
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return fmt.Sprintf("%012x", h)
}

// EnsureStateDir はステートディレクトリを作成する。
// シンボリックリンクの場合はエラーを返す（セキュリティ対策）。
func EnsureStateDir(stateDir string) error {
	parent := filepath.Dir(stateDir)

	// シンボリックリンクチェック
	if isSymlink(parent) || isSymlink(stateDir) {
		return fmt.Errorf("security: symlinked state path refused: %s", stateDir)
	}

	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return fmt.Errorf("creating state dir: %w", err)
	}
	return nil
}

// isSymlink はパスがシンボリックリンクかどうかを返す。
func isSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// RotateJSONL は JSONL ファイルが 500 行を超えた場合に 400 行に切り詰める。
func RotateJSONL(path string) {
	if isSymlink(path) || isSymlink(path+".tmp") {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	// 行数カウント
	lines := splitLines(data)
	if len(lines) <= 500 {
		return
	}

	// 末尾 400 行を保持
	kept := lines[len(lines)-400:]
	content := joinLines(kept)
	_ = os.WriteFile(path+".tmp", []byte(content), 0600)
	_ = os.Rename(path+".tmp", path)
}

// splitLines は改行で分割し、空行は除外する。
func splitLines(data []byte) []string {
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			line := string(data[start:i])
			if line != "" {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(data) {
		line := string(data[start:])
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// joinLines は行のスライスを改行で結合する（末尾に改行付き）。
func joinLines(lines []string) string {
	result := ""
	for _, l := range lines {
		result += l + "\n"
	}
	return result
}
