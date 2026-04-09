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

// instructionsLoadedInput は instructions-loaded.sh に渡される stdin JSON。
type instructionsLoadedInput struct {
	SessionID     string `json:"session_id"`
	CWD           string `json:"cwd"`
	AgentID       string `json:"agent_id"`
	AgentType     string `json:"agent_type"`
	HookEventName string `json:"hook_event_name"`
	EventName     string `json:"event_name"`
}

// approveOutput は {"decision":"approve","reason":"..."} レスポンス。
type approveOutput struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// HandleInstructionsLoaded は instructions-loaded.sh の Go 移植。
//
// InstructionsLoaded イベントで呼び出され、以下を実行する:
// 1. .claude/state/instructions-loaded.jsonl にイベントを記録
// 2. hooks.json の存在を軽量に検証
//
// 常に {"decision":"approve",...} を返す（ブロックしない）。
func HandleInstructionsLoaded(in io.Reader, out io.Writer) error {
	// stdin から JSON を読み取る
	data, err := io.ReadAll(in)
	if err != nil {
		return writeJSON(out, approveOutput{
			Decision: "approve",
			Reason:   "InstructionsLoaded: read error",
		})
	}

	payload := strings.TrimSpace(string(data))
	if payload == "" {
		return writeJSON(out, approveOutput{
			Decision: "approve",
			Reason:   "InstructionsLoaded: no payload",
		})
	}

	// ペイロードをパース
	var input instructionsLoadedInput
	if jsonErr := json.Unmarshal([]byte(payload), &input); jsonErr != nil {
		return writeJSON(out, approveOutput{
			Decision: "approve",
			Reason:   "InstructionsLoaded: parse error",
		})
	}

	// PROJECT_ROOT を解決（CWD フィールド優先、なければ env、なければ pwd）
	projectRoot := input.CWD
	if projectRoot == "" {
		projectRoot = os.Getenv("PROJECT_ROOT")
	}
	if projectRoot == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr == nil {
			projectRoot = cwd
		}
	}

	// event_name を解決（hook_event_name 優先、event_name フォールバック）
	eventName := input.HookEventName
	if eventName == "" {
		eventName = input.EventName
	}
	if eventName == "" {
		eventName = "InstructionsLoaded"
	}

	// .claude/state/instructions-loaded.jsonl にイベントを記録
	stateDir := filepath.Join(projectRoot, ".claude", "state")
	logFile := filepath.Join(stateDir, "instructions-loaded.jsonl")

	if mkdirErr := os.MkdirAll(stateDir, 0o755); mkdirErr == nil {
		ts := time.Now().UTC().Format(time.RFC3339)
		event := map[string]string{
			"event":      eventName,
			"timestamp":  ts,
			"session_id": input.SessionID,
			"agent_id":   input.AgentID,
			"agent_type": input.AgentType,
			"cwd":        projectRoot,
		}
		if eventData, marshalErr := json.Marshal(event); marshalErr == nil {
			f, openErr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if openErr == nil {
				fmt.Fprintf(f, "%s\n", eventData)
				f.Close()
			}
		}
	}

	// hooks.json の存在を軽量に検証
	hooksFound := fileExists(filepath.Join(projectRoot, "hooks", "hooks.json")) ||
		fileExists(filepath.Join(projectRoot, ".claude-plugin", "hooks.json"))

	if !hooksFound {
		return writeJSON(out, approveOutput{
			Decision: "approve",
			Reason:   "InstructionsLoaded: hooks.json not found in project root",
		})
	}

	return writeJSON(out, approveOutput{
		Decision: "approve",
		Reason:   "InstructionsLoaded tracked",
	})
}

// fileExists はファイルが存在するかを確認する。
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
