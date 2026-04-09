package hookhandler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BreezingSignalInjectorHandler は UserPromptSubmit フックハンドラ（breezing シグナル注入）。
// breezing-signals.jsonl から未消費シグナルを読み取り、systemMessage として注入する。
// Breezing 非アクティブ時はスキップする。
//
// shell 版: scripts/hook-handlers/breezing-signal-injector.sh
type BreezingSignalInjectorHandler struct {
	// ProjectRoot はプロジェクトルートのパス。空の場合は cwd を使用する。
	ProjectRoot string
}

// breezingSignal は breezing-signals.jsonl の1行を表す。
type breezingSignal struct {
	Signal         string  `json:"signal"`
	Type           string  `json:"type"`
	Timestamp      string  `json:"timestamp"`
	ConsumedAt     *string `json:"consumed_at"`
	Conclusion     string  `json:"conclusion"`
	TriggerCommand string  `json:"trigger_command"`
	Reason         string  `json:"reason"`
	TaskID         string  `json:"task_id"`
}

// injectorResponse は BreezingSignalInjector フックのレスポンス。
type injectorResponse struct {
	SystemMessage string `json:"systemMessage,omitempty"`
}

// Handle は stdin からペイロードを読み取り（使用しない）、
// breezing シグナルを systemMessage として注入する。
func (h *BreezingSignalInjectorHandler) Handle(r io.Reader, w io.Writer) error {
	// stdin は読み捨て（このハンドラは入力を使用しない）
	_, _ = io.ReadAll(r)

	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	stateDir := filepath.Join(projectRoot, ".claude", "state")
	activeFile := filepath.Join(stateDir, "breezing-active.json")
	signalsFile := filepath.Join(stateDir, "breezing-signals.jsonl")

	// breezing セッションが存在するかチェック
	if _, err := os.Stat(activeFile); os.IsNotExist(err) {
		// breezing セッション外はスキップ（出力なし = exit 0 相当）
		return nil
	}

	// シグナルファイルが存在するかチェック
	if _, err := os.Stat(signalsFile); os.IsNotExist(err) {
		return nil
	}

	// 未消費シグナルを読み取る
	unconsumedSignals, err := h.readUnconsumedSignals(signalsFile)
	if err != nil || len(unconsumedSignals) == 0 {
		return nil
	}

	// シグナルをメッセージ形式に整形
	var messageParts []string
	for _, sig := range unconsumedSignals {
		msg := h.formatSignalMessage(sig)
		if msg != "" {
			messageParts = append(messageParts, msg)
		}
	}

	if len(messageParts) == 0 {
		return nil
	}

	// consumed_at を設定してシグナルをマーク済みにする
	_ = h.markSignalsConsumed(signalsFile)

	header := fmt.Sprintf("[breezing-signal-injector] %d 件の未消費シグナルがあります:\n", len(unconsumedSignals))
	fullMessage := header + strings.Join(messageParts, "")

	resp := injectorResponse{SystemMessage: fullMessage}
	return writeInjectorJSON(w, resp)
}

// readUnconsumedSignals は JSONL ファイルから consumed_at が null のシグナルを返す。
func (h *BreezingSignalInjectorHandler) readUnconsumedSignals(signalsFile string) ([]breezingSignal, error) {
	f, err := os.Open(signalsFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var result []breezingSignal
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var sig breezingSignal
		if err := json.Unmarshal([]byte(line), &sig); err != nil {
			continue
		}

		if sig.ConsumedAt == nil {
			result = append(result, sig)
		}
	}
	return result, scanner.Err()
}

// formatSignalMessage はシグナルをメッセージ文字列に変換する。
func (h *BreezingSignalInjectorHandler) formatSignalMessage(sig breezingSignal) string {
	signalType := sig.Signal
	if signalType == "" {
		signalType = sig.Type
	}
	if signalType == "" {
		signalType = "unknown"
	}

	switch signalType {
	case "ci_failure_detected":
		conclusion := sig.Conclusion
		if conclusion == "" {
			conclusion = "unknown"
		}
		triggerCmd := sig.TriggerCommand
		return fmt.Sprintf(
			"[SIGNAL:ci_failure_detected] CI が失敗しました（%s）。トリガー: %s。ci-cd-fixer エージェントで自動修復することを検討してください。\n",
			conclusion, triggerCmd,
		)
	case "retake_requested":
		return fmt.Sprintf(
			"[SIGNAL:retake_requested] タスク #%s のやり直しが要求されました。理由: %s\n",
			sig.TaskID, sig.Reason,
		)
	case "reviewer_approved":
		return fmt.Sprintf(
			"[SIGNAL:reviewer_approved] タスク #%s がレビュアーに承認されました。\n",
			sig.TaskID,
		)
	case "escalation_required":
		return fmt.Sprintf(
			"[SIGNAL:escalation_required] タスク #%s でエスカレーションが必要です。理由: %s\n",
			sig.TaskID, sig.Reason,
		)
	default:
		raw, _ := json.Marshal(sig)
		return fmt.Sprintf("[SIGNAL:%s] %s\n", signalType, string(raw))
	}
}

// markSignalsConsumed は signalsFile 内の未消費シグナルに consumed_at を付与して上書きする。
// ロックはディレクトリ作成によるアトミック操作で実現する。
func (h *BreezingSignalInjectorHandler) markSignalsConsumed(signalsFile string) error {
	stateDir := filepath.Dir(signalsFile)
	lockDir := filepath.Join(stateDir, ".breezing-signals.lock")

	// ロック取得（最大 2 秒、100ms ポーリング）
	const maxRetries = 20
	acquired := false
	for i := 0; i < maxRetries; i++ {
		if err := os.Mkdir(lockDir, 0700); err == nil {
			acquired = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !acquired {
		return fmt.Errorf("could not acquire lock")
	}
	defer func() { _ = os.Remove(lockDir) }()

	// ファイルを読み込み、consumed_at を付与して書き直す
	f, err := os.Open(signalsFile)
	if err != nil {
		return err
	}

	consumedTS := time.Now().UTC().Format(time.RFC3339)
	var newLines bytes.Buffer
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var sig map[string]interface{}
		if err := json.Unmarshal([]byte(line), &sig); err != nil {
			newLines.WriteString(line + "\n")
			continue
		}

		// consumed_at が null のものにタイムスタンプを付与
		if sig["consumed_at"] == nil {
			sig["consumed_at"] = consumedTS
		}

		updated, err := json.Marshal(sig)
		if err != nil {
			newLines.WriteString(line + "\n")
			continue
		}
		newLines.Write(updated)
		newLines.WriteByte('\n')
	}
	f.Close()

	if err := scanner.Err(); err != nil {
		return err
	}

	return os.WriteFile(signalsFile, newLines.Bytes(), 0600)
}

// writeInjectorJSON は v を JSON として w に書き出す。
func writeInjectorJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
