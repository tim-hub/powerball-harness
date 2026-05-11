package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// worktreeInput is the stdin JSON payload for the WorktreeCreate hook.
type worktreeInput struct {
	SessionID     string `json:"session_id"`
	CWD           string `json:"cwd"`
	HookEventName string `json:"hook_event_name"`
}

// worktreeInfo is written to .claude/state/worktree-info.json.
type worktreeInfo struct {
	WorkerID  string `json:"worker_id"`
	CreatedAt string `json:"created_at"`
	CWD       string `json:"cwd"`
}

// HandleWorktreeCreate ports scripts/hook-handlers/worktree-create.sh.
//
// WorktreeCreate is a lifecycle/notification event — CC does not expect or
// consume any stdout response. The handler writes nothing to out; it only
// creates .claude/state/ and worktree-info.json inside the worktree CWD.
// Writing JSON to stdout here was the root cause of CC re-feeding the hook
// output as a cwd value on subsequent invocations.
func HandleWorktreeCreate(in io.Reader, _ io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(data) == 0 {
		return nil
	}

	var input worktreeInput
	if jsonErr := json.Unmarshal(data, &input); jsonErr != nil {
		return nil
	}

	if input.CWD == "" {
		return nil
	}

	stateDir := input.CWD + "/.claude/state"
	if mkErr := os.MkdirAll(stateDir, 0o755); mkErr != nil {
		fmt.Fprintf(os.Stderr, "[claude-code-harness] worktree-create: mkdir %s: %v\n", stateDir, mkErr)
	}

	info := worktreeInfo{
		WorkerID:  input.SessionID,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		CWD:       input.CWD,
	}
	infoData, err := json.Marshal(info)
	if err == nil {
		infoPath := stateDir + "/worktree-info.json"
		_ = os.WriteFile(infoPath, append(infoData, '\n'), 0o644)
	}

	return nil
}
