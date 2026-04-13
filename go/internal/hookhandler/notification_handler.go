package hookhandler

// notification_handler.go
//

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type notificationInput struct {
	NotificationType string `json:"notification_type"`
	Type             string `json:"type"`
	Matcher          string `json:"matcher"`
	SessionID        string `json:"session_id"`
	AgentType        string `json:"agent_type"`
}

type notificationLogEntry struct {
	Event            string `json:"event"`
	NotificationType string `json:"notification_type"`
	SessionID        string `json:"session_id"`
	AgentType        string `json:"agent_type"`
	Timestamp        string `json:"timestamp"`
}

//
func HandleNotification(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}

	var input notificationInput
	if jsonErr := json.Unmarshal(data, &input); jsonErr != nil {
		return nil
	}

	notificationType := input.NotificationType
	if notificationType == "" {
		notificationType = input.Type
	}
	if notificationType == "" {
		notificationType = input.Matcher
	}

	stateDir := resolveNotificationStateDir()
	if mkErr := ensureNotificationStateDir(stateDir); mkErr != nil {
		return nil
	}

	logFile := filepath.Join(stateDir, "notification-events.jsonl")
	entry := notificationLogEntry{
		Event:            "notification",
		NotificationType: notificationType,
		SessionID:        input.SessionID,
		AgentType:        input.AgentType,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
	}
	if logErr := appendNotificationLog(logFile, entry); logErr != nil {
		_ = logErr
	}

	if notificationType == "permission_prompt" && input.AgentType != "" {
		fmt.Fprintf(os.Stderr, "Notification: permission_prompt for agent_type=%s\n", input.AgentType)
	}
	if notificationType == "elicitation_dialog" && input.AgentType != "" {
		fmt.Fprintf(os.Stderr,
			"Notification: elicitation_dialog for agent_type=%s (auto-skipped in background)\n",
			input.AgentType)
	}

	return nil
}

func resolveNotificationStateDir() string {
	pluginData := os.Getenv("CLAUDE_PLUGIN_DATA")
	if pluginData != "" {
		projectRoot := os.Getenv("PROJECT_ROOT")
		if projectRoot == "" {
			cwd, err := os.Getwd()
			if err == nil {
				projectRoot = cwd
			}
		}
		hash := shortHashNotification(projectRoot)
		return filepath.Join(pluginData, "projects", hash)
	}

	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err == nil {
			projectRoot = cwd
		}
	}
	return filepath.Join(projectRoot, ".claude", "state")
}

func ensureNotificationStateDir(stateDir string) error {
	parent := filepath.Dir(stateDir)

	if isSymlink(parent) || isSymlink(stateDir) {
		return fmt.Errorf("symlinked state path refused: %s", stateDir)
	}

	if mkErr := os.MkdirAll(stateDir, 0o700); mkErr != nil {
		return fmt.Errorf("mkdir state dir: %w", mkErr)
	}

	info, statErr := os.Lstat(stateDir)
	if statErr != nil {
		return fmt.Errorf("stat state dir: %w", statErr)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("state dir is symlink: %s", stateDir)
	}
	return nil
}

func appendNotificationLog(logFile string, entry notificationLogEntry) error {
	if isSymlink(logFile) {
		return fmt.Errorf("symlinked log file refused: %s", logFile)
	}

	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal log entry: %w", err)
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	if _, writeErr := fmt.Fprintf(f, "%s\n", entryJSON); writeErr != nil {
		return fmt.Errorf("write log entry: %w", writeErr)
	}

	return rotateJSONL(logFile, 500, 400)
}

func shortHashNotification(input string) string {
	if input == "" {
		return "default"
	}
	var h uint64 = 14695981039346656037
	for i := 0; i < len(input); i++ {
		h ^= uint64(input[i])
		h *= 1099511628211
	}
	return fmt.Sprintf("%012x", h&0xffffffffffff)
}
