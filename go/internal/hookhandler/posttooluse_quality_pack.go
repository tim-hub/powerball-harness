package hookhandler

// posttooluse_quality_pack.go
// posttooluse-quality-pack.sh の Go 移植。
//
// PostToolUse Write/Edit 後にオプショナル品質チェックを実行する:
//   - .claude-code-harness.config.yaml から設定を読み込む
//   - Prettier チェック（warn/run モード）
//   - tsc --noEmit チェック（warn/run モード）
//   - console.log 検出
//   - 各チェック結果を systemMessage（additionalContext）に集約
//   - 設定が無効/未設定の場合はスキップ

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// qualityPackInput は PostToolUse フックの stdin JSON。
type qualityPackInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
	ToolResponse struct {
		FilePath string `json:"filePath"`
	} `json:"tool_response"`
	CWD string `json:"cwd"`
}

// qualityPackConfig は .claude-code-harness.config.yaml の quality_pack セクション。
type qualityPackConfig struct {
	Enabled    bool   // enabled: true/false（デフォルト false）
	Mode       string // warn または run（デフォルト warn）
	Prettier   bool   // prettier: true/false（デフォルト true）
	TSC        bool   // tsc: true/false（デフォルト true）
	ConsoleLog bool   // console_log: true/false（デフォルト true）
}

// HandlePostToolUseQualityPack は posttooluse-quality-pack.sh の Go 移植。
//
// PostToolUse Write/Edit イベントで呼び出され、品質チェックを実行する。
// .claude-code-harness.config.yaml の quality_pack.enabled が true の場合のみ動作する。
func HandlePostToolUseQualityPack(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}

	var input qualityPackInput
	if jsonErr := json.Unmarshal(data, &input); jsonErr != nil {
		return nil
	}

	// Write/Edit のみ対象
	if input.ToolName != "Write" && input.ToolName != "Edit" {
		return nil
	}

	// ファイルパスを取得
	filePath := input.ToolInput.FilePath
	if filePath == "" {
		filePath = input.ToolResponse.FilePath
	}
	if filePath == "" {
		return nil
	}

	// CWD があれば相対パスに変換
	cwd := input.CWD
	if cwd != "" && strings.HasPrefix(filePath, cwd+"/") {
		filePath = strings.TrimPrefix(filePath, cwd+"/")
	}

	// JS/TS ファイルのみ対象
	if !isJSTSFile(filePath) {
		return nil
	}

	// 除外パスのチェック
	if isExcludedPath(filePath) {
		return nil
	}

	// 設定を読み込む
	cfg := readQualityPackConfig(".claude-code-harness.config.yaml")
	if !cfg.Enabled {
		return nil
	}

	// 品質チェックを実行して feedback を収集
	var feedbacks []string

	if cfg.Prettier {
		msg := runPrettierCheck(filePath, cfg.Mode)
		if msg != "" {
			feedbacks = append(feedbacks, msg)
		}
	}

	if cfg.TSC {
		msg := runTSCCheck(cfg.Mode)
		if msg != "" {
			feedbacks = append(feedbacks, msg)
		}
	}

	if cfg.ConsoleLog {
		msg := detectConsoleLogs(filePath)
		if msg != "" {
			feedbacks = append(feedbacks, msg)
		}
	}

	if len(feedbacks) == 0 {
		return nil
	}

	// フィードバックを additionalContext にまとめて出力
	combined := "Quality Pack (PostToolUse)\n" + strings.Join(feedbacks, "\n")

	o := postToolOutput{}
	o.HookSpecificOutput.HookEventName = "PostToolUse"
	o.HookSpecificOutput.AdditionalContext = combined
	return writeJSON(out, o)
}

// isJSTSFile は JS/TS ファイルかどうかを判定する。
func isJSTSFile(filePath string) bool {
	lower := strings.ToLower(filePath)
	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// isExcludedPath は除外パスかどうかを判定する。
// bash の case 文と同等: .claude/*, docs/*, templates/*, benchmarks/*, node_modules/*, .git/*
func isExcludedPath(filePath string) bool {
	excludePrefixes := []string{
		".claude/",
		"docs/",
		"templates/",
		"benchmarks/",
		"node_modules/",
		".git/",
	}
	for _, prefix := range excludePrefixes {
		if strings.HasPrefix(filePath, prefix) {
			return true
		}
	}
	return false
}

// readQualityPackConfig は .claude-code-harness.config.yaml から quality_pack セクションを読む。
// YAML パーサーなしで実装（bash の awk と同等のロジック）。
func readQualityPackConfig(configPath string) qualityPackConfig {
	cfg := qualityPackConfig{
		Enabled:    false,
		Mode:       "warn",
		Prettier:   true,
		TSC:        true,
		ConsoleLog: true,
	}

	f, err := os.Open(configPath)
	if err != nil {
		return cfg // ファイルが存在しない場合はデフォルト（無効）
	}
	defer f.Close()

	inQualityPack := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// quality_pack: セクションの開始検出
		if strings.TrimSpace(line) == "quality_pack:" {
			inQualityPack = true
			continue
		}

		// 別のトップレベルセクションが始まったら終了
		if inQualityPack && len(line) > 0 && line[0] != ' ' && line[0] != '\t' && line[0] != '#' {
			break
		}

		if !inQualityPack {
			continue
		}

		// キー: 値 のパース（インデント付き）
		trimmed := strings.TrimSpace(line)
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)

		switch key {
		case "enabled":
			cfg.Enabled = val == "true"
		case "mode":
			cfg.Mode = val
		case "prettier":
			cfg.Prettier = val != "false"
		case "tsc":
			cfg.TSC = val != "false"
		case "console_log":
			cfg.ConsoleLog = val != "false"
		}
	}

	return cfg
}

// runPrettierCheck は Prettier チェックを実行する。
// mode=run: prettier --write を実行
// mode=warn: 推奨メッセージを返す
func runPrettierCheck(filePath, mode string) string {
	if mode == "run" {
		prettierBin := "./node_modules/.bin/prettier"
		if _, statErr := os.Stat(prettierBin); statErr != nil {
			return "Prettier: 未実行（prettier が見つかりません）"
		}
		cmd := exec.Command(prettierBin, "--write", filePath)
		var errBuf bytes.Buffer
		cmd.Stderr = &errBuf
		if runErr := cmd.Run(); runErr != nil {
			return "Prettier: 未実行（prettier が見つかりません）"
		}
		return "Prettier: 実行済み"
	}
	// warn モード
	return fmt.Sprintf("Prettier: 推奨（例: npx prettier --write \"%s\"）", filePath)
}

// runTSCCheck は TypeScript 型チェックを実行する。
// mode=run: tsc --noEmit を実行
// mode=warn: 推奨メッセージを返す
func runTSCCheck(mode string) string {
	if mode == "run" {
		// tsconfig.json の存在確認
		if _, statErr := os.Stat("tsconfig.json"); statErr != nil {
			return "tsc --noEmit: 未実行（tsconfig/tsc 未検出）"
		}
		tscBin := "./node_modules/.bin/tsc"
		if _, statErr := os.Stat(tscBin); statErr != nil {
			return "tsc --noEmit: 未実行（tsconfig/tsc 未検出）"
		}
		cmd := exec.Command(tscBin, "--noEmit")
		if runErr := cmd.Run(); runErr != nil {
			return "tsc --noEmit: 未実行（tsconfig/tsc 未検出）"
		}
		return "tsc --noEmit: 実行済み"
	}
	// warn モード
	return "tsc --noEmit: 推奨"
}

// detectConsoleLogs はファイル内の console.log の個数を検出する。
func detectConsoleLogs(filePath string) string {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	count := 0
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "console.log") {
			count++
		}
	}

	if count > 0 {
		return fmt.Sprintf("console.log が %d 件見つかりました", count)
	}
	return ""
}
