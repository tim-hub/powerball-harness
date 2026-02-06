import type { ExperimentConfig } from "@vercel/agent-eval";

export default {
  agent: "claude-code",
  model: "claude-haiku-4-5-20251001",
  runs: 1,
  earlyExit: true,
  timeout: 600,
  scripts: ["test"],
  sandbox: "docker",
  evals: ["task-01"],
  setup: async (sandbox) => {
    await sandbox.writeFiles({
      ".claude/settings.json": JSON.stringify(
        {
          env: { CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS: "1" },
        },
        null,
        2
      ),
      "CLAUDE.md": [
        "You are a Breezing team lead. Complete the task described in PROMPT.md using Agent Teams.",
        "",
        "## Instructions",
        "",
        "1. TeamCreate to create a team",
        "2. TaskCreate to create implementation task(s)",
        "3. Spawn an Implementer (subagent_type=general-purpose) to write code",
        "   - Implementer must write clean TypeScript, handle edge cases, and create tests",
        "   - Implementer must run `npm test` and `npx tsc --noEmit` before reporting completion",
        "4. Spawn a Reviewer (subagent_type=general-purpose) to review the implementation",
        "   - Reviewer checks: correctness, type safety, edge cases, test quality",
        "   - Reviewer reports CRITICAL / ACCEPTABLE verdict",
        "5. If CRITICAL findings, send fix requests to Implementer (max 2 retake cycles)",
        "6. After approval, cleanup the team",
        "",
        "## Rules",
        "",
        "- Do NOT modify existing test files",
        "- Handle all edge cases in the implementation",
        "- All tests must pass before completion",
      ].join("\n"),
    });
  },
} satisfies ExperimentConfig;
