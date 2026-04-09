// Package hookhandler implements Go ports of the bash hook handler scripts.
package hookhandler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CheckInterval is the minimum duration between inbox checks (5 minutes).
const CheckInterval = 5 * time.Minute

// broadcastMsgRe matches broadcast.md message headers.
// Format: ## 2026-04-09T12:34:56Z [sender-prefix]
var broadcastMsgRe = regexp.MustCompile(`^## (\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z) \[([^\]]+)\]`)

// inboxLine represents a single line parsed from session-inbox.jsonl.
// Kept for backward compatibility in case inbox JSONL is present.
type inboxLine struct {
	Read bool   `json:"read"`
	Msg  string `json:"msg"`
}

// inboxCheckInput is the stdin JSON payload for PreToolUse hooks.
type inboxCheckInput struct {
	SessionID string `json:"session_id"`
	CWD       string `json:"cwd"`
}

// preToolAllowOutput matches the hookSpecificOutput format for PreToolUse.
type preToolAllowOutput struct {
	HookSpecificOutput struct {
		HookEventName      string `json:"hookEventName"`
		PermissionDecision string `json:"permissionDecision"`
		AdditionalContext  string `json:"additionalContext,omitempty"`
	} `json:"hookSpecificOutput"`
}

// HandleInboxCheck ports pretooluse-inbox-check.sh.
//
// Reads .claude/sessions/broadcast.md (same source as the bash version),
// counts unread messages since the last read timestamp, and if there are any
// injects them as additionalContext. A 5-minute throttle is enforced via
// .claude/sessions/.last_inbox_check.
func HandleInboxCheck(in io.Reader, out io.Writer) error {
	// Read stdin (ignored — the hook payload is not needed for this handler).
	_, _ = io.ReadAll(in)

	// Resolve project root from CWD or environment, same pattern as bash script.
	projectRoot := resolveProjectRoot()

	sessionsDir := projectRoot + "/.claude/sessions"
	checkIntervalFile := sessionsDir + "/.last_inbox_check"
	broadcastFile := sessionsDir + "/broadcast.md"

	// Throttle: exit 0 (no output) if last check was < 5 minutes ago.
	if !throttleAllowed(checkIntervalFile) {
		return nil
	}

	// Update the last-check timestamp.
	if err := os.MkdirAll(sessionsDir, 0o755); err == nil {
		now := strconv.FormatInt(time.Now().Unix(), 10)
		_ = os.WriteFile(checkIntervalFile, []byte(now+"\n"), 0o644)
	}

	// If broadcast.md does not exist, nothing to do.
	if _, err := os.Stat(broadcastFile); os.IsNotExist(err) {
		return nil
	}

	// Read unread messages from broadcast.md (markdown format — matches bash version).
	messages, err := readBroadcastMessages(broadcastFile, 5)
	if err != nil || len(messages) == 0 {
		// Fallback: try session-inbox.jsonl for backward compatibility.
		inboxFile := projectRoot + "/.claude/state/session-inbox.jsonl"
		messages, _ = readUnreadMessages(inboxFile, 5)
		if len(messages) == 0 {
			return nil
		}
	}

	// Build additionalContext string.
	context := fmt.Sprintf("📨 他セッションからのメッセージ %d件:\n---\n%s\n---",
		len(messages), strings.Join(messages, "\n"))

	output := preToolAllowOutput{}
	output.HookSpecificOutput.HookEventName = "PreToolUse"
	output.HookSpecificOutput.PermissionDecision = "allow"
	output.HookSpecificOutput.AdditionalContext = context

	data, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("marshaling output: %w", err)
	}
	_, err = fmt.Fprintf(out, "%s\n", data)
	return err
}

// throttleAllowed returns true when enough time has passed since the last check.
func throttleAllowed(checkIntervalFile string) bool {
	data, err := os.ReadFile(checkIntervalFile)
	if err != nil {
		// File doesn't exist yet — first check is always allowed.
		return true
	}
	raw := strings.TrimSpace(string(data))
	lastCheck, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return true
	}
	elapsed := time.Since(time.Unix(lastCheck, 0))
	return elapsed >= CheckInterval
}

// readBroadcastMessages reads up to maxCount unread messages from broadcast.md.
// The markdown format used by session-inbox-check.sh:
//
//	## 2026-04-09T12:34:56Z [sender-short-id]
//	message content line
//
// Messages are considered unread if they appear in the file (no per-session
// read-state is tracked in the Go implementation; the 5-minute throttle
// prevents excessive notifications).
func readBroadcastMessages(path string, maxCount int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var msgs []string
	var currentTimestamp, currentSender, currentContent string
	inMessage := false

	flush := func() {
		if inMessage && currentContent != "" && len(msgs) < maxCount {
			// Format: [HH:MM] sender: content
			ts := currentTimestamp
			if len(ts) >= 16 {
				ts = ts[11:16] // extract HH:MM from ISO timestamp
			}
			msgs = append(msgs, fmt.Sprintf("[%s] %s: %s", ts, currentSender, currentContent))
		}
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if m := broadcastMsgRe.FindStringSubmatch(line); m != nil {
			// Flush previous message before starting a new one.
			flush()
			currentTimestamp = m[1]
			currentSender = m[2]
			currentContent = ""
			inMessage = true
			continue
		}

		if inMessage && strings.TrimSpace(line) != "" {
			currentContent = strings.TrimSpace(line)
		}
	}
	// Flush the last message.
	flush()

	return msgs, scanner.Err()
}

// readUnreadMessages reads up to maxCount unread messages from a JSONL inbox
// file. Each line is expected to be a JSON object; lines that are not valid
// JSON are treated as raw text messages. Lines starting with '[' are treated
// as pre-formatted message lines (matching the bash grep pattern).
func readUnreadMessages(path string, maxCount int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var msgs []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() && len(msgs) < maxCount {
		line := scanner.Text()
		if line == "" {
			continue
		}
		// Try to parse as JSON to check read status.
		var entry inboxLine
		if jsonErr := json.Unmarshal([]byte(line), &entry); jsonErr == nil {
			if !entry.Read && entry.Msg != "" {
				msgs = append(msgs, entry.Msg)
			}
			continue
		}
		// Fallback: treat lines beginning with '[' as unread messages (bash compat).
		if strings.HasPrefix(line, "[") {
			msgs = append(msgs, line)
		}
	}
	return msgs, scanner.Err()
}
