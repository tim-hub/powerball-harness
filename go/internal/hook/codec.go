// Package hook implements stdin/stdout codec for the Claude Code hooks protocol.
package hook

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Chachamaru127/claude-code-harness/go/pkg/protocol"
)

// ReadInput reads and parses the hook input JSON from the given reader.
// It validates the required "tool_name" field per the official protocol.
func ReadInput(r io.Reader) (protocol.HookInput, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return protocol.HookInput{}, fmt.Errorf("reading stdin: %w", err)
	}

	raw := string(data)
	if len(raw) == 0 || raw == "\n" || raw == "\r\n" {
		return protocol.HookInput{}, fmt.Errorf("empty input")
	}

	var input protocol.HookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return protocol.HookInput{}, fmt.Errorf("parsing JSON: %w", err)
	}

	if input.ToolName == "" {
		return protocol.HookInput{}, fmt.Errorf("missing required field 'tool_name'")
	}

	// Ensure ToolInput is not nil
	if input.ToolInput == nil {
		input.ToolInput = make(map[string]interface{})
	}

	return input, nil
}

// WriteResult writes a HookResult as JSON to the given writer, followed by a newline.
func WriteResult(w io.Writer, result protocol.HookResult) error {
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
func SafeResult(err error) protocol.HookResult {
	return protocol.HookResult{
		Decision: protocol.DecisionApprove,
		Reason:   fmt.Sprintf("Core engine error (safe fallback): %s", err.Error()),
	}
}
