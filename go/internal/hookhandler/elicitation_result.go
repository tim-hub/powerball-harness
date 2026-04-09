package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// elicitationResultInput は ElicitationResult フックの stdin JSON ペイロード。
type elicitationResultInput struct {
	MCPServerName string `json:"mcp_server_name"`
	ServerName    string `json:"server_name"`
	Matcher       string `json:"matcher"`
	ElicitationID string `json:"elicitation_id"`
	ID            string `json:"id"`
	ResultStatus  string `json:"result_status"`
	Status        string `json:"status"`
}

// elicitationResultLogEntry は elicitation-events.jsonl に追記するエントリ。
type elicitationResultLogEntry struct {
	Event         string `json:"event"`
	MCPServer     string `json:"mcp_server"`
	ElicitationID string `json:"elicitation_id"`
	ResultStatus  string `json:"result_status"`
	Timestamp     string `json:"timestamp"`
}

// ElicitationResultHandler は scripts/hook-handlers/elicitation-result.sh の Go 移植。
//
// ElicitationResult イベントを受け取り、軽量ロギングのみを行う。
// 結果は .claude/state/elicitation-events.jsonl に追記される。
// 常に approve を返す。
type ElicitationResultHandler struct {
	// ProjectRoot はプロジェクトルートのパス。空の場合は環境変数/CWD から解決。
	ProjectRoot string
}

// Handle は ElicitationResult フックを処理する。
func (h *ElicitationResultHandler) Handle(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return writeJSON(out, elicitationDecision{
			Decision: "approve",
			Reason:   "ElicitationResult: no payload",
		})
	}

	var input elicitationResultInput
	if jsonErr := json.Unmarshal(data, &input); jsonErr != nil {
		return writeJSON(out, elicitationDecision{
			Decision: "approve",
			Reason:   "ElicitationResult: no payload",
		})
	}

	// フィールド正規化
	mcpServer := firstNonEmpty(input.MCPServerName, input.ServerName, input.Matcher)
	elicitationID := firstNonEmpty(input.ElicitationID, input.ID)
	resultStatus := firstNonEmpty(input.ResultStatus, input.Status)

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
		entry := elicitationResultLogEntry{
			Event:         "elicitation_result",
			MCPServer:     mcpServer,
			ElicitationID: elicitationID,
			ResultStatus:  resultStatus,
			Timestamp:     ts,
		}
		if lineData, merr := json.Marshal(entry); merr == nil {
			f, ferr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if ferr == nil {
				fmt.Fprintf(f, "%s\n", lineData)
				f.Close()
				rotateJSONL(logFile, 500, 400)
			}
		}
	}

	// 常に approve
	return writeJSON(out, elicitationDecision{
		Decision: "approve",
		Reason:   "ElicitationResult tracked",
	})
}
