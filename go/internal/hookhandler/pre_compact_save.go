// Package hookhandler implements Go ports of the bash hook handler scripts.
package hookhandler

import (
	"bufio"
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

// PreCompactSave は pre-compact-save.js の Go 移植。
// PreCompact フックで handoff-artifact.json と precompact-snapshot.json を生成する。
//
// 元: scripts/hook-handlers/pre-compact-save.js
type PreCompactSave struct {
	// RepoRoot はリポジトリルートを指定する。
	// 空の場合は cwd から自動検出する。
	RepoRoot string
	// StateDir はスナップショットファイルの場所を指定する。
	// 空の場合は RepoRoot/.claude/state を使う。
	StateDir string
	// PlansFile は Plans.md のパスを指定する。
	// 空の場合は RepoRoot/Plans.md を使う。
	PlansFile string
	// Now は現在時刻を返す関数（テスト用に差し替え可能）。
	Now func() string
}

// artifactVersion はハンドオフアーティファクトのバージョン。
const artifactVersion = "2.0.0"

// legacySnapshotVersion はレガシースナップショットのバージョン。
const legacySnapshotVersion = "1.0.0"

// gitTimeoutSec は git コマンドのタイムアウト秒数。
const gitTimeoutSec = 5

// planRow は Plans.md の1行分のパース結果。
type planRow struct {
	TaskID  string   `json:"taskId"`
	Title   string   `json:"title"`
	DoD     string   `json:"dod"`
	Depends string   `json:"depends"`
	Status  string   `json:"status"`
	Tags    planTags `json:"tags"`
}

// planTags は planRow のタグ情報。
type planTags struct {
	Todo    bool `json:"todo"`
	Wip     bool `json:"wip"`
	Blocked bool `json:"blocked"`
}

// openRisk はリスクエントリ。
type openRisk struct {
	Severity string `json:"severity"`
	Kind     string `json:"kind"`
	Summary  string `json:"summary"`
	Detail   string `json:"detail"`
}

// failedCheck は失敗チェックエントリ。
type failedCheck struct {
	Source string `json:"source"`
	Check  string `json:"check"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// decisionLogEntry は決定ログエントリ。
type decisionLogEntry struct {
	Timestamp string `json:"timestamp"`
	Actor     string `json:"actor"`
	Decision  string `json:"decision"`
	Rationale string `json:"rationale"`
}

// contextResetPolicy はコンテキストリセットポリシー。
type contextResetPolicy struct {
	Mode   string                    `json:"mode"`
	DryRun bool                      `json:"dryRun"`
	Thresholds contextResetThresholds `json:"thresholds"`
}

// contextResetThresholds はコンテキストリセット閾値。
type contextResetThresholds struct {
	WIPTasks          int `json:"wipTasks"`
	BlockedTasks      int `json:"blockedTasks"`
	RecentEdits       int `json:"recentEdits"`
	FailedChecks      int `json:"failedChecks"`
	SessionAgeMinutes int `json:"sessionAgeMinutes"`
}

// contextResetCandidate はリセット判定候補。
type contextResetCandidate struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Actual    int    `json:"actual"`
	Threshold int    `json:"threshold"`
	Triggered bool   `json:"triggered"`
}

// contextResetCounters はリセット判定カウンター。
type contextResetCounters struct {
	WIPTasks          int  `json:"wipTasks"`
	BlockedTasks      int  `json:"blockedTasks"`
	RecentEdits       int  `json:"recentEdits"`
	FailedChecks      int  `json:"failedChecks"`
	SessionAgeMinutes *int `json:"sessionAgeMinutes"`
}

// contextResetRecommendation はコンテキストリセット推奨情報。
type contextResetRecommendation struct {
	Policy      contextResetPolicy      `json:"policy"`
	Recommended bool                    `json:"recommended"`
	Summary     string                  `json:"summary"`
	Reasons     []string                `json:"reasons"`
	Candidates  []contextResetCandidate `json:"candidates"`
	Counters    contextResetCounters    `json:"counters"`
}

// continuityCTX は継続性コンテキスト。
type continuityCTX struct {
	PluginFirstWorkflow         bool   `json:"plugin_first_workflow"`
	ResumeAwareEffortContinuity bool   `json:"resume_aware_effort_continuity"`
	EffortHint                  string `json:"effort_hint"`
	ActiveSkill                 string `json:"active_skill,omitempty"`
	Summary                     string `json:"summary"`
}

// planCounts はプランカウント情報。
type planCounts struct {
	Total       int `json:"total"`
	WIP         int `json:"wip"`
	Blocked     int `json:"blocked"`
	RecentEdits int `json:"recent_edits"`
}

// sessionStateSnapshot はセッション状態スナップショット。
type sessionStateSnapshot struct {
	State        string `json:"state,omitempty"`
	ResumedAt    string `json:"resumed_at,omitempty"`
	ActiveSkill  string `json:"active_skill,omitempty"`
	ReviewStatus string `json:"review_status,omitempty"`
}

// previousState は以前の状態情報。
type previousState struct {
	Summary      string                `json:"summary"`
	SessionState *sessionStateSnapshot `json:"session_state,omitempty"`
	PlanCounts   planCounts            `json:"plan_counts"`
}

// nextAction は次のアクション情報。
type nextAction struct {
	Summary  string `json:"summary"`
	TaskID   string `json:"taskId,omitempty"`
	Task     string `json:"task,omitempty"`
	DoD      string `json:"dod,omitempty"`
	Depends  string `json:"depends,omitempty"`
	Status   string `json:"status,omitempty"`
	Source   string `json:"source"`
	Priority string `json:"priority"`
}

// handoffArtifact はハンドオフアーティファクト全体。
type handoffArtifact struct {
	Version       string                     `json:"version"`
	LegacyVersion string                     `json:"legacy_version"`
	ArtifactType  string                     `json:"artifactType"`
	Timestamp     string                     `json:"timestamp"`
	SessionID     string                     `json:"sessionId"`
	PreviousState previousState              `json:"previous_state"`
	NextAction    nextAction                 `json:"next_action"`
	OpenRisks     []openRisk                 `json:"open_risks"`
	FailedChecks  []failedCheck              `json:"failed_checks"`
	DecisionLog   []decisionLogEntry         `json:"decision_log"`
	ContextReset  contextResetRecommendation `json:"context_reset"`
	Continuity    continuityCTX              `json:"continuity"`
	PlanItems     []planRow                  `json:"planItems"`
	WIPTasks      []string                   `json:"wipTasks"`
	RecentEdits   []string                   `json:"recentEdits"`
	Metrics       interface{}                `json:"metrics,omitempty"`
}

// preCompactResponse は PreCompact フックの出力。
type preCompactResponse struct {
	Continue bool   `json:"continue"`
	Message  string `json:"message"`
}

// sessionStateFile はセッション状態ファイルのスキーマ（最低限）。
type sessionStateFile struct {
	State       string `json:"state"`
	ResumedAt   string `json:"resumed_at"`
	ActiveSkill string `json:"active_skill"`
	StartedAt   string `json:"started_at"`
}

// workActiveFile は work-active.json のスキーマ（最低限）。
type workActiveFile map[string]interface{}

// sessionMetricsFile は session-metrics.json のスキーマ（最低限）。
type sessionMetricsFile map[string]interface{}

// Handle は stdin から PreCompact ペイロードを読み取り、
// handoff-artifact.json と precompact-snapshot.json を生成する。
func (h *PreCompactSave) Handle(r io.Reader, w io.Writer) error {
	now := h.getNow()
	sessionID := os.Getenv("CLAUDE_SESSION_ID")

	repoRoot := h.RepoRoot
	if repoRoot == "" {
		repoRoot = pcsFindRepoRoot()
	}

	stateDir := h.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(repoRoot, ".claude", "state")
	}

	plansFile := h.PlansFile
	if plansFile == "" {
		plansFile = filepath.Join(repoRoot, "Plans.md")
	}

	claudeDir := filepath.Join(repoRoot, ".claude")
	artifactPath := filepath.Join(stateDir, "handoff-artifact.json")
	snapshotPath := filepath.Join(stateDir, "precompact-snapshot.json")

	// セキュリティ: .claude がシンボリックリンクでないことを確認
	if info, err := os.Lstat(claudeDir); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return writePreCompactJSON(w, preCompactResponse{
			Continue: true,
			Message:  "Skipped: security check failed (.claude symlink)",
		})
	}

	// stateDir 作成・セキュリティチェック
	if err := h.ensureStateDir(stateDir); err != nil {
		return writePreCompactJSON(w, preCompactResponse{
			Continue: true,
			Message:  fmt.Sprintf("Skipped: %v", err),
		})
	}

	// シンボリックリンクチェック
	if isPreCompactSymlink(artifactPath) || isPreCompactSymlink(snapshotPath) {
		return writePreCompactJSON(w, preCompactResponse{
			Continue: true,
			Message:  "Skipped: artifact or snapshot is symlink",
		})
	}

	artifact := h.buildHandoffArtifact(repoRoot, plansFile, sessionID, now)

	// 書き出し
	if err := pcsWriteJSONFile(artifactPath, artifact); err != nil {
		return writePreCompactJSON(w, preCompactResponse{
			Continue: true,
			Message:  fmt.Sprintf("Error saving artifact: %v", err),
		})
	}

	// レガシースナップショット（互換性維持）
	snapshot := map[string]interface{}{
		"version":        legacySnapshotVersion,
		"legacy_version": artifact.LegacyVersion,
		"artifactType":   "precompact-snapshot",
		"timestamp":      artifact.Timestamp,
		"sessionId":      artifact.SessionID,
		"wipTasks":       artifact.WIPTasks,
		"recentEdits":    artifact.RecentEdits,
		"metrics":        artifact.Metrics,
		"context_reset":  artifact.ContextReset,
		"continuity":     artifact.Continuity,
	}
	if err := pcsWriteJSONFile(snapshotPath, snapshot); err != nil {
		// スナップショット失敗は警告のみ（アーティファクトは保存済み）
		_ = err
	}

	return writePreCompactJSON(w, preCompactResponse{
		Continue: true,
		Message: fmt.Sprintf(
			"Saved structured handoff artifact: %d WIP tasks, %d recent edits",
			len(artifact.WIPTasks), len(artifact.RecentEdits),
		),
	})
}

// getNow は現在時刻文字列を返す。
func (h *PreCompactSave) getNow() string {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now().UTC().Format(time.RFC3339)
}

// ensureStateDir はステートディレクトリを作成する。
func (h *PreCompactSave) ensureStateDir(stateDir string) error {
	if info, err := os.Lstat(stateDir); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("stateDir is a symlink")
		}
		_ = os.Chmod(stateDir, 0700)
		return nil
	}
	return os.MkdirAll(stateDir, 0700)
}

// buildHandoffArtifact はハンドオフアーティファクトを構築する。
func (h *PreCompactSave) buildHandoffArtifact(repoRoot, plansFile, sessionID, now string) handoffArtifact {
	planRows := h.getPlanRows(plansFile)
	wipTasks := getWIPTasks(planRows)
	recentEdits := h.getRecentEdits(repoRoot)
	metrics := h.getSessionMetrics(repoRoot)
	workState := h.getWorkState(repoRoot)
	sessionState := h.getSessionStateFile(repoRoot)
	na := h.pickNextAction(planRows)
	openRisks := h.buildOpenRisks(planRows, recentEdits, workState, metrics)
	failedChecks := h.buildFailedChecks(workState, metrics)
	decisionLog := h.buildDecisionLog(now, na, workState)
	contextReset := h.buildContextResetRecommendation(planRows, recentEdits, workState, metrics, sessionState)
	continuity := h.buildContinuityContext(sessionState, na)

	wipCount := countWIP(planRows)
	blockedCount := countBlocked(planRows)

	var summaryParts []string
	if wipCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d WIP", wipCount))
	}
	if blockedCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d blocked", blockedCount))
	}
	if len(recentEdits) > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d recent edit(s)", len(recentEdits)))
	}

	prevSummary := "Before compaction: no active WIP tasks detected"
	if len(summaryParts) > 0 {
		prevSummary = "Before compaction: " + strings.Join(summaryParts, ", ")
	}

	var sessSnap *sessionStateSnapshot
	if sessionState != nil {
		var reviewStatus string
		if workState != nil {
			reviewStatus = getStringField(workState, "review_status", "reviewStatus")
		}
		sessSnap = &sessionStateSnapshot{
			State:        sessionState.State,
			ResumedAt:    sessionState.ResumedAt,
			ActiveSkill:  sessionState.ActiveSkill,
			ReviewStatus: reviewStatus,
		}
	}

	naItem := nextAction{
		Summary:  "Re-read Plans.md and determine the next task",
		Source:   "fallback",
		Priority: "normal",
	}
	if na != nil {
		naItem = *na
	}

	return handoffArtifact{
		Version:       artifactVersion,
		LegacyVersion: legacySnapshotVersion,
		ArtifactType:  "structured-handoff",
		Timestamp:     now,
		SessionID:     sessionID,
		PreviousState: previousState{
			Summary:      prevSummary,
			SessionState: sessSnap,
			PlanCounts: planCounts{
				Total:       len(planRows),
				WIP:         wipCount,
				Blocked:     blockedCount,
				RecentEdits: len(recentEdits),
			},
		},
		NextAction:   naItem,
		OpenRisks:    openRisks,
		FailedChecks: failedChecks,
		DecisionLog:  decisionLog,
		ContextReset: contextReset,
		Continuity:   continuity,
		PlanItems:    planRows,
		WIPTasks:     wipTasks,
		RecentEdits:  recentEdits,
		Metrics:      metrics,
	}
}

// getPlanRows は Plans.md を読み込んでパースする。
func (h *PreCompactSave) getPlanRows(plansFile string) []planRow {
	f, err := os.Open(plansFile)
	if err != nil {
		return nil
	}
	defer f.Close()

	// cc:TODO / cc:WIP / cc:blocked の行を抽出
	reTodo := regexp.MustCompile("(?i)`?cc:TODO`?")
	reWip := regexp.MustCompile("(?i)`?cc:WIP`?|\\[in_progress\\]")
	reBlocked := regexp.MustCompile("(?i)`?cc:blocked`?|\\[blocked\\]")

	var rows []planRow
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "|") {
			continue
		}

		cells := splitPipeRow(line)
		if len(cells) < 5 {
			continue
		}

		taskID := strings.TrimSpace(cells[0])
		if taskID == "" || taskID == "Task" || strings.Contains(taskID, "---") {
			continue
		}

		status := strings.TrimSpace(cells[len(cells)-1])
		depends := strings.TrimSpace(cells[len(cells)-2])
		title := ""
		if len(cells) > 2 {
			title = strings.TrimSpace(cells[1])
		}
		dod := ""
		if len(cells) > 3 {
			dod = strings.TrimSpace(strings.Join(cells[2:len(cells)-2], "|"))
		}

		isTodo := reTodo.MatchString(status)
		isWip := reWip.MatchString(status)
		isBlocked := reBlocked.MatchString(status)

		if !isTodo && !isWip && !isBlocked {
			continue
		}

		rows = append(rows, planRow{
			TaskID:  taskID,
			Title:   title,
			DoD:     dod,
			Depends: depends,
			Status:  status,
			Tags: planTags{
				Todo:    isTodo,
				Wip:     isWip,
				Blocked: isBlocked,
			},
		})
	}
	return rows
}

// splitPipeRow はエスケープされた \| を考慮してパイプで行を分割する。
func splitPipeRow(line string) []string {
	const placeholder = "\x00PIPE\x00"
	escaped := strings.ReplaceAll(line, `\|`, placeholder)
	rawCells := strings.Split(escaped, "|")

	// 先頭・末尾の空セルを除去
	start := 0
	end := len(rawCells)
	if end > 0 && strings.TrimSpace(rawCells[0]) == "" {
		start = 1
	}
	if end > start && strings.TrimSpace(rawCells[end-1]) == "" {
		end--
	}
	cells := rawCells[start:end]

	for i, c := range cells {
		cells[i] = strings.ReplaceAll(c, placeholder, "|")
	}
	return cells
}

// getWIPTasks はプランROWのタイトルを返す。
func getWIPTasks(rows []planRow) []string {
	var titles []string
	for _, row := range rows {
		if row.Title != "" {
			titles = append(titles, row.Title)
		}
	}
	return titles
}

// getRecentEdits は git から最近変更されたファイルを取得する。
func (h *PreCompactSave) getRecentEdits(repoRoot string) []string {
	run := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = repoRoot
		out, err := cmd.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}

	staged := run("diff", "--name-only", "--cached")
	unstaged := run("diff", "--name-only")
	untracked := run("ls-files", "--others", "--exclude-standard")

	seen := map[string]bool{}
	var files []string
	for _, block := range []string{staged, unstaged, untracked} {
		if block == "" {
			continue
		}
		for _, f := range strings.Split(block, "\n") {
			if f == "" || seen[f] {
				continue
			}
			seen[f] = true
			files = append(files, f)
			if len(files) >= 20 {
				break
			}
		}
		if len(files) >= 20 {
			break
		}
	}
	return files
}

// getSessionMetrics はセッションメトリクスを読み込む。
func (h *PreCompactSave) getSessionMetrics(repoRoot string) interface{} {
	p := filepath.Join(repoRoot, ".claude", "state", "session-metrics.json")
	return pcsReadJSONFile(p)
}

// getWorkState は work-active.json を読み込む。
func (h *PreCompactSave) getWorkState(repoRoot string) interface{} {
	stateDir := filepath.Join(repoRoot, ".claude", "state")
	for _, name := range []string{"work-active.json", "ultrawork-active.json"} {
		if v := pcsReadJSONFile(filepath.Join(stateDir, name)); v != nil {
			return v
		}
	}
	return nil
}

// getSessionStateFile はセッション状態ファイルを読み込む。
func (h *PreCompactSave) getSessionStateFile(repoRoot string) *sessionStateFile {
	p := filepath.Join(repoRoot, ".claude", "state", "session.json")
	data, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var s sessionStateFile
	if err := json.Unmarshal(data, &s); err != nil {
		return nil
	}
	return &s
}

// pickNextAction は最優先の次アクションを選ぶ。
func (h *PreCompactSave) pickNextAction(rows []planRow) *nextAction {
	if len(rows) == 0 {
		return nil
	}
	// WIP > TODO > blocked の優先順
	var preferred *planRow
	for i := range rows {
		if rows[i].Tags.Wip {
			preferred = &rows[i]
			break
		}
	}
	if preferred == nil {
		for i := range rows {
			if rows[i].Tags.Todo {
				preferred = &rows[i]
				break
			}
		}
	}
	if preferred == nil {
		preferred = &rows[0]
	}

	priority := "normal"
	if preferred.Tags.Blocked {
		priority = "blocked"
	} else if preferred.Tags.Wip {
		priority = "high"
	}

	summary := strings.TrimSpace("Continue " + preferred.TaskID + " " + preferred.Title)

	return &nextAction{
		TaskID:   preferred.TaskID,
		Task:     preferred.Title,
		DoD:      preferred.DoD,
		Depends:  preferred.Depends,
		Status:   preferred.Status,
		Source:   "Plans.md",
		Priority: priority,
		Summary:  summary,
	}
}

// buildOpenRisks はリスクエントリを構築する。
//
// JS版 pre-compact-save.js の buildOpenRisks に相当。
// metrics 引数を受け取り、session-metrics.json の failed validations を quality risk として変換する。
func (h *PreCompactSave) buildOpenRisks(rows []planRow, recentEdits []string, workState interface{}, metrics interface{}) []openRisk {
	var risks []openRisk

	wipCount := countWIP(rows)
	blockedCount := countBlocked(rows)

	if wipCount > 0 {
		var details []string
		for _, row := range rows {
			if row.Tags.Wip {
				details = append(details, row.TaskID+" "+row.Title)
			}
			if len(details) >= 5 {
				break
			}
		}
		risks = append(risks, openRisk{
			Severity: "medium",
			Kind:     "continuity",
			Summary:  fmt.Sprintf("%d WIP task(s) remain in Plans.md", wipCount),
			Detail:   strings.Join(details, "; "),
		})
	}

	if blockedCount > 0 {
		var details []string
		for _, row := range rows {
			if row.Tags.Blocked {
				details = append(details, row.TaskID+" "+row.Title)
			}
			if len(details) >= 5 {
				break
			}
		}
		risks = append(risks, openRisk{
			Severity: "high",
			Kind:     "dependency",
			Summary:  fmt.Sprintf("%d blocked task(s) need attention before finish", blockedCount),
			Detail:   strings.Join(details, "; "),
		})
	}

	if len(recentEdits) > 0 {
		detail := strings.Join(recentEdits, ", ")
		if len(recentEdits) > 5 {
			detail = strings.Join(recentEdits[:5], ", ")
		}
		risks = append(risks, openRisk{
			Severity: "medium",
			Kind:     "verification",
			Summary:  fmt.Sprintf("%d recent edit(s) should be re-validated after resume", len(recentEdits)),
			Detail:   detail,
		})
	}

	if workState != nil {
		reviewStatus := getStringField(workState, "review_status", "reviewStatus")
		if reviewStatus == "failed" {
			risks = append(risks, openRisk{
				Severity: "high",
				Kind:     "review",
				Summary:  "work review_status is failed",
				Detail:   getStringField(workState, "last_failure", "failure_reason", "reason"),
			})
		} else if reviewStatus != "" {
			risks = append(risks, openRisk{
				Severity: "medium",
				Kind:     "review",
				Summary:  fmt.Sprintf("work review_status is %s", reviewStatus),
				Detail:   "Independent review is still required before finalizing the work.",
			})
		}
	}

	// session-metrics.json の failed validations → quality risk に変換（JS版との対称性）
	if metrics != nil {
		failureCount := countFailures(metrics)
		if failureCount > 0 {
			risks = append(risks, openRisk{
				Severity: "high",
				Kind:     "quality",
				Summary:  fmt.Sprintf("%d recorded failed check(s) in session metrics", failureCount),
				Detail:   "Review the latest validation results before resuming work.",
			})
		}
	}

	if len(rows) > 0 && len(risks) == 0 {
		risks = append(risks, openRisk{
			Severity: "low",
			Kind:     "continuity",
			Summary:  "Open plan items still exist and should be re-read after compaction",
			Detail:   fmt.Sprintf("%d plan row(s) captured from Plans.md", len(rows)),
		})
	}

	if len(risks) > 8 {
		risks = risks[:8]
	}
	return risks
}

// buildFailedChecks は失敗チェックエントリを構築する。
func (h *PreCompactSave) buildFailedChecks(workState, metrics interface{}) []failedCheck {
	var checks []failedCheck

	addFromSource := func(source string, value interface{}) {
		if value == nil {
			return
		}
		switch v := value.(type) {
		case []interface{}:
			for _, item := range v {
				switch entry := item.(type) {
				case string:
					checks = append(checks, failedCheck{Source: source, Check: entry, Status: "failed"})
				case map[string]interface{}:
					check := getStringField(entry, "check", "name", "type")
					if check == "" {
						check = "unknown"
					}
					checks = append(checks, failedCheck{
						Source: source,
						Check:  check,
						Status: getStringFieldDefault(entry, "failed", "status"),
						Detail: getStringField(entry, "detail", "message", "reason", "description"),
					})
				}
			}
		}
	}

	if workState != nil {
		wsMap, ok := workState.(map[string]interface{})
		if ok {
			addFromSource("work-active.json", firstNonNil(
				wsMap["failed_checks"], wsMap["failedChecks"],
				wsMap["failures"], wsMap["checks_failed"],
			))
			reviewStatus := getStringField(wsMap, "review_status", "reviewStatus")
			if reviewStatus == "failed" && len(checks) == 0 {
				checks = append(checks, failedCheck{
					Source: "work-active.json",
					Check:  "review_status",
					Status: "failed",
					Detail: getStringField(wsMap, "last_failure", "failure_reason"),
				})
			}
		}
	}

	if metrics != nil {
		mMap, ok := metrics.(map[string]interface{})
		if ok {
			addFromSource("session-metrics.json", firstNonNil(
				mMap["failed_checks"], mMap["failedChecks"], mMap["failures"],
			))
		}
	}

	if len(checks) > 8 {
		checks = checks[:8]
	}
	return checks
}

// buildDecisionLog は決定ログを構築する。
func (h *PreCompactSave) buildDecisionLog(now string, na *nextAction, workState interface{}) []decisionLogEntry {
	entries := []decisionLogEntry{
		{
			Timestamp: now,
			Actor:     "pre-compact-save",
			Decision:  "canonical_handoff_artifact_written",
			Rationale: "Persist a stable JSON artifact in .claude/state for long-running session handoff.",
		},
		{
			Timestamp: now,
			Actor:     "pre-compact-save",
			Decision:  "legacy_snapshot_mirrored",
			Rationale: "Keep precompact-snapshot.json for backward compatibility with older hooks.",
		},
	}

	if na != nil {
		entries = append(entries, decisionLogEntry{
			Timestamp: now,
			Actor:     "pre-compact-save",
			Decision:  "next_action_selected",
			Rationale: na.Summary + func() string {
				if na.Source != "" {
					return " (source: " + na.Source + ")"
				}
				return ""
			}(),
		})
	}

	if workState != nil {
		reviewStatus := getStringField(workState, "review_status", "reviewStatus")
		if reviewStatus != "" {
			entries = append(entries, decisionLogEntry{
				Timestamp: now,
				Actor:     "pre-compact-save",
				Decision:  "active_work_status_captured",
				Rationale: "work review_status=" + reviewStatus,
			})
		}
	}

	if len(entries) > 6 {
		entries = entries[:6]
	}
	return entries
}

// buildContextResetRecommendation はコンテキストリセット推奨を構築する。
func (h *PreCompactSave) buildContextResetRecommendation(
	rows []planRow, recentEdits []string,
	workState, metrics interface{},
	sessionState *sessionStateFile,
) contextResetRecommendation {
	policy := getContextResetPolicy()

	wipCount := countWIP(rows)
	blockedCount := countBlocked(rows)

	failureCount := 0
	if workState != nil {
		failureCount = countFailures(workState)
	}
	if metrics != nil && failureCount == 0 {
		failureCount = countFailures(metrics)
	}

	var sessionAgeMinutes *int
	if sessionState != nil && sessionState.StartedAt != "" {
		startedAt, err := time.Parse(time.RFC3339, sessionState.StartedAt)
		if err == nil {
			age := int(time.Since(startedAt).Minutes())
			if age < 0 {
				age = 0
			}
			sessionAgeMinutes = &age
		}
	}

	ageMinutes := 0
	if sessionAgeMinutes != nil {
		ageMinutes = *sessionAgeMinutes
	}

	type candidate struct {
		key, label string
		actual, threshold int
	}
	candidates := []candidate{
		{"wip_tasks", "WIP task count", wipCount, policy.Thresholds.WIPTasks},
		{"blocked_tasks", "blocked task count", blockedCount, policy.Thresholds.BlockedTasks},
		{"recent_edits", "recent edit count", len(recentEdits), policy.Thresholds.RecentEdits},
		{"failed_checks", "failed check count", failureCount, policy.Thresholds.FailedChecks},
		{"session_age_minutes", "session age (minutes)", ageMinutes, policy.Thresholds.SessionAgeMinutes},
	}

	var reasons []string
	var candidateResults []contextResetCandidate
	for _, c := range candidates {
		triggered := c.actual >= c.threshold
		if triggered {
			reasons = append(reasons, fmt.Sprintf("%d %s exceed threshold %d", c.actual, c.label, c.threshold))
		}
		candidateResults = append(candidateResults, contextResetCandidate{
			Key:       c.key,
			Label:     c.label,
			Actual:    c.actual,
			Threshold: c.threshold,
			Triggered: triggered,
		})
	}

	recommended := len(reasons) > 0
	modeSuffix := policy.Mode
	if policy.DryRun {
		modeSuffix += ", dry-run"
	}

	var summary string
	if recommended {
		reasonStr := strings.Join(reasons, "; ")
		if len(reasons) > 4 {
			reasonStr = strings.Join(reasons[:4], "; ")
		}
		summary = fmt.Sprintf("Context reset recommended (%s): %s", modeSuffix, reasonStr)
	} else {
		summary = fmt.Sprintf("Context reset not required (%s)", modeSuffix)
	}

	return contextResetRecommendation{
		Policy:      policy,
		Recommended: recommended,
		Summary:     summary,
		Reasons:     reasons,
		Candidates:  candidateResults,
		Counters: contextResetCounters{
			WIPTasks:          wipCount,
			BlockedTasks:      blockedCount,
			RecentEdits:       len(recentEdits),
			FailedChecks:      failureCount,
			SessionAgeMinutes: sessionAgeMinutes,
		},
	}
}

// buildContinuityContext は継続性コンテキストを構築する。
func (h *PreCompactSave) buildContinuityContext(sessionState *sessionStateFile, na *nextAction) continuityCTX {
	effortHint := os.Getenv("HARNESS_EFFORT_DEFAULT")
	if effortHint == "" {
		effortHint = "medium"
	}

	var activeSkill string
	if sessionState != nil {
		activeSkill = sessionState.ActiveSkill
	}

	var summaryParts []string
	summaryParts = append(summaryParts, "plugin-first workflow: enabled")
	summaryParts = append(summaryParts, "resume-aware effort continuity: "+effortHint)
	if activeSkill != "" {
		summaryParts = append(summaryParts, "active_skill="+activeSkill)
	}
	if na != nil && na.TaskID != "" {
		summaryParts = append(summaryParts, "next_task="+na.TaskID)
	}

	return continuityCTX{
		PluginFirstWorkflow:         true,
		ResumeAwareEffortContinuity: true,
		EffortHint:                  effortHint,
		ActiveSkill:                 activeSkill,
		Summary:                     strings.Join(summaryParts, "; "),
	}
}

// getContextResetPolicy は環境変数からポリシーを読み込む。
func getContextResetPolicy() contextResetPolicy {
	mode := os.Getenv("HARNESS_CONTEXT_RESET_MODE")
	if mode == "" {
		mode = "auto"
	}
	dryRunStr := os.Getenv("HARNESS_CONTEXT_RESET_DRY_RUN")
	dryRun := dryRunStr == "1" || strings.EqualFold(dryRunStr, "true") ||
		strings.EqualFold(dryRunStr, "yes") || strings.EqualFold(dryRunStr, "on")

	return contextResetPolicy{
		Mode:   mode,
		DryRun: dryRun,
		Thresholds: contextResetThresholds{
			WIPTasks:          parseEnvInt("HARNESS_CONTEXT_RESET_WIP_THRESHOLD", 4),
			BlockedTasks:      parseEnvInt("HARNESS_CONTEXT_RESET_BLOCKED_THRESHOLD", 1),
			RecentEdits:       parseEnvInt("HARNESS_CONTEXT_RESET_RECENT_EDITS_THRESHOLD", 8),
			FailedChecks:      parseEnvInt("HARNESS_CONTEXT_RESET_FAILED_CHECKS_THRESHOLD", 1),
			SessionAgeMinutes: parseEnvInt("HARNESS_CONTEXT_RESET_AGE_MINUTES", 120),
		},
	}
}

// parseEnvInt は環境変数を int として読み込む（失敗時はデフォルト値を返す）。
func parseEnvInt(key string, defaultVal int) int {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil || n <= 0 {
		return defaultVal
	}
	return n
}

// countWIP は WIP タスク数を返す。
func countWIP(rows []planRow) int {
	n := 0
	for _, r := range rows {
		if r.Tags.Wip {
			n++
		}
	}
	return n
}

// countBlocked はブロックされたタスク数を返す。
func countBlocked(rows []planRow) int {
	n := 0
	for _, r := range rows {
		if r.Tags.Blocked {
			n++
		}
	}
	return n
}

// countFailures は失敗数を取得する。
func countFailures(v interface{}) int {
	m, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}
	for _, key := range []string{"failed_checks", "failedChecks", "failures"} {
		if val, ok := m[key]; ok {
			if arr, ok := val.([]interface{}); ok {
				return len(arr)
			}
		}
	}
	for _, key := range []string{"failure_count", "failed_count"} {
		if val, ok := m[key]; ok {
			switch n := val.(type) {
			case float64:
				return int(n)
			case int:
				return n
			}
		}
	}
	return 0
}

// getStringField はマップから最初に見つかった文字列フィールドを返す。
func getStringField(v interface{}, keys ...string) string {
	m, ok := v.(map[string]interface{})
	if !ok {
		return ""
	}
	for _, k := range keys {
		if val, ok := m[k]; ok {
			if s, ok := val.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

// getStringFieldDefault はマップから文字列フィールドを返す。見つからない場合はデフォルト値を返す。
func getStringFieldDefault(m map[string]interface{}, defaultVal string, keys ...string) string {
	for _, k := range keys {
		if val, ok := m[k]; ok {
			if s, ok := val.(string); ok && s != "" {
				return s
			}
		}
	}
	return defaultVal
}

// firstNonNil は最初の非 nil 値を返す。
func firstNonNil(vals ...interface{}) interface{} {
	for _, v := range vals {
		if v != nil {
			return v
		}
	}
	return nil
}

// pcsFindRepoRoot は cwd から .git を探してリポジトリルートを返す。
func pcsFindRepoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	cwd, _ := os.Getwd()
	return cwd
}

// pcsReadJSONFile は JSON ファイルを読み込む（失敗時は nil を返す）。
func pcsReadJSONFile(path string) interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil
	}
	return v
}

// pcsWriteJSONFile は v を JSON ファイルとして書き出す（perm 0600）。
func pcsWriteJSONFile(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// writePreCompactJSON は PreCompact 用 JSON を w に書き出す。
func writePreCompactJSON(w io.Writer, resp preCompactResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

// isPreCompactSymlink はパスがシンボリックリンクかどうかを返す。
func isPreCompactSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}
