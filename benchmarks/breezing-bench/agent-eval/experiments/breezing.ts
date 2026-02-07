import type { ExperimentConfig } from "@vercel/agent-eval";

// Breezing condition: structured verification with validate script
export default {
  agent: "claude-code",
  model: "claude-haiku-4-5-20251001",
  runs: 3,
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
        "Complete the task described in PROMPT.md.",
        "Read the existing source files in src/ carefully.",
        "Run `npm run validate` to verify your implementation.",
        "Fix any issues found by the validation before finishing.",
      ].join("\n"),
    });
  },
} satisfies ExperimentConfig;
