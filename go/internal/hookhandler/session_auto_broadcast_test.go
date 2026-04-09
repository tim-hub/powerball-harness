package hookhandler

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleSessionAutoBroadcast_NoInput(t *testing.T) {
	var out bytes.Buffer
	err := HandleSessionAutoBroadcast(strings.NewReader(""), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v, raw: %s", jsonErr, out.String())
	}
	if result.HookSpecificOutput.HookEventName != "PostToolUse" {
		t.Errorf("expected hookEventName=PostToolUse, got %q", result.HookSpecificOutput.HookEventName)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty additionalContext, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandleSessionAutoBroadcast_NoFilePath(t *testing.T) {
	input := `{"tool_input":{}}`
	var out bytes.Buffer
	err := HandleSessionAutoBroadcast(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty context for no file_path, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandleSessionAutoBroadcast_NoPatternMatch(t *testing.T) {
	input := `{"tool_input":{"file_path":"src/utils/helper.go"}}`
	var out bytes.Buffer
	err := HandleSessionAutoBroadcast(strings.NewReader(input), &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty context for non-matching file, got %q", result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandleSessionAutoBroadcast_MatchesSrcAPI(t *testing.T) {
	// テスト用の一時ディレクトリに移動
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	input := `{"tool_input":{"file_path":"src/api/users.ts"}}`
	var out bytes.Buffer
	handlerErr := HandleSessionAutoBroadcast(strings.NewReader(input), &out)
	if handlerErr != nil {
		t.Fatalf("unexpected error: %v", handlerErr)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v, raw: %s", jsonErr, out.String())
	}

	// additionalContext にファイル名が含まれること
	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "users.ts") {
		t.Errorf("expected additionalContext to contain 'users.ts', got %q",
			result.HookSpecificOutput.AdditionalContext)
	}
	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "自動ブロードキャスト") {
		t.Errorf("expected additionalContext to contain '自動ブロードキャスト', got %q",
			result.HookSpecificOutput.AdditionalContext)
	}

	// broadcast.md が作成されていること
	broadcastFile := filepath.Join(".claude", "state", "broadcast.md")
	data, readErr := os.ReadFile(broadcastFile)
	if readErr != nil {
		t.Fatalf("broadcast.md not created: %v", readErr)
	}
	if !strings.Contains(string(data), "src/api/users.ts") {
		t.Errorf("broadcast.md should contain file path, got: %s", string(data))
	}
}

func TestHandleSessionAutoBroadcast_MatchesSchemaPrisma(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	input := `{"tool_input":{"file_path":"prisma/schema.prisma"}}`
	var out bytes.Buffer
	if handlerErr := HandleSessionAutoBroadcast(strings.NewReader(input), &out); handlerErr != nil {
		t.Fatalf("unexpected error: %v", handlerErr)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "schema.prisma") {
		t.Errorf("expected file name in additionalContext, got %q",
			result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandleSessionAutoBroadcast_MatchesPathField(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// file_path の代わりに path フィールドを使う
	input := `{"tool_input":{"path":"src/types/user.ts"}}`
	var out bytes.Buffer
	if handlerErr := HandleSessionAutoBroadcast(strings.NewReader(input), &out); handlerErr != nil {
		t.Fatalf("unexpected error: %v", handlerErr)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "user.ts") {
		t.Errorf("expected 'user.ts' in additionalContext, got %q",
			result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandleSessionAutoBroadcast_DisabledByConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// 設定ファイルで無効化
	configDir := filepath.Join(".claude", "sessions")
	if mkdirErr := os.MkdirAll(configDir, 0o755); mkdirErr != nil {
		t.Fatal(mkdirErr)
	}
	if writeErr := os.WriteFile(
		filepath.Join(configDir, "auto-broadcast.json"),
		[]byte(`{"enabled":false}`),
		0o644,
	); writeErr != nil {
		t.Fatal(writeErr)
	}

	input := `{"tool_input":{"file_path":"src/api/users.ts"}}`
	var out bytes.Buffer
	if handlerErr := HandleSessionAutoBroadcast(strings.NewReader(input), &out); handlerErr != nil {
		t.Fatalf("unexpected error: %v", handlerErr)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	// 無効な場合は追加コンテキストなし
	if result.HookSpecificOutput.AdditionalContext != "" {
		t.Errorf("expected empty context when disabled, got %q",
			result.HookSpecificOutput.AdditionalContext)
	}
}

func TestHandleSessionAutoBroadcast_CustomPattern(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// カスタムパターンを設定
	configDir := filepath.Join(".claude", "sessions")
	if mkdirErr := os.MkdirAll(configDir, 0o755); mkdirErr != nil {
		t.Fatal(mkdirErr)
	}
	if writeErr := os.WriteFile(
		filepath.Join(configDir, "auto-broadcast.json"),
		[]byte(`{"enabled":true,"patterns":["custom/contracts/"]}`),
		0o644,
	); writeErr != nil {
		t.Fatal(writeErr)
	}

	input := `{"tool_input":{"file_path":"custom/contracts/order.ts"}}`
	var out bytes.Buffer
	if handlerErr := HandleSessionAutoBroadcast(strings.NewReader(input), &out); handlerErr != nil {
		t.Fatalf("unexpected error: %v", handlerErr)
	}

	var result postToolOutput
	if jsonErr := json.Unmarshal(out.Bytes(), &result); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v", jsonErr)
	}
	if !strings.Contains(result.HookSpecificOutput.AdditionalContext, "order.ts") {
		t.Errorf("expected 'order.ts' in additionalContext (custom pattern), got %q",
			result.HookSpecificOutput.AdditionalContext)
	}
}
