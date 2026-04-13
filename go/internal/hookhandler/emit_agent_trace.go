package hookhandler

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//
type EmitAgentTrace struct {
	RepoRoot string
	StateDir string
	MaxFileSize int64
	MaxGenerations int
	HTTPClient *http.Client
	Now func() string
}

const traceVersion = "0.3.0"

const eatMaxFileSizeDefault int64 = 10 * 1024 * 1024

const eatMaxGenerationsDefault = 3

const traceCacheTTL = 60 * time.Second

const vcsCacheTTL = 5 * time.Second

type traceRecord struct {
	Version     string                 `json:"version"`
	ID          string                 `json:"id"`
	Timestamp   string                 `json:"timestamp"`
	Tool        string                 `json:"tool"`
	Files       []traceFile            `json:"files"`
	VCS         *traceVCS              `json:"vcs,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	Attribution *traceAttribution      `json:"attribution,omitempty"`
	Metrics     *traceMetrics          `json:"metrics,omitempty"`
}

type traceFile struct {
	Path   string `json:"path"`
	Action string `json:"action"`
	Range  string `json:"range"`
}

type traceVCS struct {
	Revision string `json:"revision"`
	Branch   string `json:"branch"`
	Dirty    bool   `json:"dirty"`
}

type traceAttribution struct {
	Plugin  string `json:"plugin"`
	Version string `json:"version"`
	License string `json:"license,omitempty"`
	Author  string `json:"author,omitempty"`
}

type traceMetrics struct {
	TokenCount *int64   `json:"tokenCount,omitempty"`
	ToolUses   *int64   `json:"toolUses,omitempty"`
	Duration   *float64 `json:"duration,omitempty"`
}

type tracerCache struct {
	mu sync.Mutex

	projMeta     map[string]string
	projMetaTime time.Time

	vcsInfo  *traceVCS
	vcsTime  time.Time

	attr     *traceAttribution
	attrTime time.Time
}

var eatCache = &tracerCache{}

func (e *EmitAgentTrace) Handle(r io.Reader, w io.Writer) error {
	toolName := os.Getenv("CLAUDE_TOOL_NAME")
	toolInput := os.Getenv("CLAUDE_TOOL_INPUT")
	toolResult := os.Getenv("CLAUDE_TOOL_RESULT")
	sessionID := os.Getenv("CLAUDE_SESSION_ID")

	if !eatIsSupportedTool(toolName) {
		return nil
	}

	repoRoot := e.RepoRoot
	if repoRoot == "" {
		repoRoot = pcsFindRepoRoot()
	}

	files := e.parseToolInput(toolName, toolInput, repoRoot)
	if len(files) == 0 && toolName != "Task" {
		return nil
	}

	for i, f := range files {
		if filepath.IsAbs(f.Path) {
			rel, err := filepath.Rel(repoRoot, f.Path)
			if err == nil {
				files[i].Path = rel
			}
		}
	}

	rec := traceRecord{
		Version:   traceVersion,
		ID:        eatGenerateUUID(),
		Timestamp: e.getNow(),
		Tool:      toolName,
		Files:     files,
		Metadata:  make(map[string]interface{}),
	}

	if sessionID != "" {
		rec.Metadata["sessionId"] = sessionID
	}

	if vcs := e.getVCSInfo(); vcs != nil {
		rec.VCS = vcs
	}

	meta := e.getProjectMetadata(repoRoot)
	for k, v := range meta {
		rec.Metadata[k] = v
	}

	// Attribution
	if attr := e.getAttribution(); attr != nil {
		rec.Attribution = attr
	}

	if toolName == "Task" {
		if m := e.extractTaskMetrics(toolResult); m != nil {
			rec.Metrics = m
		}
		e.extractTaskMetadata(toolInput, &rec)
	}

	stateDir := e.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(repoRoot, ".claude", "state")
	}
	tracePath := filepath.Join(stateDir, "agent-trace.jsonl")

	if err := e.appendTrace(repoRoot, stateDir, tracePath, &rec); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[agent-trace] %v\n", err)
	}

	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		go e.emitOtelSpan(endpoint, &rec)
	}

	return nil
}

func (e *EmitAgentTrace) appendTrace(repoRoot, stateDir, tracePath string, rec *traceRecord) error {
	claudeDir := filepath.Join(repoRoot, ".claude")

	if info, err := os.Lstat(claudeDir); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf(".claude symlink detected, refusing to write trace")
	}

	if info, err := os.Lstat(stateDir); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("stateDir is symlink, refusing to write trace")
		}
		_ = os.Chmod(stateDir, 0700)
	} else {
		if err := os.MkdirAll(stateDir, 0700); err != nil {
			return fmt.Errorf("creating stateDir: %w", err)
		}
	}

	if info, err := os.Lstat(tracePath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("tracePath is symlink, refusing to write trace")
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("tracePath is not a regular file")
		}
	}

	e.rotateIfNeeded(tracePath)

	line, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshaling trace record: %w", err)
	}

	f, err := os.OpenFile(tracePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("opening trace file: %w", err)
	}
	defer f.Close()

	_ = f.Chmod(0600)

	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("opened fd is not a regular file")
	}

	if _, err := fmt.Fprintf(f, "%s\n", line); err != nil {
		return fmt.Errorf("writing trace: %w", err)
	}
	return nil
}

func (e *EmitAgentTrace) rotateIfNeeded(tracePath string) {
	maxSize := e.MaxFileSize
	if maxSize == 0 {
		maxSize = eatMaxFileSizeDefault
	}
	maxGen := e.MaxGenerations
	if maxGen == 0 {
		maxGen = eatMaxGenerationsDefault
	}

	info, err := os.Stat(tracePath)
	if err != nil || info.Size() < maxSize {
		return
	}

	lockPath := tracePath + ".lock"
	lockF, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	lockF.Close()
	defer os.Remove(lockPath)

	info, err = os.Stat(tracePath)
	if err != nil || info.Size() < maxSize {
		return
	}

	for i := maxGen - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", tracePath, i)
		newPath := fmt.Sprintf("%s.%d", tracePath, i+1)
		if _, err := os.Stat(oldPath); err == nil {
			if i == maxGen-1 {
				_ = os.Remove(oldPath)
			} else {
				_ = os.Rename(oldPath, newPath)
			}
		}
	}
	_ = os.Rename(tracePath, tracePath+".1")
}

func (e *EmitAgentTrace) parseToolInput(toolName, toolInput, repoRoot string) []traceFile {
	if toolInput == "" {
		return nil
	}

	var input map[string]interface{}
	if err := json.Unmarshal([]byte(toolInput), &input); err != nil {
		return nil
	}

	var files []traceFile

	switch toolName {
	case "Edit":
		if fp, ok := input["file_path"].(string); ok && fp != "" {
			if eatIsPathWithinRepo(fp, repoRoot) {
				files = append(files, traceFile{Path: fp, Action: "modify", Range: "unknown"})
			}
		}
	case "Write":
		if fp, ok := input["file_path"].(string); ok && fp != "" {
			if eatIsPathWithinRepo(fp, repoRoot) {
				files = append(files, traceFile{Path: fp, Action: "create", Range: "unknown"})
			}
		}
	case "MultiEdit":
		if fp, ok := input["file_path"].(string); ok && fp != "" {
			if eatIsPathWithinRepo(fp, repoRoot) {
				files = append(files, traceFile{Path: fp, Action: "modify", Range: "unknown"})
			}
		}
	}

	return files
}

func (e *EmitAgentTrace) extractTaskMetadata(toolInput string, rec *traceRecord) {
	if toolInput == "" {
		return
	}
	var input map[string]interface{}
	if err := json.Unmarshal([]byte(toolInput), &input); err != nil {
		return
	}
	if taskID, ok := input["task_id"].(string); ok && taskID != "" {
		rec.Metadata["taskId"] = taskID
	}
	if subagentType, ok := input["subagent_type"].(string); ok && subagentType != "" {
		rec.Metadata["subagentType"] = subagentType
		rec.Metadata["agentRole"] = eatNormalizeAgentRole(subagentType)
	} else if agentName, ok := input["agent_name"].(string); ok && agentName != "" {
		rec.Metadata["agentRole"] = eatNormalizeAgentRole(agentName)
	}
}

func (e *EmitAgentTrace) extractTaskMetrics(toolResult string) *traceMetrics {
	if toolResult == "" {
		return nil
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(toolResult), &result); err != nil {
		return nil
	}
	metricsRaw, ok := result["metrics"].(map[string]interface{})
	if !ok {
		return nil
	}

	m := &traceMetrics{}
	hasValue := false

	if tc, ok := metricsRaw["tokenCount"].(float64); ok {
		v := int64(tc)
		m.TokenCount = &v
		hasValue = true
	}
	if tu, ok := metricsRaw["toolUses"].(float64); ok {
		v := int64(tu)
		m.ToolUses = &v
		hasValue = true
	}
	if d, ok := metricsRaw["duration"].(float64); ok {
		m.Duration = &d
		hasValue = true
	}

	if !hasValue {
		return nil
	}
	return m
}

func (e *EmitAgentTrace) getVCSInfo() *traceVCS {
	eatCache.mu.Lock()
	defer eatCache.mu.Unlock()

	if eatCache.vcsInfo != nil && time.Since(eatCache.vcsTime) < vcsCacheTTL {
		return eatCache.vcsInfo
	}

	vcs := eatFetchVCSInfo()
	eatCache.vcsInfo = vcs
	eatCache.vcsTime = time.Now()
	return vcs
}

func eatFetchVCSInfo() *traceVCS {
	out, err := eatRunGitCmd("status", "--porcelain=2", "-b", "-uno")
	if err != nil || out == "" {
		return nil
	}

	var revision, branch string
	dirty := false

	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "# branch.oid ") {
			revision = strings.TrimPrefix(line, "# branch.oid ")
		} else if strings.HasPrefix(line, "# branch.head ") {
			branch = strings.TrimPrefix(line, "# branch.head ")
		} else if line != "" && !strings.HasPrefix(line, "#") {
			dirty = true
		}
	}

	if revision == "" || branch == "" {
		return nil
	}
	return &traceVCS{Revision: revision, Branch: branch, Dirty: dirty}
}

func (e *EmitAgentTrace) getProjectMetadata(repoRoot string) map[string]string {
	eatCache.mu.Lock()
	defer eatCache.mu.Unlock()

	if eatCache.projMeta != nil && time.Since(eatCache.projMetaTime) < traceCacheTTL {
		return eatCache.projMeta
	}

	meta := map[string]string{
		"project":     eatGetProjectName(repoRoot),
		"projectType": eatDetectProjectType(repoRoot),
	}
	eatCache.projMeta = meta
	eatCache.projMetaTime = time.Now()
	return meta
}

func (e *EmitAgentTrace) getAttribution() *traceAttribution {
	eatCache.mu.Lock()
	defer eatCache.mu.Unlock()

	if eatCache.attr != nil && time.Since(eatCache.attrTime) < traceCacheTTL {
		return eatCache.attr
	}

	attr := eatFetchAttribution()
	eatCache.attr = attr
	eatCache.attrTime = time.Now()
	return attr
}

func eatFetchAttribution() *traceAttribution {
	pluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	if pluginRoot == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(pluginRoot, "plugin.json"))
	if err != nil {
		return nil
	}
	var pkg map[string]interface{}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	attr := &traceAttribution{
		Plugin:  getString(pkg, "name", "unknown"),
		Version: getString(pkg, "version", "unknown"),
	}
	if lic := getString(pkg, "license", ""); lic != "" {
		attr.License = lic
	}
	if author := getString(pkg, "author", ""); author != "" {
		attr.Author = author
	}
	return attr
}

func (e *EmitAgentTrace) emitOtelSpan(otlpEndpoint string, rec *traceRecord) {
	serviceVersion := eatReadServiceVersion()
	spanJSON := eatBuildOtelSpanJSON(rec, serviceVersion)

	data, err := json.Marshal(spanJSON)
	if err != nil {
		return
	}

	url := strings.TrimRight(otlpEndpoint, "/") + "/v1/traces"

	client := e.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[agent-trace] otel export failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = fmt.Fprintf(os.Stderr, "[agent-trace] otel export HTTP %d -> %s\n", resp.StatusCode, url)
	}
}

func eatBuildOtelSpanJSON(rec *traceRecord, serviceVersion string) map[string]interface{} {
	endMs := int64(0)
	if t, err := time.Parse(time.RFC3339, rec.Timestamp); err == nil {
		endMs = t.UnixMilli()
	}
	endNano := fmt.Sprintf("%d000000", endMs)

	uuidHex := strings.ReplaceAll(rec.ID, "-", "")
	traceID := uuidHex
	spanID := uuidHex[:16]

	var attributes []map[string]interface{}

	if taskID, ok := rec.Metadata["taskId"].(string); ok && taskID != "" {
		attributes = append(attributes, map[string]interface{}{
			"key":   "task.id",
			"value": map[string]string{"stringValue": taskID},
		})
	}
	if agentRole, ok := rec.Metadata["agentRole"].(string); ok && agentRole != "" {
		attributes = append(attributes, map[string]interface{}{
			"key":   "agent.type",
			"value": map[string]string{"stringValue": agentRole},
		})
	}
	if effort, ok := rec.Metadata["effort"].(string); ok && effort != "" {
		attributes = append(attributes, map[string]interface{}{
			"key":   "effort",
			"value": map[string]string{"stringValue": effort},
		})
	}
	attributes = append(attributes, map[string]interface{}{
		"key":   "tool.name",
		"value": map[string]string{"stringValue": rec.Tool},
	})
	if rec.VCS != nil && rec.VCS.Branch != "" {
		attributes = append(attributes, map[string]interface{}{
			"key":   "vcs.branch",
			"value": map[string]string{"stringValue": rec.VCS.Branch},
		})
	}
	if sessionID, ok := rec.Metadata["sessionId"].(string); ok && sessionID != "" {
		attributes = append(attributes, map[string]interface{}{
			"key":   "session.id",
			"value": map[string]string{"stringValue": sessionID},
		})
	}

	agentRole, _ := rec.Metadata["agentRole"].(string)
	spanName := "harness." + strings.ToLower(rec.Tool)
	if agentRole != "" {
		spanName = "harness." + agentRole
	}

	return map[string]interface{}{
		"resourceSpans": []map[string]interface{}{
			{
				"resource": map[string]interface{}{
					"attributes": []map[string]interface{}{
						{"key": "service.name", "value": map[string]string{"stringValue": "claude-code-harness"}},
						{"key": "service.version", "value": map[string]string{"stringValue": serviceVersion}},
					},
				},
				"scopeSpans": []map[string]interface{}{
					{
						"scope": map[string]interface{}{"name": "harness.agent"},
						"spans": []map[string]interface{}{
							{
								"traceId":            traceID,
								"spanId":             spanID,
								"name":               spanName,
								"kind":               1,
								"startTimeUnixNano":  endNano,
								"endTimeUnixNano":    endNano,
								"attributes":         attributes,
							},
						},
					},
				},
			},
		},
	}
}

func eatReadServiceVersion() string {
	pluginRoot := os.Getenv("CLAUDE_PLUGIN_ROOT")
	if pluginRoot == "" {
		return "0.0.0"
	}

	if data, err := os.ReadFile(filepath.Join(pluginRoot, "plugin.json")); err == nil {
		var pkg map[string]interface{}
		if err := json.Unmarshal(data, &pkg); err == nil {
			if v, ok := pkg["version"].(string); ok && v != "" {
				return v
			}
		}
	}

	if data, err := os.ReadFile(filepath.Join(pluginRoot, "VERSION")); err == nil {
		return strings.TrimSpace(string(data))
	}

	return "0.0.0"
}

func eatGetProjectName(repoRoot string) string {
	pkgPath := filepath.Join(repoRoot, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg map[string]interface{}
		if err := json.Unmarshal(data, &pkg); err == nil {
			if name, ok := pkg["name"].(string); ok && name != "" {
				return name
			}
		}
	}
	return filepath.Base(repoRoot)
}

func eatDetectProjectType(repoRoot string) string {
	checks := [][2]string{
		{"next.config.js", "nextjs"},
		{"next.config.ts", "nextjs"},
		{"nuxt.config.js", "nuxt"},
		{"nuxt.config.ts", "nuxt"},
		{"svelte.config.js", "svelte"},
		{"astro.config.mjs", "astro"},
		{"Cargo.toml", "rust"},
		{"go.mod", "go"},
		{"pyproject.toml", "python"},
		{"setup.py", "python"},
		{"requirements.txt", "python"},
		{"Gemfile", "ruby"},
		{"composer.json", "php"},
		{"package.json", "node"},
	}
	for _, check := range checks {
		if _, err := os.Stat(filepath.Join(repoRoot, check[0])); err == nil {
			return check[1]
		}
	}
	return "unknown"
}

func eatNormalizeAgentRole(name string) string {
	v := strings.ToLower(strings.TrimSpace(name))
	if v == "" {
		return "unknown"
	}
	if strings.Contains(v, "review") {
		return "reviewer"
	}
	if strings.Contains(v, "lead") || strings.Contains(v, "planner") {
		return "lead"
	}
	if strings.Contains(v, "worker") || strings.Contains(v, "impl") {
		return "worker"
	}
	return v
}

func eatIsPathWithinRepo(filePath, repoRoot string) bool {
	if strings.Contains(filePath, "..") {
		return false
	}

	absPath := filePath
	if !filepath.IsAbs(filePath) {
		absPath = filepath.Join(repoRoot, filePath)
	}

	resolvedRepo, err := filepath.EvalSymlinks(repoRoot)
	if err != nil {
		resolvedRepo = repoRoot
	}

	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		parentDir := filepath.Dir(absPath)
		resolvedParent, err := filepath.EvalSymlinks(parentDir)
		if err != nil {
			resolvedPath = filepath.Clean(absPath)
		} else {
			resolvedPath = filepath.Join(resolvedParent, filepath.Base(absPath))
		}
	}

	return strings.HasPrefix(resolvedPath, resolvedRepo+string(filepath.Separator)) ||
		resolvedPath == resolvedRepo
}

func eatIsSupportedTool(toolName string) bool {
	return toolName == "Edit" || toolName == "Write" || toolName == "MultiEdit" || toolName == "Task"
}

func eatGenerateUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		ts := time.Now().UnixNano()
		for i := 0; i < 8; i++ {
			b[i] = byte(ts >> (i * 8))
		}
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func (e *EmitAgentTrace) getNow() string {
	if e.Now != nil {
		return e.Now()
	}
	return time.Now().UTC().Format(time.RFC3339)
}


func getString(m map[string]interface{}, key, defaultVal string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return defaultVal
}

func eatRunGitCmd(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
