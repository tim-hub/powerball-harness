package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// stopFailureInput は StopFailure フックの stdin JSON ペイロード。
type stopFailureInput struct {
	Error     stopFailureError `json:"error"`
	SessionID string           `json:"session_id"`
}

// stopFailureError は error フィールドの構造体と文字列の両方に対応する。
type stopFailureError struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Code    string `json:"code"`
	Raw     string // error が文字列だった場合
}

// UnmarshalJSON は error フィールドが string / object 両方に対応する。
func (e *stopFailureError) UnmarshalJSON(data []byte) error {
	// まず文字列として試みる
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		e.Raw = s
		return nil
	}
	// 次に object として試みる
	type plain struct {
		Message string `json:"message"`
		Status  string `json:"status"`
		Code    string `json:"code"`
	}
	var p plain
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	e.Message = p.Message
	e.Status = p.Status
	e.Code = p.Code
	return nil
}

// stopFailureLogEntry は stop-failures.jsonl に記録するエントリ。
type stopFailureLogEntry struct {
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id"`
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

// stopFailureSystemMessage は 429 レート制限時の systemMessage レスポンス。
type stopFailureSystemMessage struct {
	SystemMessage string `json:"systemMessage"`
}

// StopFailureHandler は scripts/hook-handlers/stop-failure.sh の Go 移植。
//
// StopFailure イベント（API エラーでセッション停止が失敗した際）を処理する。
//   - エラー情報を .claude/state/stop-failures.jsonl に記録する
//   - エラー種別を分類（rate_limit, auth_error, network_error, unknown）
//   - 429 レート制限時は systemMessage で Lead に通知
type StopFailureHandler struct {
	// ProjectRoot はプロジェクトルートのパス。空の場合は環境変数/CWD から解決。
	ProjectRoot string
}

// Handle は StopFailure フックを処理する。
func (h *StopFailureHandler) Handle(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		// ペイロードなし: ログ不要、何も出力しない
		return nil
	}

	// プロジェクトルート解決
	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}
	stateDir := projectRoot + "/.claude/state"

	// ステートディレクトリを確保
	if mkErr := os.MkdirAll(stateDir, 0o700); mkErr != nil {
		// ディレクトリ作成失敗: stderr に出力して終了
		fmt.Fprintf(os.Stderr, "[StopFailure] mkdir %s: %v\n", stateDir, mkErr)
		return nil
	}

	logFile := stateDir + "/stop-failures.jsonl"

	// シンボリックリンクチェック（セキュリティ）
	if isStopFailureLogSymlink(logFile) {
		fmt.Fprintf(os.Stderr, "[StopFailure] symlink detected at %s, aborting\n", logFile)
		return nil
	}

	// JSON パース
	var input stopFailureInput
	_ = json.Unmarshal(data, &input)

	// エラー情報の正規化
	errorMsg, errorCode := normalizeStopFailureError(input.Error)
	sessionID := input.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}

	ts := time.Now().UTC().Format(time.RFC3339)

	// JSONL ログ記録
	entry := stopFailureLogEntry{
		Event:     "stop_failure",
		Timestamp: ts,
		SessionID: sessionID,
		ErrorCode: errorCode,
		Message:   errorMsg,
	}
	if lineData, merr := json.Marshal(entry); merr == nil {
		f, ferr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if ferr == nil {
			fmt.Fprintf(f, "%s\n", lineData)
			f.Close()
			rotateJSONL(logFile, 500, 400)
		}
	}

	// 429 レート制限時: systemMessage で Lead に通知
	if errorCode == "429" || errorCode == "rate_limit" {
		msg := fmt.Sprintf(
			"[StopFailure] Worker %s がレート制限 (429) で停止。Breezing Lead は指数バックオフ後に自動再開を試みてください。",
			sessionID,
		)
		if err := writeJSON(out, stopFailureSystemMessage{SystemMessage: msg}); err != nil {
			fmt.Fprintf(os.Stderr, "[StopFailure] write systemMessage: %v\n", err)
		}
	}

	// stderr にデバッグ出力
	fmt.Fprintf(os.Stderr, "[StopFailure] session=%s code=%s msg=%s\n", sessionID, errorCode, errorMsg)

	return nil
}

// normalizeStopFailureError はエラー情報を正規化し、メッセージとコードを返す。
// エラー種別の分類:
//   - "429" / メッセージに "rate" を含む → rate_limit
//   - "401", "403" / メッセージに "auth" を含む → auth_error
//   - メッセージに "network", "connection", "timeout" を含む → network_error
//   - 上記以外 → unknown
func normalizeStopFailureError(e stopFailureError) (msg, code string) {
	// Raw (文字列 error フィールド) を優先
	if e.Raw != "" {
		msg = e.Raw
		code = classifyErrorCode("", msg)
		return
	}

	msg = e.Message
	if msg == "" {
		msg = "unknown"
	}

	rawCode := firstNonEmpty(e.Status, e.Code)
	if rawCode == "" {
		rawCode = "unknown"
	}

	code = classifyErrorCode(rawCode, msg)
	return
}

// classifyErrorCode はエラーコードとメッセージからエラー種別を分類する。
func classifyErrorCode(rawCode, msg string) string {
	// HTTP ステータスコードによる分類
	switch rawCode {
	case "429":
		return "429"
	case "401", "403":
		return "auth_error"
	}

	// メッセージによる分類（コードが "unknown" の場合も含む）
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "rate") || strings.Contains(lower, "429") {
		return "rate_limit"
	}
	if strings.Contains(lower, "auth") || strings.Contains(lower, "unauthorized") || strings.Contains(lower, "forbidden") {
		return "auth_error"
	}
	if strings.Contains(lower, "network") || strings.Contains(lower, "connection") || strings.Contains(lower, "timeout") {
		return "network_error"
	}

	if rawCode != "" && rawCode != "unknown" {
		return rawCode
	}
	return "unknown"
}

// isStopFailureLogSymlink はログファイルがシンボリックリンクかどうかを返す。
// セキュリティチェック用。isSymlink は userprompt_track_command.go に定義済み。
func isStopFailureLogSymlink(path string) bool {
	return isSymlink(path)
}
