package session

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// MonitorHandler は SessionStart フックハンドラ（プロジェクト状態収集）。
// セッション開始時にプロジェクト状態を収集し、session.json と tooling-policy.json を生成する。
//
// shell 版: scripts/session-monitor.sh
type MonitorHandler struct {
	// StateDir はステートディレクトリのパス。空の場合は cwd から推定する。
	StateDir string
	// PlansFile は Plans.md のパス。空の場合は projectRoot/Plans.md を使う。
	PlansFile string
	// now は現在時刻の注入関数（テスト用）。nil の場合は time.Now() を使う。
	now func() time.Time
}

// monitorInput は SessionStart フックの stdin JSON。
type monitorInput struct {
	SessionID string `json:"session_id,omitempty"`
	CWD       string `json:"cwd,omitempty"`
}

// sessionStateJSON は session.json の完全なスキーマ。
type sessionStateJSON struct {
	SessionID    string            `json:"session_id"`
	ParentID     interface{}       `json:"parent_session_id"`
	State        string            `json:"state"`
	StateVersion int               `json:"state_version"`
	StartedAt    string            `json:"started_at"`
	UpdatedAt    string            `json:"updated_at"`
	ResumeToken  string            `json:"resume_token"`
	EventSeq     int               `json:"event_seq"`
	LastEventID  string            `json:"last_event_id"`
	ForkCount    int               `json:"fork_count"`
	Orchestration orchestrationJSON `json:"orchestration"`
	CWD          string            `json:"cwd"`
	ProjectName  string            `json:"project_name"`
	PromptSeq    int               `json:"prompt_seq"`
	Git          gitStateJSON      `json:"git"`
	Plans        plansStateJSON    `json:"plans"`
	ChangesThisSession []interface{} `json:"changes_this_session"`
}

type orchestrationJSON struct {
	MaxStateRetries      int `json:"max_state_retries"`
	RetryBackoffSeconds  int `json:"retry_backoff_seconds"`
}

type gitStateJSON struct {
	Branch             string `json:"branch"`
	UncommittedChanges int    `json:"uncommitted_changes"`
	LastCommit         string `json:"last_commit"`
}

type plansStateJSON struct {
	Exists         bool   `json:"exists"`
	LastModified   int64  `json:"last_modified"`
	WIPTasks       int    `json:"wip_tasks"`
	TODOTasks      int    `json:"todo_tasks"`
	PendingTasks   int    `json:"pending_tasks"`
	CompletedTasks int    `json:"completed_tasks"`
}

// toolingPolicyJSON は tooling-policy.json のスキーマ（簡略版）。
// LSP/MCP 検出の重い外部コマンド依存を避け、基本情報のみを生成する。
type toolingPolicyJSON struct {
	LSP     lspPolicyJSON     `json:"lsp"`
	Plugins pluginPolicyJSON  `json:"plugins"`
	MCP     mcpPolicyJSON     `json:"mcp"`
	Skills  skillsPolicyJSON  `json:"skills"`
}

type lspPolicyJSON struct {
	Available         bool              `json:"available"`
	Plugins           string            `json:"plugins"`
	AvailableByExt    map[string]bool   `json:"available_by_ext"`
	LastUsedPromptSeq int               `json:"last_used_prompt_seq"`
	LastUsedToolName  string            `json:"last_used_tool_name"`
	UsedSinceLastPrompt bool            `json:"used_since_last_prompt"`
}

type pluginPolicyJSON struct {
	Installed       *int    `json:"installed"`
	EnabledEstimate *int    `json:"enabled_estimate"`
	Source          string  `json:"source"`
}

type mcpPolicyJSON struct {
	Configured      *int     `json:"configured"`
	Disabled        *int     `json:"disabled"`
	EnabledEstimate *int     `json:"enabled_estimate"`
	Sources         []string `json:"sources"`
}

type skillsPolicyJSON struct {
	Index           []interface{} `json:"index"`
	DecisionRequired bool         `json:"decision_required"`
}

// Handle は stdin から SessionStart ペイロードを読み取り、
// session.json と tooling-policy.json を生成して stdout に状態サマリーを書き出す。
func (h *MonitorHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	var inp monitorInput
	if len(data) > 0 {
		_ = json.Unmarshal(data, &inp)
	}

	projectRoot := resolveProjectRoot(inp.CWD)
	if projectRoot == "" {
		cwd, _ := os.Getwd()
		projectRoot = cwd
	}

	stateDir := h.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(projectRoot, ".claude", "state")
	}

	// シンボリックリンクチェック
	if isSymlink(stateDir) || isSymlink(filepath.Dir(stateDir)) {
		fmt.Fprintf(os.Stderr, "[session-monitor] Warning: symlink detected in state directory path, aborting\n")
		return nil
	}

	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return nil
	}

	plansFile := h.PlansFile
	if plansFile == "" {
		plansFile = filepath.Join(projectRoot, "Plans.md")
	}

	now := h.currentTime()
	nowStr := now.UTC().Format(time.RFC3339)

	// プロジェクト情報を収集
	projectName := filepath.Base(projectRoot)
	gitState := h.collectGitState(projectRoot)
	plansState := h.collectPlansState(plansFile)

	// session.json を生成（resume/新規を判定）
	sessionFile := filepath.Join(stateDir, "session.json")
	h.generateSessionFile(sessionFile, projectRoot, projectName, nowStr, gitState, plansState)

	// tooling-policy.json を生成
	policyFile := filepath.Join(stateDir, "tooling-policy.json")
	h.generateToolingPolicy(policyFile)

	// サマリーを stdout に出力
	h.writeSummary(w, projectName, gitState, plansState)
	return nil
}

// collectGitState は git 情報を収集する。
func (h *MonitorHandler) collectGitState(projectRoot string) gitStateJSON {
	gitDir := filepath.Join(projectRoot, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return gitStateJSON{
			Branch:             "(no git)",
			UncommittedChanges: 0,
			LastCommit:         "none",
		}
	}

	// HEAD からブランチ名を読み取る
	branch := h.readGitBranch(projectRoot)
	return gitStateJSON{
		Branch:             branch,
		UncommittedChanges: 0, // 重い操作を避けるため 0 固定
		LastCommit:         h.readGitLastCommit(projectRoot),
	}
}

// readGitBranch は .git/HEAD からブランチ名を読み取る。
func (h *MonitorHandler) readGitBranch(projectRoot string) string {
	headFile := filepath.Join(projectRoot, ".git", "HEAD")
	data, err := os.ReadFile(headFile)
	if err != nil {
		return "unknown"
	}
	line := strings.TrimSpace(string(data))
	// "ref: refs/heads/<branch>"
	if strings.HasPrefix(line, "ref: refs/heads/") {
		return strings.TrimPrefix(line, "ref: refs/heads/")
	}
	// detached HEAD: SHA のみ
	if len(line) >= 7 {
		return line[:7]
	}
	return "unknown"
}

// readGitLastCommit は .git/refs/heads/<branch> から最新コミット SHA を読み取る。
func (h *MonitorHandler) readGitLastCommit(projectRoot string) string {
	headFile := filepath.Join(projectRoot, ".git", "HEAD")
	data, err := os.ReadFile(headFile)
	if err != nil {
		return "none"
	}
	line := strings.TrimSpace(string(data))
	if strings.HasPrefix(line, "ref: ") {
		ref := strings.TrimPrefix(line, "ref: ")
		refFile := filepath.Join(projectRoot, ".git", filepath.FromSlash(ref))
		refData, err := os.ReadFile(refFile)
		if err != nil {
			// packed-refs をフォールバックで確認
			return h.readPackedRef(projectRoot, ref)
		}
		sha := strings.TrimSpace(string(refData))
		if len(sha) >= 7 {
			return sha[:7]
		}
		return sha
	}
	// detached HEAD
	if len(line) >= 7 {
		return line[:7]
	}
	return "none"
}

// readPackedRef は .git/packed-refs から ref を検索する。
func (h *MonitorHandler) readPackedRef(projectRoot, ref string) string {
	packedFile := filepath.Join(projectRoot, ".git", "packed-refs")
	data, err := os.ReadFile(packedFile)
	if err != nil {
		return "none"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasSuffix(line, " "+ref) || strings.HasSuffix(line, "\t"+ref) {
			parts := strings.Fields(line)
			if len(parts) >= 1 && len(parts[0]) >= 7 {
				return parts[0][:7]
			}
		}
	}
	return "none"
}

// collectPlansState は Plans.md の状態を収集する。
func (h *MonitorHandler) collectPlansState(plansFile string) plansStateJSON {
	fi, err := os.Stat(plansFile)
	if err != nil {
		return plansStateJSON{Exists: false}
	}

	wipCount := countMatches(plansFile, "cc:WIP")
	todoCount := countMatches(plansFile, "cc:TODO")
	pendingCount := countMatches(plansFile, "pm:依頼中", "cursor:依頼中")
	completedCount := countMatches(plansFile, "cc:完了")

	return plansStateJSON{
		Exists:         true,
		LastModified:   fi.ModTime().Unix(),
		WIPTasks:       wipCount,
		TODOTasks:      todoCount,
		PendingTasks:   pendingCount,
		CompletedTasks: completedCount,
	}
}

// generateSessionFile は session.json を生成する（resume/新規を判定）。
func (h *MonitorHandler) generateSessionFile(
	sessionFile, projectRoot, projectName, nowStr string,
	git gitStateJSON,
	plans plansStateJSON,
) {
	if isSymlink(sessionFile) {
		return
	}

	resumeMode := false
	var existing sessionStateJSON

	// 既存セッションを読み込む
	if data, err := os.ReadFile(sessionFile); err == nil {
		if json.Unmarshal(data, &existing) == nil {
			// ended_at が設定されていないなら resume モード
			// ここでは EventSeq > 0 かつ State が active 系なら resume
			if existing.SessionID != "" && existing.State != "stopped" && existing.State != "completed" && existing.State != "failed" {
				resumeMode = true
			}
		}
	}

	forkMode := os.Getenv("HARNESS_SESSION_FORK") == "true"
	if forkMode {
		resumeMode = false
	}

	var sess sessionStateJSON

	if resumeMode {
		// 既存セッションを更新
		existing.CWD = projectRoot
		existing.ProjectName = projectName
		existing.UpdatedAt = nowStr
		existing.Git = git
		existing.Plans = plans
		existing.StateVersion = 1
		sess = existing
	} else {
		// 新規セッション
		sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())
		resumeToken := fmt.Sprintf("resume-%d", time.Now().UnixNano())
		parentID := interface{}(nil)
		if forkMode && existing.SessionID != "" {
			parentID = existing.SessionID
		}

		sess = sessionStateJSON{
			SessionID:    sessionID,
			ParentID:     parentID,
			State:        "initialized",
			StateVersion: 1,
			StartedAt:    nowStr,
			UpdatedAt:    nowStr,
			ResumeToken:  resumeToken,
			EventSeq:     0,
			LastEventID:  "",
			ForkCount:    0,
			Orchestration: orchestrationJSON{
				MaxStateRetries:     3,
				RetryBackoffSeconds: 10,
			},
			CWD:          projectRoot,
			ProjectName:  projectName,
			PromptSeq:    0,
			Git:          git,
			Plans:        plans,
			ChangesThisSession: []interface{}{},
		}
	}

	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return
	}

	_ = writeFileAtomic(sessionFile, append(data, '\n'), 0600)
}

// generateToolingPolicy は tooling-policy.json を生成する。
// 重い外部コマンド依存（claude plugin list, MCP サーバー検索等）は避け、
// 基本的なスキャフォールドのみを生成する。
func (h *MonitorHandler) generateToolingPolicy(policyFile string) {
	if isSymlink(policyFile) {
		return
	}

	policy := toolingPolicyJSON{
		LSP: lspPolicyJSON{
			Available:         false,
			Plugins:           "",
			AvailableByExt:    map[string]bool{},
			LastUsedPromptSeq: 0,
			LastUsedToolName:  "",
			UsedSinceLastPrompt: false,
		},
		Plugins: pluginPolicyJSON{
			Installed:       nil,
			EnabledEstimate: nil,
			Source:          "",
		},
		MCP: mcpPolicyJSON{
			Configured:      nil,
			Disabled:        nil,
			EnabledEstimate: nil,
			Sources:         []string{},
		},
		Skills: skillsPolicyJSON{
			Index:            []interface{}{},
			DecisionRequired: false,
		},
	}

	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return
	}

	_ = writeFileAtomic(policyFile, append(data, '\n'), 0644)
}

// writeSummary はセッション状態サマリーを w に書き出す。
func (h *MonitorHandler) writeSummary(w io.Writer, projectName string, git gitStateJSON, plans plansStateJSON) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "セッション開始 - プロジェクト状態")
	fmt.Fprintln(w, strings.Repeat("─", 36))
	fmt.Fprintf(w, "プロジェクト: %s\n", projectName)
	fmt.Fprintf(w, "ブランチ: %s\n", git.Branch)

	if plans.Exists {
		total := plans.WIPTasks + plans.TODOTasks + plans.PendingTasks
		if total > 0 {
			fmt.Fprintf(w, "Plans.md: WIP %d件 / TODO %d件\n", plans.WIPTasks, plans.TODOTasks+plans.PendingTasks)
		}
	}

	fmt.Fprintln(w, strings.Repeat("─", 36))
	fmt.Fprintln(w, "")
}

// currentTime は現在時刻を返す。
func (h *MonitorHandler) currentTime() time.Time {
	if h.now != nil {
		return h.now()
	}
	return time.Now()
}

// formatInt は int をポインタで返す（tooling-policy の null 値用）。
func formatInt(v int) *int {
	return &v
}

// atoi は文字列を int に変換する。
func atoi(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}
