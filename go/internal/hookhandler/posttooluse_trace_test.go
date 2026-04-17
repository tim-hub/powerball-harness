package hookhandler

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// plansWithWIP is a minimal Plans.md fixture with one cc:WIP row.
const plansWithWIP = `# Test Plans

---

## Phase 72: test

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 72.3 | Wire hook | something | 72.2 | cc:WIP |
| 72.4 | Another | thing | 72.3 | cc:TODO |
`

const plansNoWIP = `# Test Plans

| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 72.1 | Done thing | | - | cc:Done [abc1234] |
| 72.2 | Not yet | | 72.1 | cc:TODO |
`

// newTraceTestEnv sets up a temp project root with a Plans.md and returns
// the root path. Use cwd=root when constructing hook input JSON.
func newTraceTestEnv(t *testing.T, plansContent string) string {
	t.Helper()
	root := t.TempDir()
	if plansContent != "" {
		if err := os.WriteFile(filepath.Join(root, "Plans.md"), []byte(plansContent), 0o600); err != nil {
			t.Fatalf("write Plans.md: %v", err)
		}
	}
	return root
}

// callHandler invokes the handler with the given input JSON. Returns stdout
// (always empty per design) and err.
func callHandler(t *testing.T, inputJSON string) ([]byte, error) {
	t.Helper()
	var out bytes.Buffer
	err := HandlePostToolUseTrace(strings.NewReader(inputJSON), &out)
	return out.Bytes(), err
}

func readTraceFile(t *testing.T, root, taskID string) []map[string]any {
	t.Helper()
	path := filepath.Join(root, ".claude", "state", "traces", taskID+".jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read trace %s: %v", path, err)
	}
	var events []map[string]any
	for line := range strings.SplitSeq(strings.TrimRight(string(data), "\n"), "\n") {
		if line == "" {
			continue
		}
		var ev map[string]any
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("invalid JSON line: %v (raw: %q)", err, line)
		}
		events = append(events, ev)
	}
	return events
}

func assertNoTraceFile(t *testing.T, root string) {
	t.Helper()
	traceDir := filepath.Join(root, ".claude", "state", "traces")
	entries, err := os.ReadDir(traceDir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatalf("read trace dir: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jsonl") {
			t.Errorf("expected no trace files, found %s", e.Name())
		}
	}
}

func TestHandlePostToolUseTrace_EmitsEventWhenWIP(t *testing.T) {
	root := newTraceTestEnv(t, plansWithWIP)
	input := `{"tool_name":"Edit","cwd":"` + root + `","tool_input":{"file_path":"foo.go"}}`

	if _, err := callHandler(t, input); err != nil {
		t.Fatalf("handler err: %v", err)
	}

	events := readTraceFile(t, root, "72.3")
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	ev := events[0]
	if ev["event_type"] != "tool_call" {
		t.Errorf("event_type=%v, want tool_call", ev["event_type"])
	}
	if ev["task_id"] != "72.3" {
		t.Errorf("task_id=%v, want 72.3", ev["task_id"])
	}
	if ev["schema"] != "trace.v1" {
		t.Errorf("schema=%v, want trace.v1", ev["schema"])
	}
	payload, _ := ev["payload"].(map[string]any)
	if payload["tool"] != "Edit" {
		t.Errorf("payload.tool=%v, want Edit", payload["tool"])
	}
	if payload["args_summary"] != "file_path=foo.go" {
		t.Errorf("payload.args_summary=%v, want file_path=foo.go", payload["args_summary"])
	}
}

func TestHandlePostToolUseTrace_SkipsWhenNoWIP(t *testing.T) {
	root := newTraceTestEnv(t, plansNoWIP)
	input := `{"tool_name":"Edit","cwd":"` + root + `","tool_input":{"file_path":"foo.go"}}`

	if _, err := callHandler(t, input); err != nil {
		t.Fatalf("handler err: %v", err)
	}
	assertNoTraceFile(t, root)
}

func TestHandlePostToolUseTrace_SkipsWhenNoPlansFile(t *testing.T) {
	root := newTraceTestEnv(t, "") // no Plans.md written
	input := `{"tool_name":"Edit","cwd":"` + root + `","tool_input":{"file_path":"foo.go"}}`

	if _, err := callHandler(t, input); err != nil {
		t.Fatalf("handler err: %v", err)
	}
	assertNoTraceFile(t, root)
}

func TestHandlePostToolUseTrace_SkipsReadOnlyTools(t *testing.T) {
	root := newTraceTestEnv(t, plansWithWIP)
	cases := []string{"Read", "Grep", "Glob", "TodoWrite", "WebFetch"}
	for _, tool := range cases {
		t.Run(tool, func(t *testing.T) {
			input := `{"tool_name":"` + tool + `","cwd":"` + root + `","tool_input":{}}`
			if _, err := callHandler(t, input); err != nil {
				t.Fatalf("handler err: %v", err)
			}
			assertNoTraceFile(t, root)
		})
	}
}

func TestHandlePostToolUseTrace_SkipsMalformedInput(t *testing.T) {
	root := newTraceTestEnv(t, plansWithWIP)
	cases := []string{
		``,                              // empty
		`not json`,                      // not JSON
		`{"tool_name":123}`,             // wrong type
		`{"cwd":"` + root + `"}`,        // missing tool_name
	}
	for i, input := range cases {
		if _, err := callHandler(t, input); err != nil {
			t.Fatalf("case %d: handler err: %v", i, err)
		}
	}
	assertNoTraceFile(t, root)
}

func TestHandlePostToolUseTrace_BashPayload(t *testing.T) {
	root := newTraceTestEnv(t, plansWithWIP)
	cmd := "go test -race ./..."
	input := `{"tool_name":"Bash","cwd":"` + root + `","tool_input":{"command":"` + cmd + `"}}`

	if _, err := callHandler(t, input); err != nil {
		t.Fatalf("handler err: %v", err)
	}
	events := readTraceFile(t, root, "72.3")
	if len(events) != 1 {
		t.Fatalf("got %d events, want 1", len(events))
	}
	payload := events[0]["payload"].(map[string]any)
	if payload["args_summary"] != "cmd="+cmd {
		t.Errorf("args_summary=%v, want cmd=%s", payload["args_summary"], cmd)
	}
}

func TestHandlePostToolUseTrace_BashTruncation(t *testing.T) {
	root := newTraceTestEnv(t, plansWithWIP)
	longCmd := strings.Repeat("x", maxBashArgsSummary+200)
	input := `{"tool_name":"Bash","cwd":"` + root + `","tool_input":{"command":"` + longCmd + `"}}`

	if _, err := callHandler(t, input); err != nil {
		t.Fatalf("handler err: %v", err)
	}
	payload := readTraceFile(t, root, "72.3")[0]["payload"].(map[string]any)
	summary, _ := payload["args_summary"].(string)
	// Summary format: "cmd=" + first 500 chars + "..."
	wantPrefix := "cmd=" + strings.Repeat("x", maxBashArgsSummary)
	if !strings.HasPrefix(summary, wantPrefix) {
		t.Errorf("summary prefix mismatch: got len=%d, want prefix len=%d", len(summary), len(wantPrefix))
	}
	if !strings.HasSuffix(summary, "...") {
		t.Errorf("long bash command should end with '...', got %q (last 10)", summary[len(summary)-10:])
	}
}

func TestHandlePostToolUseTrace_TaskPayload(t *testing.T) {
	root := newTraceTestEnv(t, plansWithWIP)
	input := `{"tool_name":"Task","cwd":"` + root + `","tool_input":{"subagent_type":"harness:worker"}}`

	if _, err := callHandler(t, input); err != nil {
		t.Fatalf("handler err: %v", err)
	}
	payload := readTraceFile(t, root, "72.3")[0]["payload"].(map[string]any)
	if payload["args_summary"] != "subagent=harness:worker" {
		t.Errorf("args_summary=%v, want subagent=harness:worker", payload["args_summary"])
	}
}

func TestHandlePostToolUseTrace_NoStdoutOutput(t *testing.T) {
	// The handler must never write anything to stdout — it's observation-only.
	// Any stdout output could be misinterpreted by the hook framework as
	// injected context.
	root := newTraceTestEnv(t, plansWithWIP)
	input := `{"tool_name":"Edit","cwd":"` + root + `","tool_input":{"file_path":"x.go"}}`
	out, err := callHandler(t, input)
	if err != nil {
		t.Fatalf("handler err: %v", err)
	}
	if len(out) > 0 {
		t.Errorf("handler wrote %d bytes to stdout, want 0: %q", len(out), string(out))
	}
}

func TestHandlePostToolUseTrace_MultipleWIPTakesFirst(t *testing.T) {
	// If two WIP tasks exist (edge case, multi-worker breezing), the first
	// one in file order wins. Plans.md is organized newest-phase-first, so
	// this means the most recent work-in-progress gets the event.
	plans := `| Task | Description | DoD | Depends | Status |
|------|-------------|-----|---------|--------|
| 99.1 | Newer | | - | cc:WIP |
| 72.3 | Older | | - | cc:WIP |
`
	root := newTraceTestEnv(t, plans)
	input := `{"tool_name":"Edit","cwd":"` + root + `","tool_input":{"file_path":"x.go"}}`

	if _, err := callHandler(t, input); err != nil {
		t.Fatalf("handler err: %v", err)
	}
	events := readTraceFile(t, root, "99.1")
	if len(events) != 1 {
		t.Fatalf("got %d events in 99.1.jsonl, want 1", len(events))
	}
	// 72.3 must NOT have a file
	other := filepath.Join(root, ".claude", "state", "traces", "72.3.jsonl")
	if _, err := os.Stat(other); !os.IsNotExist(err) {
		t.Errorf("second WIP task should not get events: %v", err)
	}
}

func TestFindActiveWIPTask_Regex(t *testing.T) {
	// Direct unit test of the regex — important because it's the contract
	// with Plans.md format. Failure here is the most likely source of
	// silent "why is no trace file being written?" bugs.
	cases := []struct {
		name string
		line string
		want string // "" = no match
	}{
		{"basic cc:WIP row", "| 72.3 | desc | dod | - | cc:WIP |", "72.3"},
		{"deep dotted id", "| 72.1.fix | desc | dod | - | cc:WIP |", "72.1.fix"},
		{"single-int id", "| 5 | desc | | | cc:WIP |", "5"},
		{"cc:Done row ignored", "| 72.1 | desc | | | cc:Done [abc1234] |", ""},
		{"cc:TODO row ignored", "| 72.4 | desc | | | cc:TODO |", ""},
		{"non-table line ignored", "some text with cc:WIP in it", ""},
		{"header row ignored", "| Task | Description | DoD | Depends | Status |", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := wipRowPattern.FindStringSubmatch(tc.line)
			got := ""
			if len(m) > 1 {
				got = m[1]
			}
			if got != tc.want {
				t.Errorf("line=%q: got %q, want %q", tc.line, got, tc.want)
			}
		})
	}
}

// Compile-time check: HandlePostToolUseTrace matches the standard
// hookhandler signature used throughout the package.
var _ = func(r io.Reader, w io.Writer) error { return HandlePostToolUseTrace(r, w) }
