package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// elicitationInput は Elicitation フックの stdin JSON ペイロード。
type elicitationInput struct {
	MCPServerName  string `json:"mcp_server_name"`
	ServerName     string `json:"server_name"`
	Matcher        string `json:"matcher"`
	ElicitationID  string `json:"elicitation_id"`
	ID             string `json:"id"`
	Message        string `json:"message"`
}

// elicitationLogEntry は elicitation-events.jsonl に書き込むエントリ。
type elicitationLogEntry struct {
	Event           string `json:"event"`
	MCPServer       string `json:"mcp_server"`
	ElicitationID   string `json:"elicitation_id"`
	Message         string `json:"message"`
	BreezingSession string `json:"breezing_session"`
	Timestamp       string `json:"timestamp"`
}

// elicitationDecision は Elicitation フックのレスポンス。
type elicitationDecision struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// ElicitationHandler は scripts/hook-handlers/elicitation-handler.sh の Go 移植。
//
// Elicitation イベントで MCP elicitation リクエストをログに記録し、
// Breezing Worker（バックグラウンド、UI なし）の場合は自動スキップ（deny）、
// 通常セッションはそのまま通過（allow）する。
//
// ログは .claude/state/elicitation-events.jsonl に記録される。
type ElicitationHandler struct {
	// ProjectRoot はプロジェクトルートのパス。空の場合は環境変数/CWD から解決。
	ProjectRoot string
}

// Handle は Elicitation フックを処理する。
func (h *ElicitationHandler) Handle(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return writeJSON(out, elicitationDecision{
			Decision: "approve",
			Reason:   "Elicitation: no payload",
		})
	}

	var input elicitationInput
	if jsonErr := json.Unmarshal(data, &input); jsonErr != nil {
		return writeJSON(out, elicitationDecision{
			Decision: "approve",
			Reason:   "Elicitation: no payload",
		})
	}

	// フィールド正規化（bash 版の // フォールバックと同等）
	mcpServer := firstNonEmpty(input.MCPServerName, input.ServerName, input.Matcher)
	elicitationID := firstNonEmpty(input.ElicitationID, input.ID)
	message := input.Message

	// プロジェクトルート解決
	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}
	stateDir := projectRoot + "/.claude/state"
	logFile := stateDir + "/elicitation-events.jsonl"

	// ログ記録
	if err := os.MkdirAll(stateDir, 0o700); err == nil {
		ts := time.Now().UTC().Format(time.RFC3339)
		breezingSession := os.Getenv("HARNESS_BREEZING_SESSION_ID")
		entry := elicitationLogEntry{
			Event:           "elicitation",
			MCPServer:       mcpServer,
			ElicitationID:   elicitationID,
			Message:         message,
			BreezingSession: breezingSession,
			Timestamp:       ts,
		}
		if lineData, merr := json.Marshal(entry); merr == nil {
			f, ferr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if ferr == nil {
				fmt.Fprintf(f, "%s\n", lineData)
				f.Close()
				_ = rotateJSONL(logFile, 500, 400)
			}
		}
	}

	// Breezing セッション中は自動スキップ（バックグラウンド Worker は UI 対話不能）
	breezingSession := os.Getenv("HARNESS_BREEZING_SESSION_ID")
	if breezingSession != "" {
		reason := fmt.Sprintf(
			"Breezing session (%s): background agent cannot interact with elicitation UI",
			breezingSession,
		)
		return writeJSON(out, elicitationDecision{
			Decision: "deny",
			Reason:   reason,
		})
	}

	// 通常セッション: そのまま通過（ユーザーが対話で応答）
	return writeJSON(out, elicitationDecision{
		Decision: "approve",
		Reason:   "Elicitation: forwarding to user",
	})
}
