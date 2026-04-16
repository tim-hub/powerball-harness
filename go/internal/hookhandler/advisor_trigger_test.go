package hookhandler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeAdvisorLog writes JSONL entries to a temp file and returns its path.
func writeAdvisorLog(t *testing.T, entries []advisorFailureEntry) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "failure-log.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create failure log: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("encode entry: %v", err)
		}
	}
	return path
}

// --- ShouldConsultAdvisor tests ---

func TestShouldConsultAdvisor_BelowThreshold(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "no-such-file.jsonl")
	got := ShouldConsultAdvisor("task-1", 1, "some error", 3, logPath)
	if got {
		t.Error("expected false when retryCount < retryThreshold, got true")
	}
}

func TestShouldConsultAdvisor_AtThreshold_BelowByOne(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "no-such-file.jsonl")
	// retryCount == retryThreshold-1 → false
	got := ShouldConsultAdvisor("task-1", 2, "some error", 3, logPath)
	if got {
		t.Error("expected false when retryCount == threshold-1, got true")
	}
}

func TestShouldConsultAdvisor_EmptyErrorSig(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "no-such-file.jsonl")
	got := ShouldConsultAdvisor("task-1", 5, "", 3, logPath)
	if got {
		t.Error("expected false when errorSig is empty, got true")
	}
}

func TestShouldConsultAdvisor_NoFailureLog_FirstTime(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "failure-log.jsonl")
	// File does not exist — first time seeing this error.
	got := ShouldConsultAdvisor("task-1", 3, "compile error", 3, logPath)
	if !got {
		t.Error("expected true when failure log does not exist (first time), got false")
	}
}

func TestShouldConsultAdvisor_DuplicateInLog_Suppressed(t *testing.T) {
	logPath := writeAdvisorLog(t, []advisorFailureEntry{
		{TaskID: "task-42", ErrorSig: "nil pointer dereference"},
	})
	got := ShouldConsultAdvisor("task-42", 4, "nil pointer dereference", 3, logPath)
	if got {
		t.Error("expected false for duplicate (taskID + errorSig) in log, got true")
	}
}

func TestShouldConsultAdvisor_DifferentTaskID_NotSuppressed(t *testing.T) {
	logPath := writeAdvisorLog(t, []advisorFailureEntry{
		{TaskID: "task-42", ErrorSig: "nil pointer dereference"},
	})
	// Different task ID — not a duplicate.
	got := ShouldConsultAdvisor("task-99", 4, "nil pointer dereference", 3, logPath)
	if !got {
		t.Error("expected true for different task ID, got false")
	}
}

func TestShouldConsultAdvisor_DifferentErrorSig_NotSuppressed(t *testing.T) {
	logPath := writeAdvisorLog(t, []advisorFailureEntry{
		{TaskID: "task-42", ErrorSig: "nil pointer dereference"},
	})
	// Same task ID, different error sig — not a duplicate.
	got := ShouldConsultAdvisor("task-42", 4, "index out of range", 3, logPath)
	if !got {
		t.Error("expected true for different errorSig, got false")
	}
}

func TestShouldConsultAdvisor_EmptyLog_Triggers(t *testing.T) {
	logPath := writeAdvisorLog(t, nil)
	got := ShouldConsultAdvisor("task-1", 3, "some error", 3, logPath)
	if !got {
		t.Error("expected true for empty log file, got false")
	}
}

func TestShouldConsultAdvisor_MultipleEntriesInLog(t *testing.T) {
	logPath := writeAdvisorLog(t, []advisorFailureEntry{
		{TaskID: "task-1", ErrorSig: "error a"},
		{TaskID: "task-2", ErrorSig: "error b"},
		{TaskID: "task-3", ErrorSig: "error c"},
	})

	// Exact match for task-2 + error b → suppressed.
	if ShouldConsultAdvisor("task-2", 5, "error b", 3, logPath) {
		t.Error("expected false for existing (task-2, error b) pair")
	}

	// task-4 not in log → consult.
	if !ShouldConsultAdvisor("task-4", 5, "error a", 3, logPath) {
		t.Error("expected true for task-4 not in log")
	}
}

func TestShouldConsultAdvisor_ExactThreshold_True(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "failure-log.jsonl")
	// retryCount == retryThreshold → should trigger.
	got := ShouldConsultAdvisor("task-1", 3, "timeout", 3, logPath)
	if !got {
		t.Error("expected true when retryCount == retryThreshold and no log, got false")
	}
}

// --- NormalizeErrorSig tests ---

func TestNormalizeErrorSig_StripLineNumbers(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"line 42: undefined variable", "undefined variable"},
		{"line 1: syntax error", "syntax error"},
		{"error at line 100: null dereference", "error at null dereference"},
	}
	for _, tt := range tests {
		got := NormalizeErrorSig(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeErrorSig(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestNormalizeErrorSig_CollapseWhitespace(t *testing.T) {
	got := NormalizeErrorSig("  multiple   spaces  and\ttabs  ")
	want := "multiple spaces and tabs"
	if got != want {
		t.Errorf("NormalizeErrorSig whitespace collapse: got %q, want %q", got, want)
	}
}

func TestNormalizeErrorSig_Lowercase(t *testing.T) {
	got := NormalizeErrorSig("NullPointerException")
	want := "nullpointerexception"
	if got != want {
		t.Errorf("NormalizeErrorSig lowercase: got %q, want %q", got, want)
	}
}

func TestNormalizeErrorSig_Empty(t *testing.T) {
	got := NormalizeErrorSig("")
	if got != "" {
		t.Errorf("NormalizeErrorSig(\"\") = %q, want \"\"", got)
	}
}

func TestNormalizeErrorSig_TrimOnly(t *testing.T) {
	got := NormalizeErrorSig("  hello world  ")
	want := "hello world"
	if got != want {
		t.Errorf("NormalizeErrorSig trim: got %q, want %q", got, want)
	}
}

func TestNormalizeErrorSig_CombinedTransforms(t *testing.T) {
	// Line number + mixed case + extra spaces
	got := NormalizeErrorSig("  Error: line 7: Cannot Read Property  ")
	want := "error: cannot read property"
	if got != want {
		t.Errorf("NormalizeErrorSig combined: got %q, want %q", got, want)
	}
}

func TestNormalizeErrorSig_TableDriven(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"already normalized", "compile error", "compile error"},
		{"uppercase only", "PANIC", "panic"},
		{"line at end", "some error line 99", "some error"},
		{"no line number", "unexpected token {", "unexpected token {"},
		{"newline in message", "error\noccurred", "error occurred"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeErrorSig(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeErrorSig(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
