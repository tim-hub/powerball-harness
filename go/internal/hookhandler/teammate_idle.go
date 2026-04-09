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

// teammateIdleInput は TeammateIdle フックの stdin JSON ペイロード。
type teammateIdleInput struct {
	TeammateeName string `json:"teammate_name"`
	AgentName     string `json:"agent_name"`
	TeamName      string `json:"team_name"`
	AgentID       string `json:"agent_id"`
	AgentType     string `json:"agent_type"`
	Continue      *bool  `json:"continue"`
	StopReason    string `json:"stopReason"`
	StopReasonAlt string `json:"stop_reason"`
}

// teammateIdleLogEntry は breezing-timeline.jsonl に記録するエントリ。
type teammateIdleLogEntry struct {
	Event     string `json:"event"`
	Teammate  string `json:"teammate"`
	Team      string `json:"team"`
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
	Timestamp string `json:"timestamp"`
}

// teammateIdleApprove は approve レスポンス。
type teammateIdleApprove struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// teammateIdleStop は停止レスポンス。
type teammateIdleStop struct {
	Continue   bool   `json:"continue"`
	StopReason string `json:"stopReason"`
}

// timelineRotateMaxLines は JSONL ローテーションのしきい値。
const timelineRotateMaxLines = 500

// timelineRotateKeepLines はローテーション後に保持する行数。
const timelineRotateKeepLines = 400

// dedupWindowSeconds は同一エージェントの重複抑制ウィンドウ（秒）。
const dedupWindowSeconds = 5

// HandleTeammateIdle は teammate-idle.sh の Go 移植。
//
// TeammateIdle イベントの処理:
//  1. stdin JSON ペイロードを読み取る
//  2. 5秒 dedup（同一 agent_id で連続発火を抑制）
//  3. breezing-timeline.jsonl にアイドル状態を記録
//  4. continue:false または stop_reason がある場合は停止シグナルを送出
//  5. それ以外は approve を返す
func HandleTeammateIdle(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return writeTeammateIdleApprove(out, "TeammateIdle: no payload")
	}

	var input teammateIdleInput
	if err := json.Unmarshal(data, &input); err != nil {
		return writeTeammateIdleApprove(out, "TeammateIdle: no payload")
	}

	// teammate_name または agent_name を取得
	teammateName := input.TeammateeName
	if teammateName == "" {
		teammateName = input.AgentName
	}

	// stop_reason の正規化
	stopReason := input.StopReason
	if stopReason == "" {
		stopReason = input.StopReasonAlt
	}

	// continue フラグ
	hookContinue := true // デフォルトは続行
	if input.Continue != nil {
		hookContinue = *input.Continue
	}

	// プロジェクトルートを取得
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		if cwd, err := os.Getwd(); err == nil {
			projectRoot = cwd
		}
	}

	// 状態ディレクトリとタイムラインファイル
	stateDir := filepath.Join(projectRoot, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		// エラーは無視（続行）
		fmt.Fprintf(os.Stderr, "[claude-code-harness] teammate-idle: mkdir %s: %v\n", stateDir, err)
	}
	timelineFile := filepath.Join(stateDir, "breezing-timeline.jsonl")

	// === 重複抑制（同一 teammate の 5 秒以内の idle をスキップ） ===
	dedupKey := teammateName
	if dedupKey == "" {
		dedupKey = input.AgentID
	}

	if dedupKey != "" {
		if shouldSkip := checkTeammateIdleDedup(timelineFile, dedupKey); shouldSkip {
			return writeTeammateIdleApprove(out, "TeammateIdle dedup: skipped")
		}
	}

	// === タイムライン記録 ===
	ts := time.Now().UTC().Format(time.RFC3339)
	logEntry := teammateIdleLogEntry{
		Event:     "teammate_idle",
		Teammate:  teammateName,
		Team:      input.TeamName,
		AgentID:   input.AgentID,
		AgentType: input.AgentType,
		Timestamp: ts,
	}
	if entryData, err := json.Marshal(logEntry); err == nil {
		appendToJSONL(timelineFile, entryData)
		rotateJSONL(timelineFile, timelineRotateMaxLines, timelineRotateKeepLines)
	}

	// === レスポンス ===
	// continue:false または stop_reason がある場合は停止シグナルを送出
	if !hookContinue || stopReason != "" {
		finalStopReason := stopReason
		if finalStopReason == "" {
			finalStopReason = "TeammateIdle requested stop"
		}
		return writeTeammateIdleStop(out, finalStopReason)
	}

	return writeTeammateIdleApprove(out, "TeammateIdle tracked")
}

// checkTeammateIdleDedup は同一エージェントの最終 idle から dedupWindowSeconds 以内かどうかを確認する。
// teammate-idle.sh の重複抑制ロジックに対応。
func checkTeammateIdleDedup(timelineFile, dedupKey string) bool {
	data, err := os.ReadFile(timelineFile)
	if err != nil {
		return false // ファイルがない場合はスキップしない
	}

	// JSONL を逆順で走査して同一 teammate の最終 idle を探す
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// "teammate_idle" イベントかつ dedupKey を含む行を探す
		if !strings.Contains(line, `"teammate_idle"`) {
			continue
		}
		if !strings.Contains(line, dedupKey) {
			continue
		}

		var entry teammateIdleLogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		// テンプレート名またはエージェントIDが一致するか確認
		if entry.Teammate != dedupKey && entry.AgentID != dedupKey {
			continue
		}

		// タイムスタンプを解析して5秒以内かどうかを確認
		lastTime, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			continue
		}

		elapsed := time.Since(lastTime)
		if elapsed < dedupWindowSeconds*time.Second {
			return true // スキップ
		}
		return false // 5秒以上経過しているのでスキップしない
	}

	return false
}

// rotateJSONL は JSONL ファイルが maxLines を超えた場合に keepLines 行に切り詰める。
// teammate-idle.sh の rotate_jsonl() 関数に対応。
func rotateJSONL(path string, maxLines, keepLines int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) <= maxLines {
		return
	}

	// keepLines 行を残す
	if keepLines > len(lines) {
		keepLines = len(lines)
	}
	kept := lines[len(lines)-keepLines:]
	content := strings.Join(kept, "\n") + "\n"

	// 一時ファイルに書き込んでアトミックにリネーム
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0o644); err != nil {
		return
	}
	_ = os.Rename(tmpPath, path)
}

// writeTeammateIdleApprove は approve レスポンスを書き込む。
func writeTeammateIdleApprove(out io.Writer, reason string) error {
	resp := teammateIdleApprove{
		Decision: "approve",
		Reason:   reason,
	}
	return writeJSON(out, resp)
}

// writeTeammateIdleStop は停止シグナルのレスポンスを書き込む。
func writeTeammateIdleStop(out io.Writer, stopReason string) error {
	resp := teammateIdleStop{
		Continue:   false,
		StopReason: stopReason,
	}
	return writeJSON(out, resp)
}
