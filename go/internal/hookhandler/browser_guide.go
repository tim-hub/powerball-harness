package hookhandler

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

// agentBrowserContext is the additionalContext injected when agent-browser is installed.
const agentBrowserContext = `💡 **agent-browser を先に試すことを推奨します**

agent-browser は AI エージェント向けに最適化されたブラウザ自動化ツールです。

` + "```bash" + `
# 基本的な使い方
agent-browser open <url>
agent-browser snapshot -i -c  # AI 向けスナップショット
agent-browser click @e1        # 要素参照でクリック
` + "```" + `

現在の MCP ツールも使用可能ですが、agent-browser の方がシンプルで高速です。

詳細: ` + "`docs/OPTIONAL_PLUGINS.md`"

// agentBrowserLookupFn is the function used to check if agent-browser is
// installed. It is a package-level variable to allow injection in tests.
var agentBrowserLookupFn = agentBrowserInstalled

// HandleBrowserGuide ports pretooluse-browser-guide.sh.
//
// When agent-browser is installed, outputs an additionalContext recommendation.
// The hook matcher (hooks.json) already filters to MCP browser tool names, so
// no additional tool-name check is needed here.
func HandleBrowserGuide(in io.Reader, out io.Writer) error {
	// Read stdin — if empty, nothing to do (matches bash `[ -z "$INPUT" ] && exit 0`).
	data, err := io.ReadAll(in)
	if err != nil {
		return nil //nolint:nilerr // fail-open
	}
	if len(data) == 0 {
		return nil
	}

	// Only recommend when agent-browser is installed.
	if !agentBrowserLookupFn() {
		return nil
	}

	output := preToolAllowOutput{}
	output.HookSpecificOutput.HookEventName = "PreToolUse"
	output.HookSpecificOutput.AdditionalContext = agentBrowserContext

	out_data, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("marshaling output: %w", err)
	}
	_, err = fmt.Fprintf(out, "%s\n", out_data)
	return err
}

// agentBrowserInstalled checks if the agent-browser binary is on PATH.
func agentBrowserInstalled() bool {
	_, err := exec.LookPath("agent-browser")
	return err == nil
}
