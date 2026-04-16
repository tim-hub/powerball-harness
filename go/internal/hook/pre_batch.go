// pre_batch.go provides the concurrent PRE_BATCH fan-out that collapses
// independent PreToolUse handlers for the Write|Edit matcher into a single
// subprocess invocation.
package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/tim-hub/powerball-harness/go/internal/hookhandler"
)

// PreBatchResult holds a single PreToolUse hook handler's output and error.
type PreBatchResult struct {
	Name   string
	Output []byte
	Err    error
}

// preBatchHooks lists the PreToolUse handlers run concurrently by
// RunPreToolBatch. They replace the individual Write|Edit PreToolUse hook
// entries that were collapsed into a single pre-tool-batch invocation.
func preBatchHooks() []hookFn {
	return []hookFn{
		{"inbox-check", hookhandler.HandleInboxCheck},
	}
}

// RunPreToolBatch runs all PRE_BATCH hooks concurrently.
// Each hook receives its own bytes.Reader over the shared pre-read stdin bytes.
// Outputs are merged: any hook requesting a deny takes precedence. If all
// approve (or produce no output), an approve response is returned.
// Hook errors are logged to stderr but do not fail the batch.
func RunPreToolBatch(in io.Reader, out io.Writer) error {
	input, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("pre-tool-batch: reading stdin: %w", err)
	}

	hooks := preBatchHooks()
	results := make([]PreBatchResult, len(hooks))
	var wg sync.WaitGroup

	for i, h := range hooks {
		wg.Add(1)
		go func(idx int, hk hookFn) {
			defer wg.Done()
			var buf bytes.Buffer
			runErr := hk.fn(bytes.NewReader(input), &buf)
			results[idx] = PreBatchResult{
				Name:   hk.name,
				Output: buf.Bytes(),
				Err:    runErr,
			}
		}(i, h)
	}
	wg.Wait()

	return mergePreBatchOutputs(results, out)
}

// mergePreBatchOutputs combines results from all PreToolUse batch hooks.
// Strategy:
//   - Any hook that produces a "deny" decision wins immediately.
//   - Hook errors are logged to stderr (non-fatal).
//   - If no deny and no non-empty output, write an empty approve response.
func mergePreBatchOutputs(results []PreBatchResult, out io.Writer) error {
	var firstOutput []byte

	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "[pre-tool-batch] hook %q error: %v\n", r.Name, r.Err)
		}
		trimmed := bytes.TrimSpace(r.Output)
		if len(trimmed) == 0 {
			continue
		}
		if !json.Valid(trimmed) {
			fmt.Fprintf(os.Stderr, "[pre-tool-batch] hook %q produced non-JSON output, skipping\n", r.Name)
			continue
		}
		// Check if this hook is requesting a deny — if so, forward immediately.
		var resp map[string]interface{}
		if jsonErr := json.Unmarshal(trimmed, &resp); jsonErr == nil {
			if decision, ok := resp["decision"].(string); ok && decision == "deny" {
				_, err := fmt.Fprintf(out, "%s\n", trimmed)
				return err
			}
		}
		if firstOutput == nil {
			firstOutput = trimmed
		}
	}

	if firstOutput != nil {
		_, err := fmt.Fprintf(out, "%s\n", firstOutput)
		return err
	}

	// No output from any hook — emit an approve so CC continues.
	resp := map[string]string{"decision": "approve", "reason": "pre-tool-batch: ok"}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%s\n", data)
	return err
}
