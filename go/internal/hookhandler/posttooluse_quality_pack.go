package hookhandler

// posttooluse_quality_pack.go
// Go port of posttooluse-quality-pack.sh.
//
// Runs optional quality checks after PostToolUse Write/Edit:
//   - Reads configuration from .claude-code-harness.config.yaml
//   - Prettier check (warn/run mode)
//   - tsc --noEmit check (warn/run mode)
//   - console.log detection
//   - Aggregates each check result into systemMessage (additionalContext)
//   - Skipped if the configuration is disabled/not set

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

// qualityPackInput is the stdin JSON for the PostToolUse hook.
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

// qualityPackConfig is the quality_pack section of .claude-code-harness.config.yaml.
type qualityPackConfig struct {
	Enabled    bool   // enabled: true/false (default false)
	Mode       string // warn or run (default warn)
	Prettier   bool   // prettier: true/false (default true)
	TSC        bool   // tsc: true/false (default true)
	ConsoleLog bool   // console_log: true/false (default true)
}

// HandlePostToolUseQualityPack is the Go port of posttooluse-quality-pack.sh.
//
// Called on PostToolUse Write/Edit events to run quality checks.
// Only runs when quality_pack.enabled is true in .claude-code-harness.config.yaml.
func HandlePostToolUseQualityPack(in io.Reader, out io.Writer) error {
	data, err := io.ReadAll(in)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}

	var input qualityPackInput
	if jsonErr := json.Unmarshal(data, &input); jsonErr != nil {
		return nil
	}

	// Only process Write/Edit tools
	if input.ToolName != "Write" && input.ToolName != "Edit" {
		return nil
	}

	// Get the file path
	filePath := input.ToolInput.FilePath
	if filePath == "" {
		filePath = input.ToolResponse.FilePath
	}
	if filePath == "" {
		return nil
	}

	// Convert to relative path if CWD is available
	cwd := input.CWD
	if cwd != "" && strings.HasPrefix(filePath, cwd+"/") {
		filePath = strings.TrimPrefix(filePath, cwd+"/")
	}

	// Only process JS/TS files
	if !isJSTSFile(filePath) {
		return nil
	}

	// Check for excluded paths
	if isExcludedPath(filePath) {
		return nil
	}

	// Load configuration
	cfg := readQualityPackConfig(".claude-code-harness.config.yaml")
	if !cfg.Enabled {
		return nil
	}

	// Run quality checks and collect feedback
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

	// Aggregate feedback into additionalContext and output
	combined := "Quality Pack (PostToolUse)\n" + strings.Join(feedbacks, "\n")

	o := postToolOutput{}
	o.HookSpecificOutput.HookEventName = "PostToolUse"
	o.HookSpecificOutput.AdditionalContext = combined
	return writeJSON(out, o)
}

// isJSTSFile returns true if the file is a JS/TS file.
func isJSTSFile(filePath string) bool {
	lower := strings.ToLower(filePath)
	for _, ext := range []string{".ts", ".tsx", ".js", ".jsx"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// isExcludedPath returns true if the file path matches an excluded prefix.
// Equivalent to the bash case statement: .claude/*, docs/*, templates/*, benchmarks/*, node_modules/*, .git/*
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

// readQualityPackConfig reads the quality_pack section from .claude-code-harness.config.yaml.
// Implemented without a YAML parser (equivalent logic to bash awk).
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
		return cfg // file not found: return defaults (disabled)
	}
	defer f.Close()

	inQualityPack := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// detect start of quality_pack: section
		if strings.TrimSpace(line) == "quality_pack:" {
			inQualityPack = true
			continue
		}

		// stop when another top-level section begins
		if inQualityPack && len(line) > 0 && line[0] != ' ' && line[0] != '\t' && line[0] != '#' {
			break
		}

		if !inQualityPack {
			continue
		}

		// parse key: value pairs (indented)
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

// runPrettierCheck runs the Prettier check.
// mode=run: executes prettier --write
// mode=warn: returns a recommendation message
func runPrettierCheck(filePath, mode string) string {
	if mode == "run" {
		prettierBin := "./node_modules/.bin/prettier"
		if _, statErr := os.Stat(prettierBin); statErr != nil {
			return "Prettier: not run (prettier not found)"
		}
		cmd := exec.Command(prettierBin, "--write", filePath)
		var errBuf bytes.Buffer
		cmd.Stderr = &errBuf
		if runErr := cmd.Run(); runErr != nil {
			return "Prettier: not run (prettier not found)"
		}
		return "Prettier: executed"
	}
	// warn mode
	return fmt.Sprintf("Prettier: recommended (e.g.: npx prettier --write \"%s\")", filePath)
}

// runTSCCheck runs the TypeScript type check.
// mode=run: executes tsc --noEmit
// mode=warn: returns a recommendation message
func runTSCCheck(mode string) string {
	if mode == "run" {
		// check for tsconfig.json
		if _, statErr := os.Stat("tsconfig.json"); statErr != nil {
			return "tsc --noEmit: not run (tsconfig/tsc not found)"
		}
		tscBin := "./node_modules/.bin/tsc"
		if _, statErr := os.Stat(tscBin); statErr != nil {
			return "tsc --noEmit: not run (tsconfig/tsc not found)"
		}
		cmd := exec.Command(tscBin, "--noEmit")
		if runErr := cmd.Run(); runErr != nil {
			return "tsc --noEmit: not run (tsconfig/tsc not found)"
		}
		return "tsc --noEmit: executed"
	}
	// warn mode
	return "tsc --noEmit: recommended"
}

// detectConsoleLogs counts the number of console.log occurrences in the file.
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
		return fmt.Sprintf("found %d console.log occurrence(s)", count)
	}
	return ""
}
