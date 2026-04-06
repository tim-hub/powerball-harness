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

// precompactSnapshot は PreCompact が保存したスナップショット JSON のスキーマ（最低限）。
type precompactSnapshot struct {
	WIPTasks    []string `json:"wipTasks"`
	RecentEdits []string `json:"recentEdits"`
	// structured handoff 用フィールド（オプション）
	PreviousState interface{} `json:"previous_state,omitempty"`
	NextAction    interface{} `json:"next_action,omitempty"`
	OpenRisks     []interface{} `json:"open_risks,omitempty"`
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

	if len(snap.WIPTasks) > 0 {
		wip := snap.WIPTasks
		if len(wip) > 5 {
			wip = wip[:5]
		}
		parts = append(parts, "- WIP tasks: "+strings.Join(wip, "; "))
	}
	if len(snap.RecentEdits) > 0 {
		edits := snap.RecentEdits
		if len(edits) > 5 {
			edits = edits[:5]
		}
		parts = append(parts, "- Recent edits: "+strings.Join(edits, ", "))
	}
	if len(snap.OpenRisks) > 0 {
		parts = append(parts, fmt.Sprintf("- Open risks: %d items", len(snap.OpenRisks)))
	}

	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts, "\n")
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
