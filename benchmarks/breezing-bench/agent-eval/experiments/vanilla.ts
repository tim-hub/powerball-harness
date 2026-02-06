import type { ExperimentConfig } from "@vercel/agent-eval";

export default {
  agent: "claude-code",
  model: "claude-haiku-4-5-20251001",
  runs: 2,
  earlyExit: false,
  timeout: 300,
  scripts: ["test"],
  sandbox: "docker",
  evals: [
    "task-01",
    "task-02",
    "task-03",
    "task-04",
    "task-05",
    "task-06",
    "task-07",
    "task-08",
    "task-09",
    "task-10",
  ],
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
