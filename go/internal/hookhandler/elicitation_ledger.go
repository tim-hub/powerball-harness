package hookhandler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	elicitationEventSchemaVersion = "elicitation-event.v1"
	elicitationLedgerRelPath      = ".claude/state/elicitation/events.jsonl"
)

var allowedElicitationPrivacyTags = map[string]bool{
	"may_train":      true,
	"do_not_train":   true,
	"synthetic_only": true,
	"legal_hold":     true,
}

var allowedElicitationEventKinds = map[string]bool{
	"capability_probe": true,
	"weak_label":       true,
	"judge_verdict":    true,
	"eval_result":      true,
	"counterexample":   true,
}

var allowedElicitationVerdicts = map[string]bool{
	"":                true,
	"APPROVE":         true,
	"REQUEST_CHANGES": true,
	"STOP":            true,
	"INFO":            true,
}

// ElicitationEvent is the local append-only ledger record used by the
// weak-supervision loop. It stays independent from any external memory system;
// scripts may forward events over CLI/HTTP/MCP contracts.
type ElicitationEvent struct {
	SchemaVersion string   `json:"schema_version"`
	EventKind     string   `json:"event_kind"`
	RunID         string   `json:"run_id"`
	TaskID        string   `json:"task_id,omitempty"`
	RubricID      string   `json:"rubric_id,omitempty"`
	RewardScore   *float64 `json:"reward_score,omitempty"`
	Verdict       string   `json:"verdict,omitempty"`
	PrivacyTags   []string `json:"privacy_tags"`
	EvidenceRefs  []string `json:"evidence_refs"`
	Source        string   `json:"source"`
	Timestamp     string   `json:"timestamp"`

	MCPServer       string `json:"mcp_server,omitempty"`
	ElicitationID   string `json:"elicitation_id,omitempty"`
	Message         string `json:"message,omitempty"`
	ResultStatus    string `json:"result_status,omitempty"`
	BreezingSession string `json:"breezing_session,omitempty"`
}

func defaultElicitationRunID() string {
	for _, key := range []string{
		"HARNESS_RUN_ID",
		"HARNESS_BREEZING_SESSION_ID",
		"CLAUDE_SESSION_ID",
		"CODEX_SESSION_ID",
	} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return "local"
}

func defaultElicitationPrivacyTags() []string {
	if raw := strings.TrimSpace(os.Getenv("HARNESS_ELICITATION_PRIVACY_TAGS")); raw != "" {
		parts := strings.Split(raw, ",")
		tags := make([]string, 0, len(parts))
		for _, part := range parts {
			tag := strings.TrimSpace(part)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
		if len(tags) > 0 {
			return tags
		}
	}
	return []string{"do_not_train"}
}

func validateElicitationPrivacyTags(tags []string) error {
	if len(tags) == 0 {
		return fmt.Errorf("privacy_tags must not be empty")
	}
	for _, tag := range tags {
		if !allowedElicitationPrivacyTags[tag] {
			return fmt.Errorf("unknown privacy tag: %s", tag)
		}
	}
	return nil
}

func validateElicitationEvent(event ElicitationEvent) error {
	if event.SchemaVersion != "" && event.SchemaVersion != elicitationEventSchemaVersion {
		return fmt.Errorf("schema_version must be %s", elicitationEventSchemaVersion)
	}
	if !allowedElicitationEventKinds[event.EventKind] {
		return fmt.Errorf("unknown event_kind: %s", event.EventKind)
	}
	if !allowedElicitationVerdicts[event.Verdict] {
		return fmt.Errorf("unknown verdict: %s", event.Verdict)
	}
	if event.RewardScore != nil && (*event.RewardScore < 0 || *event.RewardScore > 1) {
		return fmt.Errorf("reward_score must be between 0 and 1")
	}
	return validateElicitationPrivacyTags(event.PrivacyTags)
}

func appendElicitationEvent(projectRoot string, event ElicitationEvent) (string, error) {
	if projectRoot == "" {
		projectRoot = resolveProjectRoot()
	}
	if event.SchemaVersion == "" {
		event.SchemaVersion = elicitationEventSchemaVersion
	}
	if event.RunID == "" {
		event.RunID = defaultElicitationRunID()
	}
	if len(event.PrivacyTags) == 0 {
		event.PrivacyTags = defaultElicitationPrivacyTags()
	}
	if event.EvidenceRefs == nil {
		event.EvidenceRefs = []string{}
	}
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if err := validateElicitationEvent(event); err != nil {
		return "", err
	}

	ledgerPath := filepath.Join(projectRoot, filepath.FromSlash(elicitationLedgerRelPath))
	if isSymlink(ledgerPath) {
		return "", fmt.Errorf("refusing to append to symlinked ledger")
	}
	if err := os.MkdirAll(filepath.Dir(ledgerPath), 0o700); err != nil {
		return "", err
	}
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(ledgerPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "%s\n", data); err != nil {
		return "", err
	}
	return ledgerPath, nil
}

func newElicitationRequestEvent(mcpServer, elicitationID, message string) ElicitationEvent {
	return ElicitationEvent{
		SchemaVersion:   elicitationEventSchemaVersion,
		EventKind:       "capability_probe",
		RunID:           defaultElicitationRunID(),
		PrivacyTags:     defaultElicitationPrivacyTags(),
		EvidenceRefs:    []string{},
		Source:          "claude-code-hook:elicitation",
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		MCPServer:       mcpServer,
		ElicitationID:   elicitationID,
		Message:         message,
		BreezingSession: os.Getenv("HARNESS_BREEZING_SESSION_ID"),
	}
}

func newElicitationResultEvent(mcpServer, elicitationID, resultStatus string) ElicitationEvent {
	return ElicitationEvent{
		SchemaVersion: elicitationEventSchemaVersion,
		EventKind:     "eval_result",
		RunID:         defaultElicitationRunID(),
		PrivacyTags:   defaultElicitationPrivacyTags(),
		EvidenceRefs:  []string{},
		Source:        "claude-code-hook:elicitation-result",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		MCPServer:     mcpServer,
		ElicitationID: elicitationID,
		ResultStatus:  resultStatus,
	}
}
