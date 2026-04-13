package hookhandler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type trackChangesInput struct {
	ToolName string `json:"tool_name"`
	CWD      string `json:"cwd"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
	ToolResponse struct {
		FilePath string `json:"filePath"`
	} `json:"tool_response"`
}

type changedFileEntry struct {
	File      string `json:"file"`
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	Important bool   `json:"important"`
}

const trackChangesMaxLines = 500

const trackChangesDedupWindow = 2 * time.Hour

const changedFilesPath = ".claude/state/changed-files.jsonl"

var importantFilePatterns = []string{
	"Plans.md",
	"CLAUDE.md",
	"AGENTS.md",
}

//
//
func HandleTrackChanges(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil {
		return emptyPostToolOutput(out)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return emptyPostToolOutput(out)
	}

	var input trackChangesInput
	if err := json.Unmarshal(data, &input); err != nil {
		return emptyPostToolOutput(out)
	}

	filePath := input.ToolInput.FilePath
	if filePath == "" {
		filePath = input.ToolResponse.FilePath
	}

	if filePath == "" {
		return emptyPostToolOutput(out)
	}

	filePath = normalizePathSeparators(filePath)

	if input.CWD != "" {
		cwd := normalizePathSeparators(input.CWD)
		filePath = makeRelativePath(filePath, cwd)
	}

	toolName := input.ToolName
	if toolName == "" {
		toolName = "unknown"
	}

	important := isImportantFile(filePath)

	now := time.Now().UTC()
	timestamp := now.Format(time.RFC3339)

	if isDuplicateWithin(filePath, now, trackChangesDedupWindow) {
		return emptyPostToolOutput(out)
	}

	stateDir := filepath.Dir(changedFilesPath)
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return emptyPostToolOutput(out)
	}

	if err := rotateIfNeeded(changedFilesPath, trackChangesMaxLines); err != nil {
		fmt.Fprintf(os.Stderr, "[track-changes] rotate: %v\n", err)
	}

	entry := changedFileEntry{
		File:      filePath,
		Action:    toolName,
		Timestamp: timestamp,
		Important: important,
	}
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return emptyPostToolOutput(out)
	}

	f, err := os.OpenFile(changedFilesPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return emptyPostToolOutput(out)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "%s\n", entryJSON); err != nil {
		return emptyPostToolOutput(out)
	}

	return emptyPostToolOutput(out)
}

func normalizePathSeparators(p string) string {
	return strings.ReplaceAll(p, "\\", "/")
}

func makeRelativePath(filePath, cwd string) string {
	cwdWithSlash := strings.TrimRight(cwd, "/") + "/"
	if strings.HasPrefix(filePath+"/", cwdWithSlash) || filePath == strings.TrimRight(cwd, "/") {
		if strings.HasPrefix(filePath, cwdWithSlash) {
			return filePath[len(cwdWithSlash):]
		}
	}
	return filePath
}

func isImportantFile(filePath string) bool {
	for _, pattern := range importantFilePatterns {
		if strings.Contains(filePath, pattern) {
			return true
		}
	}
	if strings.Contains(filePath, ".test.") ||
		strings.Contains(filePath, ".spec.") ||
		strings.Contains(filePath, "__tests__") {
		return true
	}
	return false
}

func isDuplicateWithin(filePath string, now time.Time, window time.Duration) bool {
	f, err := os.Open(changedFilesPath)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry changedFileEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.File != filePath {
			continue
		}
		t, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			continue
		}
		if now.Sub(t) < window {
			return true
		}
	}
	return false
}

func rotateIfNeeded(path string, maxLines int) error {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	f.Close()

	if len(lines) <= maxLines {
		return nil
	}

	lines = lines[len(lines)-maxLines:]

	tmpPath := path + ".tmp"
	tmp, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create tmp: %w", err)
	}

	w := bufio.NewWriter(tmp)
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			tmp.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("write tmp: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("flush tmp: %w", err)
	}
	tmp.Close()

	return os.Rename(tmpPath, path)
}
