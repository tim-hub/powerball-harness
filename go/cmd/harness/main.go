// harness is the Claude Code Harness v4 CLI.
//
// Phase 0 implements the hook subcommands:
//
//	harness hook pre-tool     — PreToolUse guardrail evaluation
//	harness hook post-tool    — PostToolUse tampering/security checks
//	harness hook permission   — PermissionRequest auto-approval
//	harness version           — Print version
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

	"github.com/Chachamaru127/claude-code-harness/go/internal/guard"
	"github.com/Chachamaru127/claude-code-harness/go/internal/hook"
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
	fmt.Fprintln(os.Stderr, "  hook pre-tool      Evaluate PreToolUse guardrails")
	fmt.Fprintln(os.Stderr, "  hook post-tool     Evaluate PostToolUse checks")
	fmt.Fprintln(os.Stderr, "  hook permission    Evaluate PermissionRequest")
	fmt.Fprintln(os.Stderr, "  init [root]        Create harness.toml template in project root")
	fmt.Fprintln(os.Stderr, "  sync [root]        Generate CC files from harness.toml")
	fmt.Fprintln(os.Stderr, "  validate [skills|agents|all] [root]  Validate SKILL.md / agent frontmatter")
	fmt.Fprintln(os.Stderr, "  doctor [--migration] [root]          Health check; --migration shows hook migration status")
	fmt.Fprintln(os.Stderr, "  version            Print version")
}

func runHook(hookType string) {
	input, err := hook.ReadInput(os.Stdin)
	if err != nil {
		// Empty input or parse error → safe approve
		result := hook.SafeResult(err)
		hook.WriteResult(os.Stdout, result)
		return
	}

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
