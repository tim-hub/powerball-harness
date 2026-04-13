package hookhandler

import (
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

// CIStatusCheckerHandler is a PostToolUse (Bash) hook handler (CI status check).
// Detects git push / gh pr commands and synchronously checks the CI status.
// Assumes async: true hooks (CC keeps the process alive for up to 600s) and
// runs the runner as a blocking call rather than a goroutine.
// On CI failure, recommends the /ci skill via additionalContext.
//
// shell version: scripts/hook-handlers/ci-status-checker.sh
type CIStatusCheckerHandler struct {
	// ProjectRoot is the path to the project root. Uses cwd if empty.
	ProjectRoot string

	// GHCommand is the path to the gh command (for testing). Searches PATH if empty.
	GHCommand string

	// AsyncRunner is the function that executes the CI check (test mock).
	// Uses the default synchronous blocking implementation if nil.
	AsyncRunner func(projectRoot, stateDir, bashCmd, ghCommand string)
}

// ciStatusInput is the input for the PostToolUse hook.
type ciStatusInput struct {
	ToolName string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
	ToolResponse struct {
		ExitCode *int   `json:"exit_code"`
		ExitCode2 *int  `json:"exitCode"`
		Output   string `json:"output"`
		Stdout   string `json:"stdout"`
	} `json:"tool_response"`
}

// ciStatusResponse is the response from the CIStatusChecker hook.
type ciStatusResponse struct {
	Decision          string `json:"decision"`
	Reason            string `json:"reason"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

// ciRunEntry represents a single entry from gh run list.
type ciRunEntry struct {
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Name       string `json:"name"`
	URL        string `json:"url"`
}

// pushOrPRCommandRe is the regular expression to detect git push / gh pr / gh workflow run commands.
var pushOrPRCommandRe = regexp.MustCompile(`(?:^|[\s;|&])(git\s+push|gh\s+pr\s+(?:create|merge|edit)|gh\s+workflow\s+run)`)

// Handle reads the payload from stdin, detects push/PR commands, and starts CI monitoring.
func (h *CIStatusCheckerHandler) Handle(r io.Reader, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil || len(data) == 0 {
		return writeCIJSON(w, ciStatusResponse{
			Decision: "approve",
			Reason:   "ci-status-checker: no payload",
		})
	}

	var input ciStatusInput
	if err := json.Unmarshal(data, &input); err != nil {
		return writeCIJSON(w, ciStatusResponse{
			Decision: "approve",
			Reason:   "ci-status-checker: parse error",
		})
	}

	bashCmd := input.ToolInput.Command

	// skip if not a git push / gh pr command
	if !isPushOrPRCommand(bashCmd) {
		return writeCIJSON(w, ciStatusResponse{
			Decision: "approve",
			Reason:   "ci-status-checker: not a push/PR command",
		})
	}

	// skip if the gh command is not found
	ghCmd := h.resolveGHCommand()
	if ghCmd == "" {
		return writeCIJSON(w, ciStatusResponse{
			Decision: "approve",
			Reason:   "ci-status-checker: gh command not found",
		})
	}

	projectRoot := h.ProjectRoot
	if projectRoot == "" {
		projectRoot, _ = os.Getwd()
	}

	stateDir := filepath.Join(projectRoot, ".claude", "state")
	_ = os.MkdirAll(stateDir, 0700)

	// check for recent CI failure signals (before running the runner)
	additionalContext := h.checkRecentCIFailure(stateDir, bashCmd)

	// write the response to stdout first (CC keeps the process alive since this is an async: true hook)
	var writeErr error
	if additionalContext != "" {
		writeErr = writeCIJSON(w, ciStatusResponse{
			Decision:          "approve",
			Reason:            "ci-status-checker: push/PR detected, CI failure context injected",
			AdditionalContext: additionalContext,
		})
	} else {
		writeErr = writeCIJSON(w, ciStatusResponse{
			Decision: "approve",
			Reason:   "ci-status-checker: push/PR detected, CI monitoring started",
		})
	}
	if writeErr != nil {
		return writeErr
	}

	// after writing the response, poll CI status in a blocking call.
	// CC keeps the process alive up to 600s because this is an async: true hook.
	// no goroutine needed — eliminates the risk of being killed when the process exits.
	runner := h.AsyncRunner
	if runner == nil {
		runner = defaultCIRunner
	}
	runner(projectRoot, stateDir, bashCmd, ghCmd)
	return nil
}

// isPushOrPRCommand returns true if bashCmd contains a push/PR command.
func isPushOrPRCommand(cmd string) bool {
	return pushOrPRCommandRe.MatchString(cmd)
}

// resolveGHCommand returns the path to the gh command. Returns empty string if not found.
func (h *CIStatusCheckerHandler) resolveGHCommand() string {
	if h.GHCommand != "" {
		if _, err := os.Stat(h.GHCommand); err == nil {
			return h.GHCommand
		}
		return ""
	}
	path, err := exec.LookPath("gh")
	if err != nil {
		return ""
	}
	return path
}

// checkRecentCIFailure checks for a recent ci_failure_detected signal and returns a message.
func (h *CIStatusCheckerHandler) checkRecentCIFailure(stateDir, bashCmd string) string {
	signalsFile := filepath.Join(stateDir, "breezing-signals.jsonl")

	f, err := os.Open(signalsFile)
	if err != nil {
		return ""
	}
	defer f.Close()

	// find the last ci_failure_detected signal
	var lastFailureLine string
	buf := make([]byte, 1<<20) // max 1MB
	n, _ := f.Read(buf)
	content := string(buf[:n])

	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, `"ci_failure_detected"`) {
			lastFailureLine = line
		}
	}

	if lastFailureLine == "" {
		return ""
	}

	var sig map[string]interface{}
	if err := json.Unmarshal([]byte(lastFailureLine), &sig); err != nil {
		return ""
	}

	conclusion, _ := sig["conclusion"].(string)
	return fmt.Sprintf(
		"[CI failure detected]\nCI status: %s\nTrigger command: %s\n\nRecommended action: spawn /breezing or the ci-cd-fixer agent to automatically repair the CI failure.\n  Example: ask ci-cd-fixer \"CI failed. Please check the logs and fix the issue.\"",
		conclusion, bashCmd,
	)
}

// defaultCIRunner polls gh run list and writes results to the signal file.
// Runs as a synchronous blocking call, assuming async: true hooks.
// CC keeps async: true hook processes alive for up to 600s, so maxWait is set to 120s
// to give GitHub Actions enough time to complete.
// (The previous 25s allowed only 2 polls at 10s intervals, causing most CI runs to finish unmonitored.)
func defaultCIRunner(projectRoot, stateDir, bashCmd, ghCmd string) {
	const maxWait = 120 * time.Second
	const pollInterval = 10 * time.Second

	ciStatusFile := filepath.Join(stateDir, "ci-status.json")
	signalsFile := filepath.Join(stateDir, "breezing-signals.jsonl")

	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		time.Sleep(pollInterval)

		out, err := exec.Command(ghCmd, "run", "list", "--limit", "1", "--json", "status,conclusion,name,url").Output()
		if err != nil || len(out) == 0 {
			continue
		}

		var runs []ciRunEntry
		if err := json.Unmarshal(out, &runs); err != nil || len(runs) == 0 {
			continue
		}

		run := runs[0]
		if run.Status != "completed" {
			continue
		}

		// record the result
		statusData, _ := json.Marshal(map[string]string{
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
			"trigger_command": bashCmd,
			"status":          run.Status,
			"conclusion":      run.Conclusion,
		})
		_ = os.WriteFile(ciStatusFile, statusData, 0600)

		// append to the signal file on CI failure
		if run.Conclusion == "failure" || run.Conclusion == "timed_out" || run.Conclusion == "cancelled" {
			sig, _ := json.Marshal(map[string]string{
				"signal":          "ci_failure_detected",
				"timestamp":       time.Now().UTC().Format(time.RFC3339),
				"conclusion":      run.Conclusion,
				"trigger_command": bashCmd,
			})
			f, err := os.OpenFile(signalsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err == nil {
				_, _ = f.Write(sig)
				_, _ = f.Write([]byte("\n"))
				f.Close()
			}
		}

		return
	}
}

// writeCIJSON writes v as JSON to w.
func writeCIJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}
