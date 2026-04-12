// Package session implements session lifecycle handlers for Claude Code Harness.
//
// Each handler corresponds to a shell script that was previously used:
//   - Init      → scripts/session-init.sh
//   - Cleanup   → scripts/session-cleanup.sh
//   - Monitor   → scripts/session-monitor.sh
//   - Summary   → scripts/session-summary.sh
//
// Handlers read hook JSON from stdin and write the appropriate response to stdout.
package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// InitHandler
// ---------------------------------------------------------------------------

// InitHandler は SessionStart フックハンドラ。
// session-init.sh の主要機能を Go に移植する:
//  1. サブエージェント時の軽量初期化
//  2. セッション JSON の初期化 (session.json)
//  3. Plans.md のタスクカウント
//  4. additionalContext を含む JSON レスポンス
//
// shell 版: scripts/session-init.sh
type InitHandler struct {
	// StateDir は .claude/state ディレクトリのパス。空の場合は cwd から推定する。
	StateDir string
	// PlansFile は Plans.md のパス。空の場合は projectRoot/Plans.md を使う。
	PlansFile string
}

// initInput は SessionStart フックの stdin JSON。
type initInput struct {
	SessionID string `json:"session_id,omitempty"`
	AgentType string `json:"agent_type,omitempty"`
	CWD       string `json:"cwd,omitempty"`
}

// sessionJSON は session.json のスキーマ（最低限）。
type sessionJSON struct {
	SessionID  string `json:"session_id"`
	State      string `json:"state"`
	StartedAt  string `json:"started_at"`
	UpdatedAt  string `json:"updated_at"`
	EventSeq   int    `json:"event_seq"`
	LastEventID string `json:"last_event_id"`
}

// initResponse は SessionStart フックへの JSON 出力。
type initResponse struct {
	HookSpecificOutput initHookOutput `json:"hookSpecificOutput"`
}

type initHookOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

// Handle は stdin から SessionStart ペイロードを読み取り、
// セッション初期化を行い、additionalContext を含む JSON を stdout に書き出す。
func (h *InitHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	var inp initInput
	if len(data) > 0 {
		_ = json.Unmarshal(data, &inp)
	}

	// サブエージェント時は軽量初期化（session.json 操作をスキップ）
	if inp.AgentType == "subagent" {
		return writeJSON(w, initResponse{
			HookSpecificOutput: initHookOutput{
				HookEventName:     "SessionStart",
				AdditionalContext: "[subagent] 軽量初期化完了",
			},
		})
	}

	// プロジェクトルートとステートディレクトリを決定
	projectRoot := resolveProjectRoot(inp.CWD)
	stateDir := h.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(projectRoot, ".claude", "state")
	}

	// ステートディレクトリを作成（シンボリックリンクチェック付き）
	if err := ensureStateDir(stateDir); err != nil {
		// エラーでも処理を継続（バナーと Plans 情報は出力する）
		_ = err
	}

	// session.json を初期化（存在しないか停止状態の場合）
	_ = h.initSessionFile(stateDir)

	// session-skills-used.json をリセット
	skillsUsedFile := filepath.Join(stateDir, "session-skills-used.json")
	now := time.Now().UTC().Format(time.RFC3339)
	_ = writeFileAtomic(skillsUsedFile, []byte(fmt.Sprintf(`{"used":[],"session_start":%q}`, now)+"\n"), 0600)

	// SSOT 同期フラグをクリア
	_ = os.Remove(filepath.Join(stateDir, ".ssot-synced-this-session"))
	// work 警告フラグをクリア
	_ = os.Remove(filepath.Join(stateDir, ".work-review-warned"))
	_ = os.Remove(filepath.Join(stateDir, ".ultrawork-review-warned"))

	// Plans.md カウント
	plansFile := h.PlansFile
	if plansFile == "" {
		plansFile = filepath.Join(projectRoot, "Plans.md")
	}
	plansInfo := buildPlansInfo(plansFile)

	// マーカー凡例を追記
	context := buildAdditionalContext(plansInfo)

	return writeJSON(w, initResponse{
		HookSpecificOutput: initHookOutput{
			HookEventName:     "SessionStart",
			AdditionalContext: context,
		},
	})
}

// initSessionFile は session.json を初期化する。
// 既存ファイルが active 状態（initialized/running/working）なら何もしない。
func (h *InitHandler) initSessionFile(stateDir string) error {
	sessionFile := filepath.Join(stateDir, "session.json")

	if isSymlink(sessionFile) {
		return fmt.Errorf("security: symlinked session file: %s", sessionFile)
	}

	// 既存ファイルの状態を確認
	if data, err := os.ReadFile(sessionFile); err == nil {
		var s sessionJSON
		if json.Unmarshal(data, &s) == nil {
			// stopped/completed/failed 以外はそのまま
			switch s.State {
			case "stopped", "completed", "failed":
				// 新規初期化が必要
			default:
				return nil
			}
		}
	}

	// 新規セッション初期化
	now := time.Now().UTC().Format(time.RFC3339)
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	s := sessionJSON{
		SessionID:  sessionID,
		State:      "initialized",
		StartedAt:  now,
		UpdatedAt:  now,
		EventSeq:   0,
		LastEventID: "",
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return writeFileAtomic(sessionFile, append(data, '\n'), 0600)
}

// buildPlansInfo は Plans.md を読んで WIP/TODO カウントの情報文字列を返す。
func buildPlansInfo(plansFile string) string {
	if _, err := os.Stat(plansFile); err != nil {
		return "Plans.md: 未検出"
	}

	wipCount := countMatches(plansFile, "cc:WIP", "pm:依頼中", "cursor:依頼中")
	todoCount := countMatches(plansFile, "cc:TODO")

	return fmt.Sprintf("Plans.md: 進行中 %d / 未着手 %d", wipCount, todoCount)
}

// buildAdditionalContext はセッション初期化の additionalContext を構築する。
func buildAdditionalContext(plansInfo string) string {
	var sb strings.Builder
	sb.WriteString("# [claude-code-harness] セッション初期化\n\n")
	sb.WriteString(plansInfo + "\n")
	sb.WriteString("\n## マーカー凡例\n")
	sb.WriteString("| マーカー | 状態 | 説明 |\n")
	sb.WriteString("|---------|------|------|\n")
	sb.WriteString("| `cc:TODO` | 未着手 | Impl（Claude Code）が実行予定 |\n")
	sb.WriteString("| `cc:WIP` | 作業中 | Impl が実装中 |\n")
	sb.WriteString("| `cc:blocked` | ブロック中 | 依存タスク待ち |\n")
	sb.WriteString("| `pm:依頼中` | PM から依頼 | 2-Agent 運用時 |\n")
	sb.WriteString("\n> **互換**: `cursor:依頼中` / `cursor:確認済` は `pm:*` と同義として扱います。\n")
	return sb.String()
}

// ---------------------------------------------------------------------------
// ユーティリティ（package-private）
// ---------------------------------------------------------------------------

// writeJSON は v を JSON として w に書き出す。
func writeJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

// resolveProjectRoot は CWD フィールドや環境変数からプロジェクトルートを推測する。
func resolveProjectRoot(cwd string) string {
	if cwd != "" {
		return cwd
	}
	if r := os.Getenv("HARNESS_PROJECT_ROOT"); r != "" {
		return r
	}
	if r := os.Getenv("PROJECT_ROOT"); r != "" {
		return r
	}
	root, _ := os.Getwd()
	return root
}

// ensureStateDir はステートディレクトリを作成する。
// シンボリックリンクの場合はエラーを返す。
func ensureStateDir(stateDir string) error {
	parent := filepath.Dir(stateDir)
	if isSymlink(parent) || isSymlink(stateDir) {
		return fmt.Errorf("security: symlinked state path refused: %s", stateDir)
	}
	return os.MkdirAll(stateDir, 0700)
}

// isSymlink はパスがシンボリックリンクかどうかを返す。
func isSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// countMatches は patterns のいずれかを含む行数の合計を返す。
func countMatches(filePath string, patterns ...string) int {
	f, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		for _, p := range patterns {
			if strings.Contains(line, p) {
				count++
				break
			}
		}
	}
	return count
}

// writeFileAtomic はファイルを一時ファイル経由で原子的に書き出す。
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	if isSymlink(path) {
		return fmt.Errorf("security: symlinked file refused: %s", path)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
