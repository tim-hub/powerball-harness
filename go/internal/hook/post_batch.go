// Package hook implements stdin/stdout codec for the Claude Code hooks hookproto.
// post_batch.go provides the concurrent POST_BATCH fan-out that collapses 8
// independent PostToolUse handlers into a single subprocess invocation.
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

// BatchResult holds a single hook handler's output and error.
type BatchResult struct {
	Name   string
	Output []byte
	Err    error
}

// hookFn pairs a handler name with its function.
type hookFn struct {
	name string
	fn   func(io.Reader, io.Writer) error
}

// postBatchHooks lists the 8 PostToolUse handlers run concurrently by
// RunPostToolBatch. They match the hooks.json entries for the Write|Edit|Task
// matcher that were collapsed into a single post-tool-batch invocation.
func postBatchHooks() []hookFn {
	return []hookFn{
		{"emit-trace", func(r io.Reader, w io.Writer) error {
			h := &hookhandler.EmitAgentTrace{}
			return h.Handle(r, w)
		}},
		{"auto-cleanup", func(r io.Reader, w io.Writer) error {
			h := &hookhandler.AutoCleanupHandler{}
			return h.Handle(r, w)
		}},
		{"track-changes", hookhandler.HandleTrackChanges},
		{"auto-test", hookhandler.HandleAutoTestRunner},
		{"quality-pack", hookhandler.HandlePostToolUseQualityPack},
		{"plans-watcher", hookhandler.HandlePlansWatcher},
		{"tdd-check", hookhandler.HandleTDDOrderCheck},
		{"auto-broadcast", hookhandler.HandleSessionAutoBroadcast},
	}
}

// RunPostToolBatch runs all 8 POST_BATCH hooks concurrently.
// Each hook receives its own bytes.Reader over the shared pre-read stdin bytes.
// Outputs are collected; the first non-empty output wins for the final response.
// If any hook errors, the error is logged to stderr but does not fail the batch
// (hooks are best-effort tracing/cleanup operations).
func RunPostToolBatch(in io.Reader, out io.Writer) error {
	input, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("post-tool-batch: reading stdin: %w", err)
	}

	hooks := postBatchHooks()
	results := make([]BatchResult, len(hooks))
	var wg sync.WaitGroup

	for i, h := range hooks {
		wg.Add(1)
		go func(idx int, hk hookFn) {
			defer wg.Done()
			var buf bytes.Buffer
			runErr := hk.fn(bytes.NewReader(input), &buf)
			results[idx] = BatchResult{
				Name:   hk.name,
				Output: buf.Bytes(),
				Err:    runErr,
			}
		}(i, h)
	}
	wg.Wait()

	return mergePostBatchOutputs(results, out)
}

// mergePostBatchOutputs combines results from all batch hooks into a single
// response. Strategy:
//   - If any hook produced a non-empty JSON output, use the first one found.
//   - All hook errors are logged to stderr (non-fatal).
//   - If no hook produced output, write an empty approve response.
func mergePostBatchOutputs(results []BatchResult, out io.Writer) error {
	var firstOutput []byte

	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "[post-tool-batch] hook %q error: %v\n", r.Name, r.Err)
		}
		if firstOutput == nil && len(bytes.TrimSpace(r.Output)) > 0 {
			firstOutput = bytes.TrimSpace(r.Output)
		}
	}

	if firstOutput != nil {
		// Validate it's JSON before forwarding.
		if json.Valid(firstOutput) {
			_, err := fmt.Fprintf(out, "%s\n", firstOutput)
			return err
		}
		fmt.Fprintf(os.Stderr, "[post-tool-batch] first non-empty output is not valid JSON, skipping\n")
	}

	// No output from any hook — emit an empty approve so CC continues.
	resp := map[string]string{"decision": "approve", "reason": "post-tool-batch: ok"}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%s\n", data)
	return err
}
