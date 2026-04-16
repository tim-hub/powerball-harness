package hookhandler

import (
	"bufio"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"unicode"
)

// advisorFailureEntry is a single line in the failure-log.jsonl file.
type advisorFailureEntry struct {
	TaskID   string `json:"task_id"`
	ErrorSig string `json:"error_sig"`
}

// lineNumberPattern matches patterns like "line 42:" or "line 42 " that indicate
// source-location prefixes in error messages.
var lineNumberPattern = regexp.MustCompile(`\bline\s+\d+[:\s]?`)

// ShouldConsultAdvisor returns true when the advisor should be consulted
// based on task ID, retry count, and error signature.
//
// Trigger conditions:
//   - Task has been retried >= retryThreshold times with the same error signature
//   - errorSig is non-empty (indicates a known failure pattern)
//
// Duplicate suppression: returns false if the same (taskID + errorSig) combination
// already appears in the failure log (preventing repeated identical consultations).
func ShouldConsultAdvisor(taskID string, retryCount int, errorSig string, retryThreshold int, failureLogPath string) bool {
	// Guard: below retry threshold.
	if retryCount < retryThreshold {
		return false
	}

	// Guard: no error signature means no identifiable failure pattern.
	if errorSig == "" {
		return false
	}

	// Read the failure log. If it doesn't exist, this is the first time — consult.
	data, err := os.ReadFile(failureLogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
		// Any other read error: treat as missing, allow consultation.
		return true
	}

	// Parse JSONL and check for duplicate (taskID + errorSig) combination.
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry advisorFailureEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Malformed line: skip.
			continue
		}
		if entry.TaskID == taskID && entry.ErrorSig == errorSig {
			// Already logged — suppress duplicate consultation.
			return false
		}
	}

	return true
}

// NormalizeErrorSig normalizes a raw error string into a stable signature.
//
// Transformations applied (in order):
//  1. Trim leading/trailing whitespace
//  2. Lowercase
//  3. Strip line-number references (e.g. "line 42:" -> "")
//  4. Collapse runs of whitespace into a single space
//  5. Final trim
func NormalizeErrorSig(raw string) string {
	s := strings.TrimSpace(raw)

	// Lowercase for case-insensitive comparison.
	s = strings.ToLower(s)

	// Remove line number references such as "line 42:", "line 7 ".
	s = lineNumberPattern.ReplaceAllString(s, " ")

	// Collapse whitespace runs (spaces, tabs, newlines) into a single space.
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, s)

	// Collapse multiple spaces.
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}

	return strings.TrimSpace(s)
}
