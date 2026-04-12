package hookhandler

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// stopSessionInput は Stop フックの stdin JSON ペイロード。
// CC 2.1.47+ で last_assistant_message が含まれる。
type stopSessionInput struct {
	StopHookActive       bool   `json:"stop_hook_active"`
	TranscriptPath       string `json:"transcript_path"`
	LastAssistantMessage string `json:"last_assistant_message"`
}

// stopSessionResponse は Stop フックのレスポンス。
type stopSessionResponse struct {
	OK            bool   `json:"ok"`
	Reason        string `json:"reason,omitempty"`
	SystemMessage string `json:"systemMessage,omitempty"`
}

// StopSessionEvaluatorHandler は scripts/hook-handlers/stop-session-evaluator.sh の Go 移植。
//
// Stop イベントでセッション状態を評価する。
//   - last_assistant_message を長さ・ハッシュ（SHA-256 先頭 16 文字）にして session.json に記録
//   - WIP タスクがある場合は systemMessage で警告（ブロックはしない）
//   - 停止は常に許可（ok: true）
type StopSessionEvaluatorHandler struct {
	// ProjectRoot はプロジェクトルートのパス。空の場合は環境変数/CWD から解決。
	ProjectRoot string
}

// Handle は Stop フックを処理する。
func (h *StopSessionEvaluatorHandler) Handle(in io.Reader, out io.Writer) error {
	// プロジェクトルート解決
	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}
	stateFile := projectRoot + "/.claude/state/session.json"

	// stdin を読み取る（サイズ上限: 64 KiB）
	var payload []byte
	limited := io.LimitReader(in, 65536)
	payload, _ = io.ReadAll(limited)

	// last_assistant_message のメタデータを session.json に記録
	if len(payload) > 0 {
		var input stopSessionInput
		if jsonErr := json.Unmarshal(payload, &input); jsonErr == nil {
			if input.LastAssistantMessage != "" {
				h.recordLastMessage(stateFile, input.LastAssistantMessage)
			}
		}
	}

	// session.json が存在しない場合はデフォルト ok
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return writeJSON(out, stopSessionResponse{OK: true})
	}

	// セッション状態を読み取り
	sessionData, err := os.ReadFile(stateFile)
	if err != nil {
		return writeJSON(out, stopSessionResponse{OK: true})
	}

	var sessionMap map[string]interface{}
	if jsonErr := json.Unmarshal(sessionData, &sessionMap); jsonErr != nil {
		return writeJSON(out, stopSessionResponse{OK: true})
	}

	// 既に stopped 状態なら即 ok
	if state, ok := sessionMap["state"].(string); ok && state == "stopped" {
		return writeJSON(out, stopSessionResponse{OK: true})
	}

	// WIP タスクチェック: Plans.md を探して cc:WIP を数える
	wipCount := h.countWIPTasks(projectRoot)
	if wipCount > 0 {
		msg := fmt.Sprintf(
			"[StopSession] %d WIP タスクが残っています。Plans.md を確認してください。",
			wipCount,
		)
		return writeJSON(out, stopSessionResponse{
			OK:            true,
			SystemMessage: msg,
		})
	}

	return writeJSON(out, stopSessionResponse{OK: true})
}

// recordLastMessage は session.json に last_message_length と last_message_hash を記録する。
// 平文内容は保存しない（プライバシー保護）。
func (h *StopSessionEvaluatorHandler) recordLastMessage(stateFile, msg string) {
	// ファイルが存在しない場合はスキップ（bash 版と同じ動作）
	sessionData, err := os.ReadFile(stateFile)
	if err != nil {
		return
	}

	var sessionMap map[string]interface{}
	if jsonErr := json.Unmarshal(sessionData, &sessionMap); jsonErr != nil {
		return
	}

	msgLen := len(msg)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(msg)))[:16]

	sessionMap["last_message_length"] = msgLen
	sessionMap["last_message_hash"] = hash

	newData, err := json.Marshal(sessionMap)
	if err != nil {
		return
	}

	// アトミック書き込み: 一時ファイル + rename
	stateDir := stateFile[:strings.LastIndex(stateFile, "/")]
	tmpFile, err := os.CreateTemp(stateDir, "session.json.*")
	if err != nil {
		return
	}
	tmpPath := tmpFile.Name()
	defer func() {
		// rename 失敗時のクリーンアップ
		os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(append(newData, '\n')); err != nil {
		tmpFile.Close()
		return
	}
	tmpFile.Close()

	_ = os.Rename(tmpPath, stateFile)
}

// countWIPTasks は projectRoot 配下の Plans.md を探し、cc:WIP マーカーの数を返す。
func (h *StopSessionEvaluatorHandler) countWIPTasks(projectRoot string) int {
	for _, name := range plansFileNames {
		path := projectRoot + "/" + name
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		count := 0
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "cc:WIP") {
				count++
			}
		}
		f.Close()
		return count
	}
	return 0
}
