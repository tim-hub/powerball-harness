// harness is the Claude Code Harness v4 CLI.
//
// Phase 0 implements the hook subcommands:
//
//	harness hook pre-tool          — PreToolUse guardrail evaluation
//	harness hook post-tool         — PostToolUse tampering/security checks
//	harness hook permission        — PermissionRequest auto-approval
//	harness hook session-start     — SessionStart env setup
//	harness hook post-tool-failure — PostToolUseFailure counter & escalation
//	harness hook post-compact      — PostCompact WIP context re-injection
//	harness hook notification      — Notification event logging
//	harness hook permission-denied — PermissionDenied event logging
//	harness hook session-init      — SessionStart: session initialization + Plans.md summary
//	harness hook session-cleanup   — SessionEnd: temp file cleanup
//	harness hook session-monitor   — SessionStart: project state collection + session.json
//	harness hook session-summary   — Stop: session summary to session-log.md
//	harness hook ci-status         — PostToolUse: CI status check after push/PR
//	harness hook subagent-start    — SubagentStart: エージェント起動追跡
//	harness hook subagent-stop     — SubagentStop: エージェント停止追跡
//	harness evidence collect       — Collect evidence (test results, build logs)
//	harness status                 — 全追跡エージェントの状態表示
//	harness version                — Print version
//
// Usage in hooks.json:
//
//	"command": "harness hook pre-tool"
//
// The binary reads JSON from stdin and writes JSON to stdout.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Chachamaru127/claude-code-harness/go/internal/ci"
	"github.com/Chachamaru127/claude-code-harness/go/internal/event"
	"github.com/Chachamaru127/claude-code-harness/go/internal/guard"
	"github.com/Chachamaru127/claude-code-harness/go/internal/hook"
	"github.com/Chachamaru127/claude-code-harness/go/internal/lifecycle"
	"github.com/Chachamaru127/claude-code-harness/go/internal/session"
	"github.com/Chachamaru127/claude-code-harness/go/internal/state"
	"github.com/Chachamaru127/claude-code-harness/go/pkg/protocol"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "hook":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: harness hook <pre-tool|post-tool|permission>")
			os.Exit(1)
		}
		runHook(os.Args[2])
	case "evidence":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: harness evidence <collect>")
			os.Exit(1)
		}
		runEvidence(os.Args[2:])
	case "status":
		runStatus(os.Args[2:])
	case "init":
		runInit(os.Args[2:])
	case "sync":
		runSync(os.Args[2:])
	case "validate":
		runValidate(os.Args[2:])
	case "doctor":
		runDoctor(os.Args[2:])
	case "version":
		fmt.Println(version)
	case "--version", "-v":
		fmt.Println(version)
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: harness <command>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  hook pre-tool           Evaluate PreToolUse guardrails")
	fmt.Fprintln(os.Stderr, "  hook post-tool          Evaluate PostToolUse checks")
	fmt.Fprintln(os.Stderr, "  hook permission         Evaluate PermissionRequest")
	fmt.Fprintln(os.Stderr, "  hook session-start      SessionStart env setup (writes CLAUDE_ENV_FILE)")
	fmt.Fprintln(os.Stderr, "  hook post-tool-failure  PostToolUseFailure counter & escalation")
	fmt.Fprintln(os.Stderr, "  hook post-compact       PostCompact WIP context re-injection")
	fmt.Fprintln(os.Stderr, "  hook notification       Notification event logging")
	fmt.Fprintln(os.Stderr, "  hook permission-denied  PermissionDenied event logging + Worker retry")
	fmt.Fprintln(os.Stderr, "  hook session-init       SessionStart: session initialization + Plans.md summary")
	fmt.Fprintln(os.Stderr, "  hook session-cleanup    SessionEnd: temp file cleanup")
	fmt.Fprintln(os.Stderr, "  hook session-monitor    SessionStart: project state collection + session.json")
	fmt.Fprintln(os.Stderr, "  hook session-summary    Stop: session summary to session-log.md")
	fmt.Fprintln(os.Stderr, "  hook ci-status          PostToolUse: CI status check after push/PR")
	fmt.Fprintln(os.Stderr, "  hook subagent-start     SubagentStart: track agent lifecycle start")
	fmt.Fprintln(os.Stderr, "  hook subagent-stop      SubagentStop: track agent lifecycle stop")
	fmt.Fprintln(os.Stderr, "  evidence collect        Collect evidence (test results, build logs) from stdin")
	fmt.Fprintln(os.Stderr, "    --label <label>       Evidence label (default: general)")
	fmt.Fprintln(os.Stderr, "    --file <path>         Read content from file instead of stdin")
	fmt.Fprintln(os.Stderr, "  status                  Show all tracked agent states")
	fmt.Fprintln(os.Stderr, "  init [root]             Create harness.toml template in project root")
	fmt.Fprintln(os.Stderr, "  sync [root]             Generate CC files from harness.toml")
	fmt.Fprintln(os.Stderr, "  validate [skills|agents|all] [root]  Validate SKILL.md / agent frontmatter")
	fmt.Fprintln(os.Stderr, "  doctor [--migration] [root]          Health check; --migration shows hook migration status")
	fmt.Fprintln(os.Stderr, "  version                 Print version")
}

// runEvidence は evidence サブコマンドを実行する。
func runEvidence(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: harness evidence <collect>")
		os.Exit(1)
	}

	switch args[0] {
	case "collect":
		runEvidenceCollect(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown evidence subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

// runEvidenceCollect は evidence collect サブコマンドを実行する。
// stdin からコンテンツを読み取って .claude/state/evidence/{label}/ に保存する。
func runEvidenceCollect(args []string) {
	var label string
	var contentFile string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--label":
			if i+1 < len(args) {
				i++
				label = args[i]
			}
		case "--file":
			if i+1 < len(args) {
				i++
				contentFile = args[i]
			}
		}
	}

	c := &ci.EvidenceCollector{}
	opts := ci.CollectOptions{
		Label:       label,
		ContentFile: contentFile,
	}

	if contentFile != "" {
		// ファイルから収集する場合
		result := c.Collect(opts)
		if result.Error != "" {
			fmt.Fprintln(os.Stderr, "evidence collect error:", result.Error)
			os.Exit(1)
		}
		fmt.Println(result.SavedPath)
		return
	}

	// stdin から収集する場合
	if err := c.CollectFromStdin(os.Stdin, os.Stdout, opts); err != nil {
		fmt.Fprintln(os.Stderr, "evidence collect error:", err)
		os.Exit(1)
	}
}

func runHook(hookType string) {
	switch hookType {
	// --- event handlers (no tool_name validation) ---
	case "session-start":
		h := &event.SessionEnvHandler{}
		if err := h.Handle(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "session-start handler error: %v\n", err)
		}
	case "post-tool-failure":
		h := &event.PostToolFailureHandler{}
		if err := h.Handle(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "post-tool-failure handler error: %v\n", err)
		}
	case "post-compact":
		h := &event.PostCompactHandler{}
		if err := h.Handle(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "post-compact handler error: %v\n", err)
		}
	case "notification":
		h := &event.NotificationHandler{}
		if err := h.Handle(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "notification handler error: %v\n", err)
		}
	case "permission-denied":
		h := &event.PermissionDeniedHandler{}
		if err := h.Handle(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "permission-denied handler error: %v\n", err)
		}
	// --- CI ハンドラ ---
	case "ci-status":
		h := &ci.CIStatusHandler{}
		if err := h.Handle(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "ci-status handler error: %v\n", err)
		}
	// --- subagent lifecycle handlers ---
	case "subagent-start":
		runSubagentStart()
	case "subagent-stop":
		runSubagentStop()
	// --- session handlers ---
	case "session-init":
		h := &session.InitHandler{}
		if err := h.Handle(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "session-init handler error: %v\n", err)
		}
	case "session-cleanup":
		h := &session.CleanupHandler{}
		if err := h.Handle(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "session-cleanup handler error: %v\n", err)
		}
	case "session-monitor":
		h := &session.MonitorHandler{}
		if err := h.Handle(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "session-monitor handler error: %v\n", err)
		}
	case "session-summary":
		h := &session.SummaryHandler{}
		if err := h.Handle(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "session-summary handler error: %v\n", err)
		}
	default:
		// guard-fastpath handlers require tool_name validation
		input, err := hook.ReadInput(os.Stdin)
		if err != nil {
			// Empty input or parse error → safe approve
			result := hook.SafeResult(err)
			hook.WriteResult(os.Stdout, result)
			return
		}
		runGuardHook(hookType, input)
	}
}

func runGuardHook(hookType string, input protocol.HookInput) {
	switch hookType {
	case "pre-tool":
		runPreTool(input)
	case "post-tool":
		runPostTool(input)
	case "permission":
		runPermission(input)
	default:
		fmt.Fprintf(os.Stderr, "Unknown hook type: %s\n", hookType)
		// Safe fallback
		hook.WriteResult(os.Stdout, hook.SafeResult(fmt.Errorf("unknown hook type: %s", hookType)))
	}
}

func runPreTool(input protocol.HookInput) {
	result := guard.EvaluatePreTool(input)
	output, exitCode := guard.FormatPreToolResult(result)

	if output != nil {
		hook.WriteJSON(os.Stdout, output)
	}

	os.Exit(exitCode)
}

func runPostTool(input protocol.HookInput) {
	result := guard.EvaluatePostTool(input)

	// PostToolUse: if there's a systemMessage, wrap in hookSpecificOutput
	if result.SystemMessage != "" {
		out := protocol.PostToolOutput{
			HookEventName:    "PostToolUse",
			AdditionalContext: result.SystemMessage,
		}
		hook.WriteJSON(os.Stdout, out)
		return
	}

	// No output for pure approve
}

func runPermission(input protocol.HookInput) {
	_, permOutput := guard.EvaluatePermission(input)

	if permOutput != nil {
		hook.WriteJSON(os.Stdout, permOutput)
		return
	}

	// No output = pass through to user prompt
}

// openTracker は SQLite ストアを使った AgentTracker を開いて返す。
// DB 開放に失敗した場合はインメモリ tracker (store=nil) にフォールバックする。
func openTracker() (*lifecycle.AgentTracker, func()) {
	dbPath := state.ResolveStatePath("")
	store, err := state.NewHarnessStore(dbPath)
	if err != nil {
		// DB が使えなくてもフックは継続できる（インメモリのみ）
		return lifecycle.NewAgentTracker(nil), func() {}
	}
	tracker := lifecycle.NewAgentTracker(store)
	return tracker, func() { store.Close() }
}

// runSubagentStart は SubagentStart フックを処理する。
// stdin から HookInput を読み取り、AgentTracker にエージェントを登録する。
func runSubagentStart() {
	input, err := hook.ReadInput(os.Stdin)
	if err != nil {
		// 入力パースエラーは無視して通過（フックの安全原則）
		return
	}

	tracker, cleanup := openTracker()
	defer cleanup()

	if err := tracker.HandleStart(input); err != nil {
		fmt.Fprintf(os.Stderr, "subagent-start handler error: %v\n", err)
	}
	// SubagentStart は出力不要（通過フック）
}

// runSubagentStop は SubagentStop フックを処理する。
// stdin から HookInput を読み取り、AgentTracker にエージェントの停止を記録する。
func runSubagentStop() {
	input, err := hook.ReadInput(os.Stdin)
	if err != nil {
		return
	}

	tracker, cleanup := openTracker()
	defer cleanup()

	if err := tracker.HandleStop(input); err != nil {
		fmt.Fprintf(os.Stderr, "subagent-stop handler error: %v\n", err)
	}
	// SubagentStop は出力不要（通過フック）
}

// runStatus は全追跡中エージェントの状態テーブルを表示する。
// SQLite ストアが利用可能な場合、永続化済みレコードを表示する。
func runStatus(_ []string) {
	dbPath := state.ResolveStatePath("")
	store, err := state.NewHarnessStore(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "status: DB open error: %v\n", err)
		fmt.Println("Tracked Agents: (DB unavailable)")
		return
	}
	defer store.Close()

	records, err := store.ListAgentStates(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "status: list error: %v\n", err)
		return
	}

	printStatusTable(records)
}

// printStatusTable はエージェント状態テーブルを stdout に表示する。
func printStatusTable(records []state.AgentStateRecord) {
	if len(records) == 0 {
		fmt.Println("Tracked Agents: (none)")
		return
	}

	fmt.Println("Tracked Agents:")
	fmt.Printf("%-12s  %-12s  %-10s  %-8s  %s\n",
		"Agent ID", "Type", "State", "Duration", "Recovery")
	fmt.Printf("%-12s  %-12s  %-10s  %-8s  %s\n",
		"------------", "------------", "----------", "--------", "--------")

	var active, failed, completed int
	for _, rec := range records {
		dur := formatDuration(rec)
		recovery := formatRecovery(rec)
		shortID := rec.AgentID
		if len(shortID) > 12 {
			shortID = shortID[:7] + "..."
		}
		fmt.Printf("%-12s  %-12s  %-10s  %-8s  %s\n",
			shortID, rec.AgentType, rec.State, dur, recovery)

		switch rec.State {
		case "RUNNING", "SPAWNING", "REVIEWING", "APPROVED", "RECOVERING":
			active++
		case "FAILED", "ABORTED", "STALE":
			failed++
		default:
			completed++
		}
	}

	fmt.Printf("\nTotal: %d active, %d failed, %d completed\n", active, failed, completed)
}

// formatDuration は AgentStateRecord から経過時間の文字列を生成する。
// stopped_at があればその時刻まで、なければ現在時刻との差を返す。
func formatDuration(rec state.AgentStateRecord) string {
	startStr := rec.StartedAt
	if startStr == "" {
		return "-"
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return "-"
	}

	var end time.Time
	if rec.StoppedAt != nil {
		end, err = time.Parse(time.RFC3339, *rec.StoppedAt)
		if err != nil {
			end = time.Now()
		}
	} else {
		end = time.Now()
	}

	d := end.Sub(start)
	if d < 0 {
		d = 0
	}

	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh%02dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// formatRecovery はリカバリ試行回数を "N/3" 形式で返す。
// リカバリがない場合は "-" を返す。
func formatRecovery(rec state.AgentStateRecord) string {
	if rec.RecoveryAttempts == 0 {
		return "-"
	}
	return fmt.Sprintf("%d/3", rec.RecoveryAttempts)
}
