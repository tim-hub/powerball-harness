package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CIStatusCheckerHandler は PostToolUse (Bash) フックハンドラ（CI ステータスチェック）。
// git push / gh pr コマンドを検出し、CI ステータスを同期的にチェックする。
// async: true フック（CC がプロセスを最大 600s 生存させる）を前提に、
// goroutine ではなくブロッキング呼び出しで runner を実行する。
// CI 失敗時は additionalContext で /ci スキルを推奨する。
//
// shell 版: scripts/hook-handlers/ci-status-checker.sh
type CIStatusCheckerHandler struct {
	// ProjectRoot はプロジェクトルートのパス。空の場合は cwd を使用する。
	ProjectRoot string

	// GHCommand は gh コマンドのパス（テスト用）。空の場合は PATH から検索する。
	GHCommand string

	// AsyncRunner は CI チェックの実行関数（テスト用モック）。
	// nil の場合はデフォルト実装（同期ブロッキング）を使用する。
	AsyncRunner func(projectRoot, stateDir, bashCmd, ghCommand string)
}

// ciStatusInput は PostToolUse フックの入力。
type ciStatusInput struct {
	ToolName string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
	ToolResponse struct {
		ExitCode *int   `json:"exit_code"`
		ExitCode2 *int  `json:"exitCode"`
		Output   string `json:"output"`
		Stdout   string `json:"stdout"`
	} `json:"tool_response"`
}

// ciStatusResponse は CIStatusChecker フックのレスポンス。
type ciStatusResponse struct {
	Decision          string `json:"decision"`
	Reason            string `json:"reason"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

// ciRunEntry は gh run list の1件を表す。
type ciRunEntry struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Name       string `json:"name"`
	URL        string `json:"url"`
}

// pushOrPRCommandRe は git push / gh pr / gh workflow run を検出する正規表現。
var pushOrPRCommandRe = regexp.MustCompile(`(?:^|[\s;|&])(git\s+push|gh\s+pr\s+(?:create|merge|edit)|gh\s+workflow\s+run)`)

// Handle は stdin からペイロードを読み取り、push/PR コマンドを検出して CI 監視を起動する。
func (h *CIStatusCheckerHandler) Handle(r io.Reader, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil || len(data) == 0 {
		return writeCIJSON(w, ciStatusResponse{
			Decision: "approve",
			Reason:   "ci-status-checker: no payload",
		})
	}

	var input ciStatusInput
	if err := json.Unmarshal(data, &input); err != nil {
		return writeCIJSON(w, ciStatusResponse{
			Decision: "approve",
			Reason:   "ci-status-checker: parse error",
		})
	}

	bashCmd := input.ToolInput.Command

	// git push / gh pr コマンドでなければスキップ
	if !isPushOrPRCommand(bashCmd) {
		return writeCIJSON(w, ciStatusResponse{
			Decision: "approve",
			Reason:   "ci-status-checker: not a push/PR command",
		})
	}

	// gh コマンドが存在しない場合はスキップ
	ghCmd := h.resolveGHCommand()
	if ghCmd == "" {
		return writeCIJSON(w, ciStatusResponse{
			Decision: "approve",
			Reason:   "ci-status-checker: gh command not found",
		})
	}

	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	stateDir := filepath.Join(projectRoot, ".claude", "state")
	_ = os.MkdirAll(stateDir, 0700)

	// 直近の CI 失敗シグナルをチェック（runner 実行前に確認）
	additionalContext := h.checkRecentCIFailure(stateDir, bashCmd)

	// レスポンスを先に stdout に書き出す（async: true フックなので CC はプロセスを生かし続ける）
	var writeErr error
	if additionalContext != "" {
		writeErr = writeCIJSON(w, ciStatusResponse{
			Decision:          "approve",
			Reason:            "ci-status-checker: push/PR detected, CI failure context injected",
			AdditionalContext: additionalContext,
		})
	} else {
		writeErr = writeCIJSON(w, ciStatusResponse{
			Decision: "approve",
			Reason:   "ci-status-checker: push/PR detected, CI monitoring started",
		})
	}
	if writeErr != nil {
		return writeErr
	}

	// レスポンス書き出し後にブロッキングで CI ステータスをポーリング。
	// async: true フックなので CC がプロセスを最大 600s 生存させてくれる。
	// goroutine は不要 — プロセス終了で kill されるリスクを排除する。
	runner := h.AsyncRunner
	if runner == nil {
		runner = defaultCIRunner
	}
	runner(projectRoot, stateDir, bashCmd, ghCmd)
	return nil
}

// isPushOrPRCommand は bashCmd が push / PR コマンドを含むかを返す。
func isPushOrPRCommand(cmd string) bool {
	return pushOrPRCommandRe.MatchString(cmd)
}

// resolveGHCommand は gh コマンドのパスを返す。存在しない場合は空文字列。
func (h *CIStatusCheckerHandler) resolveGHCommand() string {
	if h.GHCommand != "" {
		if _, err := os.Stat(h.GHCommand); err == nil {
			return h.GHCommand
		}
		return ""
	}
	path, err := exec.LookPath("gh")
	if err != nil {
		return ""
	}
	return path
}

// checkRecentCIFailure は直近の ci_failure_detected シグナルを確認し、メッセージを返す。
func (h *CIStatusCheckerHandler) checkRecentCIFailure(stateDir, bashCmd string) string {
	signalsFile := filepath.Join(stateDir, "breezing-signals.jsonl")

	f, err := os.Open(signalsFile)
	if err != nil {
		return ""
	}
	defer f.Close()

	// 最後の ci_failure_detected シグナルを探す
	var lastFailureLine string
	buf := make([]byte, 1<<20) // 最大 1MB
	n, _ := f.Read(buf)
	content := string(buf[:n])

	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, `"ci_failure_detected"`) {
			lastFailureLine = line
		}
	}

	if lastFailureLine == "" {
		return ""
	}

	var sig map[string]interface{}
	if err := json.Unmarshal([]byte(lastFailureLine), &sig); err != nil {
		return ""
	}

	conclusion, _ := sig["conclusion"].(string)
	return fmt.Sprintf(
		"[CI 失敗を検知しました]\nCI ステータス: %s\nトリガーコマンド: %s\n\n推奨アクション: /breezing または ci-cd-fixer エージェントを spawn して CI 障害を自動修復してください。\n  例: ci-cd-fixer に「CI が失敗しました。ログを確認して修正してください。」と依頼",
		conclusion, bashCmd,
	)
}

// defaultCIRunner は gh run list でポーリングし、結果をシグナルファイルに書き込む。
// async: true フックを前提に同期ブロッキングで実行される。
// maxWait を hooks.json の timeout (30s) より短い 25s に設定して安全マージンを確保する。
func defaultCIRunner(projectRoot, stateDir, bashCmd, ghCmd string) {
	const maxWait = 25 * time.Second
	const pollInterval = 10 * time.Second

	ciStatusFile := filepath.Join(stateDir, "ci-status.json")
	signalsFile := filepath.Join(stateDir, "breezing-signals.jsonl")

	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		time.Sleep(pollInterval)

		out, err := exec.Command(ghCmd, "run", "list", "--limit", "1", "--json", "status,conclusion,name,url").Output()
		if err != nil || len(out) == 0 {
			continue
		}

		var runs []ciRunEntry
		if err := json.Unmarshal(out, &runs); err != nil || len(runs) == 0 {
			continue
		}

		run := runs[0]
		if run.Status != "completed" {
			continue
		}

		// 結果を記録
		statusData, _ := json.Marshal(map[string]string{
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
			"trigger_command": bashCmd,
			"status":          run.Status,
			"conclusion":      run.Conclusion,
		})
		_ = os.WriteFile(ciStatusFile, statusData, 0600)

		// CI 失敗の場合はシグナルファイルに追記
		if run.Conclusion == "failure" || run.Conclusion == "timed_out" || run.Conclusion == "cancelled" {
			sig, _ := json.Marshal(map[string]string{
				"signal":          "ci_failure_detected",
				"timestamp":       time.Now().UTC().Format(time.RFC3339),
				"conclusion":      run.Conclusion,
				"trigger_command": bashCmd,
			})
			f, err := os.OpenFile(signalsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err == nil {
				_, _ = f.Write(sig)
				_, _ = f.Write([]byte("\n"))
				f.Close()
			}
		}

		return
	}
}

// writeCIJSON は v を JSON として w に書き出す。
func writeCIJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
