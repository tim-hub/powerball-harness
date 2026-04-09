// Package hookhandler implements Go ports of the Harness hook handler scripts.
package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// autoBroadcastPatterns は自動ブロードキャスト対象のパターン一覧。
// session-auto-broadcast.sh の AUTO_BROADCAST_PATTERNS に対応。
var autoBroadcastPatterns = []string{
	"src/api/",
	"src/types/",
	"src/interfaces/",
	"api/",
	"types/",
	"schema.prisma",
	"openapi",
	"swagger",
	".graphql",
}

// autoBroadcastInput は session-auto-broadcast.sh に渡される stdin JSON。
type autoBroadcastInput struct {
	ToolInput struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
	} `json:"tool_input"`
}

// autoBroadcastConfig は .claude/sessions/auto-broadcast.json の設定。
type autoBroadcastConfig struct {
	Enabled  *bool    `json:"enabled"`
	Patterns []string `json:"patterns"`
}

// postToolOutput は PostToolUse フックのレスポンス形式。
type postToolOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// emptyPostToolOutput は追加コンテキストなしの PostToolUse レスポンスを返す。
func emptyPostToolOutput(w io.Writer) error {
	out := postToolOutput{}
	out.HookSpecificOutput.HookEventName = "PostToolUse"
	out.HookSpecificOutput.AdditionalContext = ""
	return writeJSON(w, out)
}

// HandleSessionAutoBroadcast は session-auto-broadcast.sh の Go 移植。
//
// PostToolUse Write/Edit イベントで呼び出され、重要なファイルの変更を
// .claude/state/broadcast.md にチームメイト通知として書き込む。
//
// 対象パターン: src/api/, src/types/, src/interfaces/, api/, types/,
// schema.prisma, openapi, swagger, .graphql
func HandleSessionAutoBroadcast(in io.Reader, out io.Writer) error {
	// stdin から JSON を読み取る
	data, err := io.ReadAll(in)
	if err != nil {
		return emptyPostToolOutput(out)
	}

	// 入力がない場合は空レスポンスを返す
	if len(strings.TrimSpace(string(data))) == 0 {
		return emptyPostToolOutput(out)
	}

	var input autoBroadcastInput
	if err := json.Unmarshal(data, &input); err != nil {
		return emptyPostToolOutput(out)
	}

	// file_path または path を取得
	filePath := input.ToolInput.FilePath
	if filePath == "" {
		filePath = input.ToolInput.Path
	}

	// ファイルパスがない場合は終了
	if filePath == "" {
		return emptyPostToolOutput(out)
	}

	// 設定ファイルを読み込む
	configFile := ".claude/sessions/auto-broadcast.json"
	enabled := true
	var customPatterns []string

	if cfgData, cfgErr := os.ReadFile(configFile); cfgErr == nil {
		var cfg autoBroadcastConfig
		if jsonErr := json.Unmarshal(cfgData, &cfg); jsonErr == nil {
			if cfg.Enabled != nil {
				enabled = *cfg.Enabled
			}
			customPatterns = cfg.Patterns
		}
	}

	// 自動ブロードキャストが無効な場合は終了
	if !enabled {
		return emptyPostToolOutput(out)
	}

	// パターンマッチング（組み込みパターン）
	matchedPattern := ""
	for _, pattern := range autoBroadcastPatterns {
		if strings.Contains(filePath, pattern) {
			matchedPattern = pattern
			break
		}
	}

	// カスタムパターンもチェック
	if matchedPattern == "" {
		for _, pattern := range customPatterns {
			if pattern != "" && strings.Contains(filePath, pattern) {
				matchedPattern = pattern
				break
			}
		}
	}

	// マッチしない場合は空レスポンスを返す
	if matchedPattern == "" {
		return emptyPostToolOutput(out)
	}

	// ブロードキャスト実行: .claude/state/broadcast.md に書き込む
	fileName := filepath.Base(filePath)
	if broadcastErr := writeBroadcastNotification(filePath, matchedPattern); broadcastErr != nil {
		// 書き込み失敗は無視（フォールバックとして空レスポンスを返す）
		return emptyPostToolOutput(out)
	}

	// 通知メッセージを出力
	o := postToolOutput{}
	o.HookSpecificOutput.HookEventName = "PostToolUse"
	o.HookSpecificOutput.AdditionalContext = fmt.Sprintf(
		"自動ブロードキャスト: %s の変更を他セッションに通知しました", fileName,
	)
	return writeJSON(out, o)
}

// writeBroadcastNotification は .claude/state/broadcast.md にチームメイト通知を書き込む。
func writeBroadcastNotification(filePath, matchedPattern string) error {
	stateDir := ".claude/state"
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return fmt.Errorf("mkdir state dir: %w", err)
	}

	broadcastFile := filepath.Join(stateDir, "broadcast.md")

	ts := time.Now().UTC().Format(time.RFC3339)
	entry := fmt.Sprintf("## %s\n\n- file: %s\n- pattern: %s\n- action: auto-broadcast\n\n",
		ts, filePath, matchedPattern)

	f, err := os.OpenFile(broadcastFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open broadcast file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("write broadcast entry: %w", err)
	}
	return nil
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
