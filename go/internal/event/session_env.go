package event

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SessionEnvHandler は SessionStart フックハンドラ。
// CLAUDE_ENV_FILE を活用してハーネス環境変数を設定する。
//
// shell 版: scripts/hook-handlers/session-env-setup.sh
type SessionEnvHandler struct {
	// PluginRoot はバージョンファイルを探すルートディレクトリ。
	// 空の場合は環境変数 CLAUDE_PLUGIN_ROOT から取得する。
	PluginRoot string
}

// SessionEnvVars はハーネス環境変数のセット。
type SessionEnvVars struct {
	HarnessVersion           string
	HarnessEffortDefault     string
	HarnessAgentType         string
	HarnessIsRemote          string
	HarnessBreezingSessionID string // 空の場合は書き出さない
}

// Handle は stdin から SessionStart ペイロードを読み取り、
// CLAUDE_ENV_FILE にハーネス環境変数を書き出す。
// CLAUDE_ENV_FILE が設定されていない場合は何もしない。
func (h *SessionEnvHandler) Handle(r io.Reader, _ io.Writer) error {
	// CLAUDE_ENV_FILE が設定されていない場合はスキップ
	envFile := os.Getenv("CLAUDE_ENV_FILE")
	if envFile == "" {
		return nil
	}

	// stdin は読み取るが、SessionStart では tool_name が不要
	// (エラーは無視して処理を継続)
	_, _ = io.ReadAll(r)

	vars := h.buildEnvVars()
	return h.writeEnvFile(envFile, vars)
}

// buildEnvVars は現在の環境変数から SessionEnvVars を構築する。
func (h *SessionEnvHandler) buildEnvVars() SessionEnvVars {
	pluginRoot := h.PluginRoot
	if pluginRoot == "" {
		pluginRoot = os.Getenv("CLAUDE_PLUGIN_ROOT")
	}

	version := h.readVersion(pluginRoot)

	agentType := os.Getenv("BREEZING_ROLE")
	if agentType == "" {
		agentType = "solo"
	}

	isRemote := os.Getenv("CLAUDE_CODE_REMOTE")
	if isRemote == "" {
		isRemote = "false"
	}

	return SessionEnvVars{
		HarnessVersion:           version,
		HarnessEffortDefault:     "medium",
		HarnessAgentType:         agentType,
		HarnessIsRemote:          isRemote,
		HarnessBreezingSessionID: os.Getenv("BREEZING_SESSION_ID"),
	}
}

// readVersion は VERSION ファイルからバージョン文字列を読み取る。
func (h *SessionEnvHandler) readVersion(pluginRoot string) string {
	if pluginRoot == "" {
		return "unknown"
	}

	data, err := os.ReadFile(filepath.Join(pluginRoot, "VERSION"))
	if err != nil {
		return "unknown"
	}

	v := strings.TrimSpace(string(data))
	if v == "" {
		return "unknown"
	}
	return v
}

// writeEnvFile は CLAUDE_ENV_FILE にハーネス環境変数を追記する。
func (h *SessionEnvHandler) writeEnvFile(envFile string, vars SessionEnvVars) error {
	// シンボリックリンクチェック（セキュリティ）
	if isSymlink(envFile) {
		return fmt.Errorf("security: symlinked env file refused: %s", envFile)
	}

	f, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("opening env file: %w", err)
	}
	defer f.Close()

	lines := []string{
		fmt.Sprintf("HARNESS_VERSION=%s", vars.HarnessVersion),
		fmt.Sprintf("HARNESS_EFFORT_DEFAULT=%s", vars.HarnessEffortDefault),
		fmt.Sprintf("HARNESS_AGENT_TYPE=%s", vars.HarnessAgentType),
		fmt.Sprintf("HARNESS_IS_REMOTE=%s", vars.HarnessIsRemote),
	}
	if vars.HarnessBreezingSessionID != "" {
		lines = append(lines, fmt.Sprintf("HARNESS_BREEZING_SESSION_ID=%s", vars.HarnessBreezingSessionID))
	}

	for _, line := range lines {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return fmt.Errorf("writing env file: %w", err)
		}
	}
	return nil
}
