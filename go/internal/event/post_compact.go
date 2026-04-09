package event

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PostCompactHandler は PostCompact フックハンドラ。
// コンテキストコンパクション完了後に発火し、WIP タスクのコンテキストを再注入する。
//
// shell 版: scripts/hook-handlers/post-compact.sh
type PostCompactHandler struct {
	// StateDir はスナップショットファイルの場所を指定する。
	// 空の場合は ResolveStateDir(projectRoot) を使う。
	StateDir string
	// PlansFile は Plans.md のパスを指定する。
	// 空の場合は projectRoot/Plans.md を使う。
	PlansFile string
}

// precompactSnapshot は PreCompact が保存したスナップショット JSON のスキーマ。
// handoff-artifact.json (v2.0.0) の全フィールドをカバーする。
type precompactSnapshot struct {
	WIPTasks    []string `json:"wipTasks"`
	RecentEdits []string `json:"recentEdits"`
	// structured handoff 用フィールド
	PreviousState *handoffPreviousState  `json:"previous_state,omitempty"`
	NextAction    *handoffNextAction     `json:"next_action,omitempty"`
	OpenRisks     []handoffRisk          `json:"open_risks,omitempty"`
	FailedChecks  []handoffFailedCheck   `json:"failed_checks,omitempty"`
	DecisionLog   []handoffDecisionEntry `json:"decision_log,omitempty"`
	ContextReset  *handoffContextReset   `json:"context_reset,omitempty"`
	Continuity    *handoffContinuity     `json:"continuity,omitempty"`
}

// handoffPreviousState は previous_state フィールドのスキーマ。
type handoffPreviousState struct {
	Summary      string                  `json:"summary"`
	SessionState *handoffSessionState    `json:"session_state,omitempty"`
	PlanCounts   *handoffPlanCounts      `json:"plan_counts,omitempty"`
}

// handoffSessionState はセッション状態。
type handoffSessionState struct {
	State        string `json:"state"`
	ReviewStatus string `json:"review_status"`
	ActiveSkill  string `json:"active_skill"`
	ResumedAt    string `json:"resumed_at"`
}

// handoffPlanCounts はプランカウント情報。
type handoffPlanCounts struct {
	Total       int `json:"total"`
	WIP         int `json:"wip"`
	Blocked     int `json:"blocked"`
	RecentEdits int `json:"recent_edits"`
}

// handoffNextAction は next_action フィールドのスキーマ。
type handoffNextAction struct {
	Summary  string `json:"summary"`
	TaskID   string `json:"taskId"`
	Task     string `json:"task"`
	DoD      string `json:"dod"`
	Depends  string `json:"depends"`
	Status   string `json:"status"`
	Source   string `json:"source"`
	Priority string `json:"priority"`
}

// handoffRisk はリスクエントリ。
type handoffRisk struct {
	Severity string `json:"severity"`
	Kind     string `json:"kind"`
	Summary  string `json:"summary"`
	Detail   string `json:"detail"`
}

// handoffFailedCheck は失敗チェックエントリ。
type handoffFailedCheck struct {
	Source string `json:"source"`
	Check  string `json:"check"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// handoffDecisionEntry は決定ログエントリ。
type handoffDecisionEntry struct {
	Timestamp string `json:"timestamp"`
	Actor     string `json:"actor"`
	Decision  string `json:"decision"`
	Rationale string `json:"rationale"`
}

// handoffContextReset はコンテキストリセット推奨情報。
type handoffContextReset struct {
	Recommended bool   `json:"recommended"`
	Summary     string `json:"summary"`
}

// handoffContinuity は継続性コンテキスト。
type handoffContinuity struct {
	EffortHint  string `json:"effort_hint"`
	ActiveSkill string `json:"active_skill"`
	Summary     string `json:"summary"`
}

// compactionLogEntry は compaction-events.jsonl に書き出すエントリ。
type compactionLogEntry struct {
	Event       string `json:"event"`
	HasWIP      bool   `json:"has_wip"`
	HasSnapshot bool   `json:"has_snapshot"`
	HasHandoff  bool   `json:"has_handoff"`
	Timestamp   string `json:"timestamp"`
}

// Handle は stdin から PostCompact ペイロードを読み取り、
// WIP コンテキストを再注入した approve レスポンスを返す。
func (h *PostCompactHandler) Handle(r io.Reader, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil || len(data) == 0 {
		return WriteJSON(w, ApproveResponse{
			Decision: "approve",
			Reason:   "PostCompact: no payload",
		})
	}

	// プロジェクトルートを決定
	projectRoot := resolveProjectRoot(data)

	stateDir := h.StateDir
	if stateDir == "" {
		stateDir = ResolveStateDir(projectRoot)
	}

	plansFile := h.PlansFile
	if plansFile == "" {
		plansFile = filepath.Join(projectRoot, "Plans.md")
	}

	_ = EnsureStateDir(stateDir)

	compactionLog := filepath.Join(stateDir, "compaction-events.jsonl")
	precompactSnapshotPath := filepath.Join(stateDir, "precompact-snapshot.json")
	handoffArtifactPath := filepath.Join(stateDir, "handoff-artifact.json")

	// WIP タスクを Plans.md から取得
	wipSummary := h.getWIPSummary(plansFile)

	// structured handoff artifact の存在確認
	hasHandoff := fileExists(handoffArtifactPath)
	hasSnapshot := fileExists(precompactSnapshotPath)

	// イベントをログに記録
	entry := compactionLogEntry{
		Event:       "post_compact",
		HasWIP:      wipSummary != "",
		HasSnapshot: hasSnapshot,
		HasHandoff:  hasHandoff,
		Timestamp:   Now(),
	}
	h.appendCompactionLog(compactionLog, entry)

	// コンテキストメッセージを構築
	systemMsg := h.buildSystemMessage(
		wipSummary,
		precompactSnapshotPath,
		handoffArtifactPath,
		hasHandoff,
		hasSnapshot,
	)

	if systemMsg != "" {
		return WriteJSON(w, ApproveResponse{
			Decision:          "approve",
			Reason:            "PostCompact: WIP context re-injected via additionalContext",
			AdditionalContext: systemMsg,
		})
	}

	return WriteJSON(w, ApproveResponse{
		Decision: "approve",
		Reason:   "PostCompact: no WIP tasks to re-inject",
	})
}

// getWIPSummary は Plans.md から WIP/TODO タスクを抽出してサマリー文字列を返す。
func (h *PostCompactHandler) getWIPSummary(plansFile string) string {
	if !fileExists(plansFile) {
		return ""
	}

	f, err := os.Open(plansFile)
	if err != nil {
		return ""
	}
	defer f.Close()

	var wipLines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "cc:WIP") || strings.Contains(line, "cc:TODO") {
			wipLines = append(wipLines, line)
			if len(wipLines) >= 20 {
				break
			}
		}
	}

	return strings.Join(wipLines, "\n")
}

// buildSystemMessage はシステムメッセージを構築する。
// structured handoff artifact を優先し、なければ WIP サマリーを使う。
func (h *PostCompactHandler) buildSystemMessage(
	wipSummary,
	precompactSnapshotPath,
	handoffArtifactPath string,
	hasHandoff, hasSnapshot bool,
) string {
	// structured handoff artifact（優先）
	if hasHandoff {
		ctx := h.extractStructuredContext(handoffArtifactPath)
		if ctx != "" {
			return "[PostCompact Re-injection] Context was just compacted.\n" + ctx
		}
	}

	// precompact snapshot（次点）
	if hasSnapshot {
		ctx := h.extractPrecompactContext(precompactSnapshotPath)
		if ctx != "" {
			msg := "[PostCompact Re-injection] Context was just compacted. " + ctx
			if wipSummary != "" {
				msg += "\n\nActive WIP/TODO tasks in Plans.md:\n" + wipSummary
			}
			return msg
		}
	}

	// WIP サマリーのみ
	if wipSummary != "" {
		return "[PostCompact Re-injection] Context was just compacted. " +
			"The following WIP/TODO tasks are active in Plans.md:\n" + wipSummary
	}

	return ""
}

// extractStructuredContext は handoff artifact JSON から要点を抽出してテキストを返す。
// bash 版 post-compact.sh の get_structured_handoff_context に相当する詳細版。
func (h *PostCompactHandler) extractStructuredContext(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var snap precompactSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return ""
	}

	var parts []string
	parts = append(parts, "## Structured Handoff")

	// previous_state
	if snap.PreviousState != nil {
		if snap.PreviousState.Summary != "" {
			parts = append(parts, "- Previous state: "+snap.PreviousState.Summary)
		}
		if ss := snap.PreviousState.SessionState; ss != nil {
			var bits []string
			if ss.State != "" {
				bits = append(bits, "state="+ss.State)
			}
			if ss.ReviewStatus != "" {
				bits = append(bits, "review_status="+ss.ReviewStatus)
			}
			if ss.ActiveSkill != "" {
				bits = append(bits, "active_skill="+ss.ActiveSkill)
			}
			if ss.ResumedAt != "" {
				bits = append(bits, "resumed_at="+ss.ResumedAt)
			}
			if len(bits) > 0 {
				parts = append(parts, "- Session state: "+strings.Join(bits, ", "))
			}
		}
		if pc := snap.PreviousState.PlanCounts; pc != nil {
			var bits []string
			if pc.Total > 0 {
				bits = append(bits, fmt.Sprintf("total=%d", pc.Total))
			}
			if pc.WIP > 0 {
				bits = append(bits, fmt.Sprintf("wip=%d", pc.WIP))
			}
			if pc.Blocked > 0 {
				bits = append(bits, fmt.Sprintf("blocked=%d", pc.Blocked))
			}
			if pc.RecentEdits > 0 {
				bits = append(bits, fmt.Sprintf("recent_edits=%d", pc.RecentEdits))
			}
			if len(bits) > 0 {
				parts = append(parts, "- Plan counts: "+strings.Join(bits, ", "))
			}
		}
	}

	// next_action
	if na := snap.NextAction; na != nil {
		var naBits []string
		if na.Summary != "" {
			naBits = append(naBits, na.Summary)
		}
		taskLabel := strings.TrimSpace(na.TaskID + " " + na.Task)
		if taskLabel != "" && taskLabel != na.Summary {
			naBits = append(naBits, taskLabel)
		}
		if na.Depends != "" {
			naBits = append(naBits, "depends="+na.Depends)
		}
		if na.DoD != "" {
			naBits = append(naBits, "DoD="+na.DoD)
		}
		if len(naBits) > 0 {
			parts = append(parts, "- Next action: "+strings.Join(naBits, " | "))
		}
	}

	// open_risks（最大4件）
	if len(snap.OpenRisks) > 0 {
		risks := snap.OpenRisks
		if len(risks) > 4 {
			risks = risks[:4]
		}
		var riskTexts []string
		for _, r := range risks {
			text := riskNormalizeText(r.Summary, r.Detail)
			if text != "" {
				riskTexts = append(riskTexts, text)
			}
		}
		if len(riskTexts) > 0 {
			parts = append(parts, "- Open risks: "+strings.Join(riskTexts, "; "))
		}
	}

	// failed_checks（最大4件）
	if len(snap.FailedChecks) > 0 {
		checks := snap.FailedChecks
		if len(checks) > 4 {
			checks = checks[:4]
		}
		var checkTexts []string
		for _, c := range checks {
			text := riskNormalizeText(c.Check, c.Detail)
			if text != "" {
				checkTexts = append(checkTexts, text)
			}
		}
		if len(checkTexts) > 0 {
			parts = append(parts, "- Failed checks: "+strings.Join(checkTexts, "; "))
		}
	}

	// decision_log（最大2件）
	if len(snap.DecisionLog) > 0 {
		entries := snap.DecisionLog
		if len(entries) > 2 {
			entries = entries[:2]
		}
		var logTexts []string
		for _, e := range entries {
			text := riskNormalizeText(e.Decision, e.Rationale)
			if text != "" {
				logTexts = append(logTexts, text)
			}
		}
		if len(logTexts) > 0 {
			parts = append(parts, "- Decision log: "+strings.Join(logTexts, "; "))
		}
	}

	// context_reset
	if cr := snap.ContextReset; cr != nil && cr.Summary != "" {
		parts = append(parts, "- Context reset: "+cr.Summary)
	}

	// continuity
	if c := snap.Continuity; c != nil && c.Summary != "" {
		parts = append(parts, "- Continuity: "+c.Summary)
	}

	// WIP tasks（最大5件）
	if len(snap.WIPTasks) > 0 {
		wip := snap.WIPTasks
		if len(wip) > 5 {
			wip = wip[:5]
		}
		parts = append(parts, "- WIP tasks: "+strings.Join(wip, "; "))
	}

	// recent edits（最大5件）
	if len(snap.RecentEdits) > 0 {
		edits := snap.RecentEdits
		if len(edits) > 5 {
			edits = edits[:5]
		}
		parts = append(parts, "- Recent edits: "+strings.Join(edits, ", "))
	}

	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts, "\n")
}

// riskNormalizeText はリスク/チェックのテキストを正規化する。
func riskNormalizeText(primary, secondary string) string {
	if primary != "" {
		return primary
	}
	return secondary
}

// extractPrecompactContext は precompact snapshot JSON から WIP タスクと最近の編集を抽出する。
func (h *PostCompactHandler) extractPrecompactContext(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var snap precompactSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return ""
	}

	var parts []string
	if len(snap.WIPTasks) > 0 {
		parts = append(parts, "Pre-compaction WIP tasks: "+strings.Join(snap.WIPTasks, ", "))
	}
	if len(snap.RecentEdits) > 0 {
		edits := snap.RecentEdits
		if len(edits) > 10 {
			edits = edits[:10]
		}
		parts = append(parts, "Recent edits: "+strings.Join(edits, ", "))
	}
	return strings.Join(parts, ". ")
}

// appendCompactionLog はコンパクションログに 1 エントリ追記する。
func (h *PostCompactHandler) appendCompactionLog(path string, entry compactionLogEntry) {
	if isSymlink(path) {
		return
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintf(f, "%s\n", data)

	RotateJSONL(path)
}

// fileExists はファイルが存在するかどうかを返す。
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
