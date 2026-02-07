import type { ExperimentConfig } from "@vercel/agent-eval";

// GLM calibration: vanilla condition (no validate)
// Uses Z.AI Anthropic-compatible endpoint via gateway mode
export default {
  agent: "vercel-ai-gateway/claude-code",
  model: "haiku",
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
      ].join("\n"),
    });
  },
} satisfies ExperimentConfig;
