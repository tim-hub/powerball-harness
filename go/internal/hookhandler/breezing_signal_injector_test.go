package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSignalsFile(t *testing.T, dir string, lines []string) string {
	t.Helper()
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	signalsFile := filepath.Join(stateDir, "breezing-signals.jsonl")
	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(signalsFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return signalsFile
}

func writeActiveFile(t *testing.T, dir string) {
	t.Helper()
	stateDir := filepath.Join(dir, ".claude", "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	activeFile := filepath.Join(stateDir, "breezing-active.json")
	if err := os.WriteFile(activeFile, []byte(`{"active":true}`), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestBreezingSignalInjector_NoActiveFile(t *testing.T) {
	dir := t.TempDir()
	h := &BreezingSignalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// breezing 非アクティブ時は何も出力しない
	if out.Len() != 0 {
		t.Errorf("expected no output when breezing is inactive, got: %s", out.String())
	}
}

func TestBreezingSignalInjector_NoSignalsFile(t *testing.T) {
	dir := t.TempDir()
	writeActiveFile(t, dir)

	h := &BreezingSignalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// シグナルファイルなし → 何も出力しない
	if out.Len() != 0 {
		t.Errorf("expected no output when signals file absent, got: %s", out.String())
	}
}

func TestBreezingSignalInjector_AllConsumed(t *testing.T) {
	dir := t.TempDir()
	writeActiveFile(t, dir)
	writeSignalsFile(t, dir, []string{
		`{"signal":"ci_failure_detected","conclusion":"failure","consumed_at":"2026-01-01T00:00:00Z"}`,
	})

	h := &BreezingSignalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 全消費済み → 何も出力しない
	if out.Len() != 0 {
		t.Errorf("expected no output when all signals consumed, got: %s", out.String())
	}
}

func TestBreezingSignalInjector_CIFailureSignal(t *testing.T) {
	dir := t.TempDir()
	writeActiveFile(t, dir)
	writeSignalsFile(t, dir, []string{
		`{"signal":"ci_failure_detected","conclusion":"failure","trigger_command":"git push origin main"}`,
	})

	h := &BreezingSignalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Len() == 0 {
		t.Fatalf("expected output for CI failure signal")
	}

	var resp injectorResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	if !strings.Contains(resp.SystemMessage, "[SIGNAL:ci_failure_detected]") {
		t.Errorf("expected ci_failure_detected in message, got: %s", resp.SystemMessage)
	}
	if !strings.Contains(resp.SystemMessage, "failure") {
		t.Errorf("expected conclusion in message, got: %s", resp.SystemMessage)
	}
	if !strings.Contains(resp.SystemMessage, "git push origin main") {
		t.Errorf("expected trigger command in message, got: %s", resp.SystemMessage)
	}
}

func TestBreezingSignalInjector_RetakeRequestedSignal(t *testing.T) {
	dir := t.TempDir()
	writeActiveFile(t, dir)
	writeSignalsFile(t, dir, []string{
		`{"signal":"retake_requested","task_id":"32.1","reason":"テスト失敗"}`,
	})

	h := &BreezingSignalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp injectorResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	if !strings.Contains(resp.SystemMessage, "[SIGNAL:retake_requested]") {
		t.Errorf("expected retake_requested in message")
	}
	if !strings.Contains(resp.SystemMessage, "32.1") {
		t.Errorf("expected task_id in message")
	}
	if !strings.Contains(resp.SystemMessage, "テスト失敗") {
		t.Errorf("expected reason in message")
	}
}

func TestBreezingSignalInjector_ReviewerApprovedSignal(t *testing.T) {
	dir := t.TempDir()
	writeActiveFile(t, dir)
	writeSignalsFile(t, dir, []string{
		`{"signal":"reviewer_approved","task_id":"33"}`,
	})

	h := &BreezingSignalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp injectorResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	if !strings.Contains(resp.SystemMessage, "[SIGNAL:reviewer_approved]") {
		t.Errorf("expected reviewer_approved in message")
	}
	if !strings.Contains(resp.SystemMessage, "33") {
		t.Errorf("expected task_id in message")
	}
}

func TestBreezingSignalInjector_EscalationRequiredSignal(t *testing.T) {
	dir := t.TempDir()
	writeActiveFile(t, dir)
	writeSignalsFile(t, dir, []string{
		`{"signal":"escalation_required","task_id":"34","reason":"3回失敗"}`,
	})

	h := &BreezingSignalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp injectorResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	if !strings.Contains(resp.SystemMessage, "[SIGNAL:escalation_required]") {
		t.Errorf("expected escalation_required in message")
	}
}

func TestBreezingSignalInjector_UnknownSignal(t *testing.T) {
	dir := t.TempDir()
	writeActiveFile(t, dir)
	writeSignalsFile(t, dir, []string{
		`{"signal":"custom_signal","data":"some_value"}`,
	})

	h := &BreezingSignalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp injectorResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	// 未知のシグナルはそのまま通知
	if !strings.Contains(resp.SystemMessage, "[SIGNAL:custom_signal]") {
		t.Errorf("expected custom_signal in message, got: %s", resp.SystemMessage)
	}
}

func TestBreezingSignalInjector_SignalsMarkedConsumedAfterInjection(t *testing.T) {
	dir := t.TempDir()
	writeActiveFile(t, dir)
	signalsFile := writeSignalsFile(t, dir, []string{
		`{"signal":"reviewer_approved","task_id":"10"}`,
		`{"signal":"reviewer_approved","task_id":"11","consumed_at":"2026-01-01T00:00:00Z"}`,
	})

	h := &BreezingSignalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 注入後、シグナルが consumed_at マーク済みになっているか確認
	signals, err := h.readUnconsumedSignals(signalsFile)
	if err != nil {
		t.Fatalf("error reading signals: %v", err)
	}
	if len(signals) != 0 {
		t.Errorf("expected all signals consumed after injection, got %d unconsumed", len(signals))
	}
}

func TestBreezingSignalInjector_MultipleUnconsumedSignals(t *testing.T) {
	dir := t.TempDir()
	writeActiveFile(t, dir)
	writeSignalsFile(t, dir, []string{
		`{"signal":"ci_failure_detected","conclusion":"failure"}`,
		`{"signal":"reviewer_approved","task_id":"5"}`,
		`{"signal":"retake_requested","task_id":"6","reason":"型エラー"}`,
	})

	h := &BreezingSignalInjectorHandler{ProjectRoot: dir}

	var out bytes.Buffer
	err := h.Handle(strings.NewReader(`{}`), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp injectorResponse
	if err := json.Unmarshal(bytes.TrimRight(out.Bytes(), "\n"), &resp); err != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}

	// 3件のシグナルがヘッダーに反映されること
	if !strings.Contains(resp.SystemMessage, "3 件") {
		t.Errorf("expected '3 件' in header, got: %s", resp.SystemMessage)
	}
}
