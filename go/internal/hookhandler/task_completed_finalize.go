package hookhandler


import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type finalizeMarkerJSON struct {
	SessionID   string `json:"session_id"`
	Project     string `json:"project"`
	SummaryMode string `json:"summary_mode"`
	FinalizedAt string `json:"finalized_at"`
	Status      string `json:"status"`
}

func (h *taskCompletedHandler) maybeFinalizeHarnessMem(ts string) {
	sessionID := h.resolveSessionID()
	if sessionID == "" {
		return
	}

	if h.finalizeMarkerExistsForSession(sessionID) {
		return
	}

	projectName := h.resolveProjectName()
	if projectName == "" {
		projectName = lastPathComponent(h.projectRoot)
	}

	payload := map[string]string{
		"project":      projectName,
		"session_id":   sessionID,
		"summary_mode": "work_completed",
	}
	payloadData, err := json.Marshal(payload)
	if err != nil {
		return
	}

	baseURL := os.Getenv("HARNESS_MEM_BASE_URL")
	if baseURL == "" {
		port := os.Getenv("HARNESS_MEM_PORT")
		if port == "" {
			port = "37888"
		}
		baseURL = "http://localhost:" + port
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Post(
		baseURL+"/v1/sessions/finalize",
		"application/json",
		strings.NewReader(string(payloadData)),
	)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body) //nolint:errcheck

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		h.writeFinalizeMarker(sessionID, projectName, ts)
	}
}

func (h *taskCompletedHandler) resolveSessionID() string {
	if id := os.Getenv("SESSION_ID"); id != "" {
		return id
	}
	return h.resolveSessionStateField("session_id")
}

func (h *taskCompletedHandler) resolveProjectName() string {
	if name := os.Getenv("PROJECT_NAME"); name != "" {
		return name
	}
	if name := h.resolveSessionStateField("project_name"); name != "" {
		return name
	}
	return lastPathComponent(h.projectRoot)
}

func (h *taskCompletedHandler) resolveSessionStateField(field string) string {
	sessionPath := h.stateDir + "/session.json"
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	if v, ok := m[field]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (h *taskCompletedHandler) finalizeMarkerExistsForSession(sessionID string) bool {
	if info, err := os.Lstat(h.finalizeMarker); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	if info, err := os.Lstat(h.stateDir); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return false
	}

	data, err := os.ReadFile(h.finalizeMarker)
	if err != nil {
		return false
	}
	var marker finalizeMarkerJSON
	if err := json.Unmarshal(data, &marker); err != nil {
		return false
	}
	return marker.SessionID == sessionID &&
		marker.SummaryMode == "work_completed" &&
		marker.Status == "success"
}

func (h *taskCompletedHandler) writeFinalizeMarker(sessionID, projectName, ts string) {
	if isSymlink(h.stateDir) {
		fmt.Fprintf(os.Stderr, "[WARNING] writeFinalizeMarker: stateDir is a symlink (%s), refusing write\n", h.stateDir)
		return
	}
	if info, err := os.Lstat(h.finalizeMarker); err == nil && info.Mode()&os.ModeSymlink != 0 {
		fmt.Fprintf(os.Stderr, "[WARNING] writeFinalizeMarker: finalizeMarker is a symlink (%s), refusing write\n", h.finalizeMarker)
		return
	}

	marker := finalizeMarkerJSON{
		SessionID:   sessionID,
		Project:     projectName,
		SummaryMode: "work_completed",
		FinalizedAt: ts,
		Status:      "success",
	}
	data, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return
	}

	tmpPath := h.finalizeMarker + ".tmp"
	if err := os.WriteFile(tmpPath, append(data, '\n'), 0o644); err != nil {
		return
	}
	os.Rename(tmpPath, h.finalizeMarker) //nolint:errcheck
}

func (h *taskCompletedHandler) fireWebhook(rawPayload []byte) {
	webhookURL := os.Getenv("HARNESS_WEBHOOK_URL")
	if webhookURL == "" {
		return
	}

	body := rawPayload
	if len(body) == 0 {
		body = []byte("{}")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, webhookURL, strings.NewReader(string(body)))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[task-completed] webhook request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Harness-Event", "task-completed")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[task-completed] webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body) //nolint:errcheck
}

func lastPathComponent(path string) string {
	path = strings.TrimRight(path, "/")
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}
