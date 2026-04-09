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

// configChangeInput は config-change.sh に渡される stdin JSON。
// ConfigChange イベントのペイロード。
type configChangeInput struct {
	FilePath   string `json:"file_path"`
	ChangeType string `json:"change_type"`
}

// breezingState は .claude/state/breezing.json の構造。
type breezingState struct {
	Status string `json:"status"`
}

// okOutput は {"ok":true} レスポンス。
type okOutput struct {
	OK bool `json:"ok"`
}

// HandleConfigChange は config-change.sh の Go 移植。
//
// ConfigChange イベントで呼び出され、Breezing がアクティブな場合のみ
// .claude/state/breezing-timeline.jsonl に記録する。
// 常に {"ok":true} を返す（Stop をブロックしない）。
func HandleConfigChange(in io.Reader, out io.Writer) error {
	// stdin から JSON を読み取る（サイズ上限 64KB）
	lr := io.LimitReader(in, 65536)
	data, err := io.ReadAll(lr)
	if err != nil {
		return writeJSON(out, okOutput{OK: true})
	}

	payload := strings.TrimSpace(string(data))
	if payload == "" {
		return writeJSON(out, okOutput{OK: true})
	}

	// PROJECT_ROOT を解決（環境変数優先、なければ cwd）
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			return writeJSON(out, okOutput{OK: true})
		}
		projectRoot = cwd
	}

	// breezing がアクティブかどうか確認
	breezingStateFile := filepath.Join(projectRoot, ".claude", "state", "breezing.json")
	if !isBreezingActive(breezingStateFile) {
		return writeJSON(out, okOutput{OK: true})
	}

	// ペイロードをパース
	var input configChangeInput
	if jsonErr := json.Unmarshal([]byte(payload), &input); jsonErr != nil {
		// パース失敗でも ok を返す
		return writeJSON(out, okOutput{OK: true})
	}

	// file_path をリポジトリ相対パスに正規化（ユーザー名等を隠蔽）
	rawPath := input.FilePath
	if rawPath == "" {
		rawPath = "unknown"
	}
	relPath := rawPath
	if rawPath != "unknown" && projectRoot != "" {
		trimmed := strings.TrimPrefix(rawPath, projectRoot+"/")
		if trimmed != rawPath {
			relPath = trimmed
		}
	}

	changeType := input.ChangeType
	if changeType == "" {
		changeType = "modified"
	}

	ts := time.Now().UTC().Format(time.RFC3339)

	// タイムラインに記録
	timelineFile := filepath.Join(projectRoot, ".claude", "state", "breezing-timeline.jsonl")
	if mkdirErr := os.MkdirAll(filepath.Dir(timelineFile), 0o755); mkdirErr == nil {
		event := map[string]string{
			"type":        "config_change",
			"timestamp":   ts,
			"file_path":   relPath,
			"change_type": changeType,
		}
		if eventData, marshalErr := json.Marshal(event); marshalErr == nil {
			f, openErr := os.OpenFile(timelineFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if openErr == nil {
				fmt.Fprintf(f, "%s\n", eventData)
				f.Close()
			}
		}
	}

	return writeJSON(out, okOutput{OK: true})
}

// isBreezingActive は breezing.json を読み込み、status が active または running かを確認する。
func isBreezingActive(stateFile string) bool {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return false
	}
	var state breezingState
	if err := json.Unmarshal(data, &state); err != nil {
		return false
	}
	return state.Status == "active" || state.Status == "running"
}
