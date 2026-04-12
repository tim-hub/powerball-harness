// Package hook implements stdin/stdout codec for the Claude Code hooks hookproto.
package hook

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Chachamaru127/claude-code-harness/go/pkg/hookproto"
)

// ReadInput reads and parses the hook input JSON from the given reader.
// It validates the required "tool_name" field per the official hookproto.
func ReadInput(r io.Reader) (hookproto.HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return hookproto.HookInput{}, fmt.Errorf("reading stdin: %w", err)
	}

	raw := string(data)
	if len(raw) == 0 || raw == "\n" || raw == "\r\n" {
		return hookproto.HookInput{}, fmt.Errorf("empty input")
	}

	var input hookproto.HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return hookproto.HookInput{}, fmt.Errorf("parsing JSON: %w", err)
	}

	if input.ToolName == "" {
		return hookproto.HookInput{}, fmt.Errorf("missing required field 'tool_name'")
	}

	// Ensure ToolInput is not nil
	if input.ToolInput == nil {
		input.ToolInput = make(map[string]interface{})
	}

	return input, nil
}

// WriteResult writes a HookResult as JSON to the given writer, followed by a newline.
func WriteResult(w io.Writer, result hookproto.HookResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshaling result: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

// WriteJSON writes an arbitrary value as JSON to the given writer, followed by a newline.
func WriteJSON(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", data)
	return err
}

// SafeResult returns an approve result with an error message (fail-open).
func SafeResult(err error) hookproto.HookResult {
	return hookproto.HookResult{
		Decision: hookproto.DecisionApprove,
		Reason:   fmt.Sprintf("Core engine error (safe fallback): %s", err.Error()),
	}
}
