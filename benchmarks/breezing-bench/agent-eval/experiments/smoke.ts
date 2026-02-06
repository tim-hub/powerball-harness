import type { ExperimentConfig } from "@vercel/agent-eval";

export default {
  agent: "claude-code",
  model: "claude-haiku-4-5-20251001",
  runs: 1,
  earlyExit: true,
  timeout: 300,
  scripts: ["test"],
  sandbox: "docker",
  evals: ["task-01"],
  setup: async (sandbox) => {
    await sandbox.writeFiles({
      "CLAUDE.md": [
        "You are a developer. Complete the task described in PROMPT.md.",
        "Write clean TypeScript with proper error handling.",
        "Create tests. Make sure all existing tests pass.",
        "Run `npm test` to verify your implementation.",
        "Run `npx tsc --noEmit` to verify type safety.",
      ].join("\n"),
    });
  },
} satisfies ExperimentConfig;
