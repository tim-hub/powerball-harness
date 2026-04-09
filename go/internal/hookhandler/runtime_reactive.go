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

// reactiveInput は runtime-reactive.sh に渡される stdin JSON ペイロード。
type reactiveInput struct {
	HookEventName string `json:"hook_event_name"`
	EventName     string `json:"event_name"`
	SessionID     string `json:"session_id"`
	CWD           string `json:"cwd"`
	ProjectRoot   string `json:"project_root"`
	FilePath      string `json:"file_path"`
	Path          string `json:"path"`
	PreviousCWD   string `json:"previous_cwd"`
	FromCWD       string `json:"from_cwd"`
	TaskID        string `json:"task_id"`
	TaskTitle     string `json:"task_title"`
	Description   string `json:"description"`
	Task          *struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"task"`
}

// reactiveLogEntry は runtime-reactive.jsonl に記録するエントリ。
type reactiveLogEntry struct {
	Event       string `json:"event"`
	Timestamp   string `json:"timestamp"`
	SessionID   string `json:"session_id"`
	CWD         string `json:"cwd"`
	FilePath    string `json:"file_path"`
	PreviousCWD string `json:"previous_cwd"`
	TaskID      string `json:"task_id"`
	TaskTitle   string `json:"task_title"`
}

// reactiveHookOutput は hookSpecificOutput 形式のレスポンス。
type reactiveHookOutput struct {
	HookSpecificOutput struct {
		HookEventName     string `json:"hookEventName"`
		AdditionalContext string `json:"additionalContext"`
	} `json:"hookSpecificOutput"`
}

// reactiveApproveOutput は approve 形式のレスポンス。
type reactiveApproveOutput struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// HandleRuntimeReactive は runtime-reactive.sh の Go 移植。
//
// TaskCreated / FileChanged / CwdChanged の3イベントを統合処理する:
//   - TaskCreated: バックグラウンドタスク作成ログ
//   - FileChanged: Plans.md, AGENTS.md, .claude/rules/, hooks.json, settings.json の変更検知 → systemMessage で再読指示
//   - CwdChanged: ディレクトリ変更 → コンテキスト再確認促し
func HandleRuntimeReactive(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return writeReactiveApprove(out, "Reactive hook: no payload")
	}

	var input reactiveInput
	if err := json.Unmarshal(data, &input); err != nil {
		return writeReactiveApprove(out, "Reactive hook: no payload")
	}

	// hook_event_name または event_name を取得
	eventName := input.HookEventName
	if eventName == "" {
		eventName = input.EventName
	}

	// project_root を決定
	projectRoot := input.CWD
	if projectRoot == "" {
		projectRoot = input.ProjectRoot
	}
	if projectRoot == "" {
		if cwd, err := os.Getwd(); err == nil {
			projectRoot = cwd
		}
	}

	// ファイルパスを取得して正規化
	filePath := input.FilePath
	if filePath == "" {
		filePath = input.Path
	}
	filePath = normalizeReactivePath(filePath, projectRoot)

	// previous_cwd を正規化
	previousCWD := input.PreviousCWD
	if previousCWD == "" {
		previousCWD = input.FromCWD
	}
	previousCWD = normalizeReactivePath(previousCWD, projectRoot)

	// task_id と task_title を取得
	taskID := input.TaskID
	taskTitle := input.TaskTitle
	if input.Task != nil {
		if taskID == "" {
			taskID = input.Task.ID
		}
		if taskTitle == "" {
			taskTitle = input.Task.Title
			if taskTitle == "" {
				taskTitle = input.Task.Description
			}
		}
	}
	if taskTitle == "" {
		taskTitle = input.Description
	}

	// session_id
	sessionID := input.SessionID

	// 状態ディレクトリとログファイル
	stateDir := filepath.Join(projectRoot, ".claude", "state")
	logFile := filepath.Join(stateDir, "runtime-reactive.jsonl")
	_ = os.MkdirAll(stateDir, 0o755)

	timestamp := time.Now().UTC().Format(time.RFC3339)

	// ログエントリを記録
	logEntry := reactiveLogEntry{
		Event:       eventName,
		Timestamp:   timestamp,
		SessionID:   sessionID,
		CWD:         projectRoot,
		FilePath:    filePath,
		PreviousCWD: previousCWD,
		TaskID:      taskID,
		TaskTitle:   taskTitle,
	}
	if logData, err := json.Marshal(logEntry); err == nil {
		appendToJSONL(logFile, logData)
	}

	// イベントごとのメッセージ生成
	message := buildReactiveMessage(eventName, filePath)

	if message != "" {
		return writeReactiveHookOutput(out, eventName, message)
	}
	return writeReactiveApprove(out, fmt.Sprintf("Reactive hook tracked: %s", eventName))
}

// buildReactiveMessage はイベントと変更パスに応じたメッセージを生成する。
// runtime-reactive.sh の case 文に対応。
func buildReactiveMessage(eventName, filePath string) string {
	switch eventName {
	case "FileChanged":
		return buildFileChangedMessage(filePath)
	case "CwdChanged":
		return "作業ディレクトリが切り替わりました。別リポジトリや worktree に移動した場合は AGENTS.md、Plans.md、ローカルルールを再確認してください。"
	case "TaskCreated":
		// TaskCreated はログ記録のみ（メッセージなし）
		return ""
	}
	return ""
}

// buildFileChangedMessage は FileChanged イベントのメッセージを生成する。
func buildFileChangedMessage(filePath string) string {
	// Plans.md の変更
	if filePath == "Plans.md" || strings.HasSuffix(filePath, "/Plans.md") {
		return "Plans.md が更新されました。次の実装やレビュー前に最新のタスク状態を読み直してください。"
	}

	// AGENTS.md, CLAUDE.md, .claude/rules/, hooks.json, settings.json の変更
	if isRuleOrConfigFile(filePath) {
		return "作業ルールまたは Harness 設定が更新されました。次の操作では最新ルールを前提に進めてください。"
	}

	return ""
}

// isRuleOrConfigFile はファイルパスがルール/設定ファイルかどうかを判定する。
// runtime-reactive.sh の case パターンに対応:
// AGENTS.md, CLAUDE.md, .claude/rules/*, hooks/hooks.json, .claude-plugin/settings.json
func isRuleOrConfigFile(filePath string) bool {
	// AGENTS.md
	if filePath == "AGENTS.md" || strings.HasSuffix(filePath, "/AGENTS.md") {
		return true
	}
	// CLAUDE.md
	if filePath == "CLAUDE.md" || strings.HasSuffix(filePath, "/CLAUDE.md") {
		return true
	}
	// .claude/rules/ ディレクトリ配下
	if strings.HasPrefix(filePath, ".claude/rules/") || strings.Contains(filePath, "/.claude/rules/") {
		return true
	}
	// hooks/hooks.json
	if filePath == "hooks/hooks.json" || strings.HasSuffix(filePath, "/hooks/hooks.json") {
		return true
	}
	// .claude-plugin/settings.json
	if filePath == ".claude-plugin/settings.json" || strings.HasSuffix(filePath, "/.claude-plugin/settings.json") {
		return true
	}
	return false
}

// normalizeReactivePath はパスをプロジェクトルートからの相対パスに正規化する。
// runtime-reactive.sh の normalize_for_match() 関数に対応。
func normalizeReactivePath(rawPath, projectRoot string) string {
	if rawPath == "" {
		return ""
	}

	// シンボリックリンクを解決（できる場合）
	if resolved, err := filepath.EvalSymlinks(rawPath); err == nil {
		rawPath = resolved
	}

	if projectRoot != "" {
		// シンボリックリンクを解決（できる場合）
		normalizedRoot := projectRoot
		if resolved, err := filepath.EvalSymlinks(projectRoot); err == nil {
			normalizedRoot = resolved
		}

		// プロジェクトルートとの相対化
		if rawPath == normalizedRoot {
			rawPath = "."
		} else if strings.HasPrefix(rawPath, normalizedRoot+"/") {
			rawPath = rawPath[len(normalizedRoot)+1:]
		}
	}

	// ./ プレフィックスを除去
	rawPath = strings.TrimPrefix(rawPath, "./")
	return rawPath
}

// appendToJSONL は JSONL ファイルにエントリを追記する。
func appendToJSONL(path string, data []byte) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(data, '\n'))
}

// writeReactiveHookOutput は hookSpecificOutput 形式のレスポンスを書き込む。
func writeReactiveHookOutput(out io.Writer, eventName, message string) error {
	var resp reactiveHookOutput
	resp.HookSpecificOutput.HookEventName = eventName
	resp.HookSpecificOutput.AdditionalContext = message
	return writeJSON(out, resp)
}

// writeReactiveApprove は approve 形式のレスポンスを書き込む。
func writeReactiveApprove(out io.Writer, reason string) error {
	resp := reactiveApproveOutput{
		Decision: "approve",
		Reason:   reason,
	}
	return writeJSON(out, resp)
}
