import type { ExperimentConfig } from "@vercel/agent-eval";

// Calibration: breezing condition (with validate)
export default {
  agent: "claude-code",
  model: "claude-haiku-4-5-20251001",
  runs: 3,
  earlyExit: false,
  timeout: 300,
  scripts: ["test"],
  sandbox: "docker",
  evals: [
    "task-02",
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
