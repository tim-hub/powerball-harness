package session

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SummaryHandler は Stop フックハンドラ（セッション終了サマリー）。
// セッション終了時に session-log.md へ追記し、session.json に終了情報を記録する。
//
// shell 版: scripts/session-summary.sh
type SummaryHandler struct {
	// StateDir はステートディレクトリのパス。空の場合は cwd から推定する。
	StateDir string
	// MemoryDir は .claude/memory のパス。空の場合は projectRoot/.claude/memory を使う。
	MemoryDir string
	// PlansFile は Plans.md のパス。空の場合は projectRoot/Plans.md を使う。
	PlansFile string
	// now は現在時刻の注入関数（テスト用）。nil の場合は time.Now() を使う。
	now func() time.Time
}

// summaryInput は Stop フックの stdin JSON。
type summaryInput struct {
	CWD       string `json:"cwd,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

// summarySessionData は session.json から読み取るデータ。
type summarySessionData struct {
	SessionID     string  `json:"session_id"`
	State         string  `json:"state"`
	StartedAt     string  `json:"started_at"`
	ProjectName   string  `json:"project_name"`
	GitBranch     string  `json:"-"` // nested
	MemoryLogged  bool    `json:"memory_logged"`
	EventSeq      int     `json:"event_seq"`
	ChangesCount  int     `json:"-"` // 計算値

	// 生 JSON（フィールド追加用）
	raw map[string]interface{}
}

// Handle は stdin から Stop ペイロードを読み取り、
// セッション終了サマリーを session-log.md に書き出す。
func (h *SummaryHandler) Handle(r io.Reader, w io.Writer) error {
	data, _ := io.ReadAll(r)

	var inp summaryInput
	if len(data) > 0 {
		_ = json.Unmarshal(data, &inp)
	}

	projectRoot := resolveProjectRoot(inp.CWD)

	stateDir := h.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(projectRoot, ".claude", "state")
	}

	memoryDir := h.MemoryDir
	if memoryDir == "" {
		memoryDir = filepath.Join(projectRoot, ".claude", "memory")
	}

	plansFile := h.PlansFile
	if plansFile == "" {
		plansFile = filepath.Join(projectRoot, "Plans.md")
	}

	sessionFile := filepath.Join(stateDir, "session.json")
	sessionLogFile := filepath.Join(memoryDir, "session-log.md")
	eventLogFile := filepath.Join(stateDir, "session.events.jsonl")
	archiveDir := filepath.Join(stateDir, "sessions")

	now := h.currentTime()
	nowStr := now.UTC().Format(time.RFC3339)

	// session.json が存在しない場合はスキップ
	if _, err := os.Stat(sessionFile); err != nil {
		return nil
	}

	// session.json を読み取る
	sessData := h.readSessionData(sessionFile)

	// 二重実行防止（memory_logged が true の場合はスキップ）
	if sessData.MemoryLogged {
		return nil
	}

	// セッション時間を計算
	durationMinutes := h.calcDurationMinutes(sessData.StartedAt, now)

	// Plans.md の WIP タスクを取得
	wipTasks := h.readWIPTasks(plansFile)

	// 変更ファイル情報を取得
	changedFiles, importantFiles := h.readChangedFiles(sessData.raw)

	// session-log.md を作成（存在しない場合）
	if err := h.ensureSessionLog(sessionLogFile); err != nil {
		_ = err // エラーは無視して継続
	}

	// session-log.md に追記
	if sessData.StartedAt != "" && sessData.StartedAt != "null" {
		_ = h.appendSessionLog(sessionLogFile, sessionLogEntry{
			SessionID:      sessData.SessionID,
			ProjectName:    sessData.ProjectName,
			GitBranch:      h.readGitBranchFromSession(sessData.raw),
			StartedAt:      sessData.StartedAt,
			EndedAt:        nowStr,
			DurationMinutes: durationMinutes,
			ChangedFiles:   changedFiles,
			ImportantFiles: importantFiles,
			WIPTasks:       wipTasks,
		})
	}

	// session.stop イベントをログに記録
	h.appendEvent(eventLogFile, sessionFile, "session.stop", "stopped", nowStr)

	// session.json を更新（終了情報を記録）
	h.finalizeSessionFile(sessionFile, nowStr, durationMinutes)

	// アーカイブ保存
	h.archiveSession(sessionFile, eventLogFile, archiveDir, sessData.SessionID)

	// stdout にサマリーを出力（変更がある場合のみ）
	if len(changedFiles) > 0 || len(wipTasks) > 0 {
		h.writeSummaryOutput(w, sessData, durationMinutes, changedFiles, wipTasks)
	}

	_ = w
	return nil
}

// sessionLogEntry は session-log.md に書き出すエントリ。
type sessionLogEntry struct {
	SessionID       string
	ProjectName     string
	GitBranch       string
	StartedAt       string
	EndedAt         string
	DurationMinutes int
	ChangedFiles    []string
	ImportantFiles  []string
	WIPTasks        []string
}

// readSessionData は session.json から必要な情報を読み取る。
func (h *SummaryHandler) readSessionData(sessionFile string) summarySessionData {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return summarySessionData{}
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return summarySessionData{}
	}

	sess := summarySessionData{raw: raw}
	sess.SessionID = stringField(raw, "session_id", "unknown")
	sess.State = stringField(raw, "state", "")
	sess.StartedAt = stringField(raw, "started_at", "")
	sess.ProjectName = stringField(raw, "project_name", "")
	sess.MemoryLogged = boolField(raw, "memory_logged", false)
	sess.EventSeq = intField(raw, "event_seq", 0)

	// 変更ファイル数
	if changes, ok := raw["changes_this_session"].([]interface{}); ok {
		sess.ChangesCount = len(changes)
	}

	return sess
}

// readGitBranchFromSession は session.json の git.branch を読み取る。
func (h *SummaryHandler) readGitBranchFromSession(raw map[string]interface{}) string {
	if git, ok := raw["git"].(map[string]interface{}); ok {
		return stringField(git, "branch", "")
	}
	return ""
}

// readChangedFiles は session.json から変更ファイル一覧を読み取る。
func (h *SummaryHandler) readChangedFiles(raw map[string]interface{}) ([]string, []string) {
	changes, ok := raw["changes_this_session"].([]interface{})
	if !ok {
		return nil, nil
	}

	seen := map[string]bool{}
	var changedFiles, importantFiles []string

	for _, item := range changes {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		file := stringField(m, "file", "")
		if file == "" || seen[file] {
			continue
		}
		seen[file] = true
		changedFiles = append(changedFiles, file)
		if boolField(m, "important", false) {
			importantFiles = append(importantFiles, file)
		}
	}
	return changedFiles, importantFiles
}

// readWIPTasks は Plans.md から WIP/依頼中 タスクを読み取る。
func (h *SummaryHandler) readWIPTasks(plansFile string) []string {
	f, err := os.Open(plansFile)
	if err != nil {
		return nil
	}
	defer f.Close()

	var tasks []string
	scanner := newLineScanner(f)
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "cc:WIP") || strings.Contains(line, "pm:依頼中") || strings.Contains(line, "cursor:依頼中") {
			tasks = append(tasks, line)
			count++
			if count >= 20 {
				break
			}
		}
	}
	return tasks
}

// ensureSessionLog は session-log.md を作成する（存在しない場合）。
func (h *SummaryHandler) ensureSessionLog(logFile string) error {
	if _, err := os.Stat(logFile); err == nil {
		return nil // 既に存在する
	}

	if err := os.MkdirAll(filepath.Dir(logFile), 0700); err != nil {
		return err
	}

	header := `# Session Log

セッション単位の作業ログ（基本はローカル運用向け）。
重要な意思決定は ` + "`.claude/memory/decisions.md`" + `、再利用できる解法は ` + "`.claude/memory/patterns.md`" + ` に昇格してください。

## Index

- （必要に応じて追記）

---
`
	return os.WriteFile(logFile, []byte(header), 0644)
}

// appendSessionLog は session-log.md にセッション情報を追記する。
func (h *SummaryHandler) appendSessionLog(logFile string, entry sessionLogEntry) error {
	if isSymlink(logFile) {
		return fmt.Errorf("security: symlinked session log: %s", logFile)
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "\n## セッション: %s\n\n", entry.EndedAt)
	fmt.Fprintf(f, "- session_id: `%s`\n", entry.SessionID)
	if entry.ProjectName != "" {
		fmt.Fprintf(f, "- project: `%s`\n", entry.ProjectName)
	}
	if entry.GitBranch != "" {
		fmt.Fprintf(f, "- branch: `%s`\n", entry.GitBranch)
	}
	fmt.Fprintf(f, "- started_at: `%s`\n", entry.StartedAt)
	fmt.Fprintf(f, "- ended_at: `%s`\n", entry.EndedAt)
	if entry.DurationMinutes > 0 {
		fmt.Fprintf(f, "- duration_minutes: %d\n", entry.DurationMinutes)
	}
	fmt.Fprintf(f, "- changes: %d\n", len(entry.ChangedFiles))

	fmt.Fprintf(f, "\n### 変更ファイル\n")
	if len(entry.ChangedFiles) > 0 {
		for _, file := range entry.ChangedFiles {
			fmt.Fprintf(f, "- `%s`\n", file)
		}
	} else {
		fmt.Fprintln(f, "- （なし）")
	}

	fmt.Fprintf(f, "\n### 重要な変更（important=true）\n")
	if len(entry.ImportantFiles) > 0 {
		for _, file := range entry.ImportantFiles {
			fmt.Fprintf(f, "- `%s`\n", file)
		}
	} else {
		fmt.Fprintln(f, "- （なし）")
	}

	fmt.Fprintf(f, "\n### 次回への引き継ぎ（任意）\n")
	if len(entry.WIPTasks) > 0 {
		fmt.Fprintf(f, "\n**Plans.md のWIP/依頼中（抜粋）**:\n\n```\n")
		for _, task := range entry.WIPTasks {
			fmt.Fprintln(f, task)
		}
		fmt.Fprintf(f, "```\n")
	} else {
		fmt.Fprintln(f, "- （必要に応じて追記）")
	}

	fmt.Fprintln(f, "\n---")
	return nil
}

// appendEvent はイベントログに 1 エントリを追記し、session.json の EventSeq を更新する。
func (h *SummaryHandler) appendEvent(eventLogFile, sessionFile, eventType, state, ts string) {
	if isSymlink(eventLogFile) || isSymlink(sessionFile) {
		return
	}

	// session.json から EventSeq を読み取ってインクリメント
	seq := 0
	if data, err := os.ReadFile(sessionFile); err == nil {
		var raw map[string]interface{}
		if json.Unmarshal(data, &raw) == nil {
			seq = intField(raw, "event_seq", 0)
		}
	}
	seq++
	eventID := fmt.Sprintf("event-%06d", seq)

	entry := fmt.Sprintf(`{"id":%q,"type":%q,"ts":%q,"state":%q}`, eventID, eventType, ts, state)
	appendLine(eventLogFile, entry)
}

// finalizeSessionFile は session.json に終了情報を記録する。
func (h *SummaryHandler) finalizeSessionFile(sessionFile, endedAt string, durationMinutes int) {
	if isSymlink(sessionFile) {
		return
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}

	raw["ended_at"] = endedAt
	raw["duration_minutes"] = durationMinutes
	raw["memory_logged"] = true
	raw["state"] = "stopped"

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return
	}

	_ = writeFileAtomic(sessionFile, append(out, '\n'), 0600)
}

// archiveSession はセッションファイルをアーカイブディレクトリにコピーする。
func (h *SummaryHandler) archiveSession(sessionFile, eventLogFile, archiveDir, sessionID string) {
	if sessionID == "" {
		return
	}
	if err := os.MkdirAll(archiveDir, 0700); err != nil {
		return
	}

	archivePath := filepath.Join(archiveDir, sessionID+".json")
	if !isSymlink(archivePath) {
		_ = copyFile(sessionFile, archivePath, 0600)
	}

	archiveEvents := filepath.Join(archiveDir, sessionID+".events.jsonl")
	if !isSymlink(archiveEvents) {
		_ = copyFile(eventLogFile, archiveEvents, 0600)
	}
}

// writeSummaryOutput はセッション終了サマリーを出力する。
func (h *SummaryHandler) writeSummaryOutput(w io.Writer, sess summarySessionData, durationMinutes int, changedFiles, wipTasks []string) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "セッションサマリー")
	fmt.Fprintln(w, strings.Repeat("─", 32))

	if sess.ProjectName != "" {
		fmt.Fprintf(w, "プロジェクト: %s\n", sess.ProjectName)
	}
	fmt.Fprintf(w, "変更ファイル: %d件\n", len(changedFiles))
	if durationMinutes > 0 {
		fmt.Fprintf(w, "セッション時間: %d分\n", durationMinutes)
	}

	fmt.Fprintln(w, strings.Repeat("─", 32))
	fmt.Fprintln(w, "")
}

// calcDurationMinutes はセッション開始時刻から現在までの分数を計算する。
func (h *SummaryHandler) calcDurationMinutes(startedAt string, now time.Time) int {
	if startedAt == "" || startedAt == "null" {
		return 0
	}
	t, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return 0
	}
	return int(now.Sub(t).Minutes())
}

// currentTime は現在時刻を返す。
func (h *SummaryHandler) currentTime() time.Time {
	if h.now != nil {
		return h.now()
	}
	return time.Now()
}

// ---------------------------------------------------------------------------
// ヘルパー関数
// ---------------------------------------------------------------------------

// stringField は map から string フィールドを安全に取得する。
func stringField(m map[string]interface{}, key, defaultVal string) string {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}
	s, ok := v.(string)
	if !ok {
		return defaultVal
	}
	return s
}

// boolField は map から bool フィールドを安全に取得する。
func boolField(m map[string]interface{}, key string, defaultVal bool) bool {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}
	b, ok := v.(bool)
	if !ok {
		return defaultVal
	}
	return b
}

// intField は map から int フィールドを安全に取得する（JSON number は float64）。
func intField(m map[string]interface{}, key string, defaultVal int) int {
	v, ok := m[key]
	if !ok {
		return defaultVal
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	}
	return defaultVal
}

// appendLine は path にテキスト行を追記する。
func appendLine(path, line string) {
	if isSymlink(path) {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = fmt.Fprintln(f, line)
}

// copyFile はファイルをコピーする。
func copyFile(src, dst string, perm os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, perm)
}

// newLineScanner は bufio.Scanner のラッパー（テスト依存を避けるため）。
func newLineScanner(r io.Reader) *lineScanner {
	return &lineScanner{inner: &bufioScanner{r: r}}
}

type lineScanner struct {
	inner interface {
		Scan() bool
		Text() string
	}
}

func (ls *lineScanner) Scan() bool {
	return ls.inner.Scan()
}

func (ls *lineScanner) Text() string {
	return ls.inner.Text()
}

// bufioScanner は io.Reader をラップして行スキャンを提供する。
type bufioScanner struct {
	r    io.Reader
	buf  []byte
	pos  int
	end  int
	done bool
	cur  string
}

func (bs *bufioScanner) Scan() bool {
	if bs.done {
		return false
	}
	// 読み込みバッファを初期化
	if bs.buf == nil {
		data, err := io.ReadAll(bs.r)
		if err != nil {
			bs.done = true
			return false
		}
		bs.buf = data
	}
	if bs.pos >= len(bs.buf) {
		bs.done = true
		return false
	}
	// 次の改行を探す
	end := bs.pos
	for end < len(bs.buf) && bs.buf[end] != '\n' {
		end++
	}
	bs.cur = string(bs.buf[bs.pos:end])
	if end < len(bs.buf) {
		bs.pos = end + 1
	} else {
		bs.pos = end
		bs.done = true
	}
	return true
}

func (bs *bufioScanner) Text() string {
	return bs.cur
}
