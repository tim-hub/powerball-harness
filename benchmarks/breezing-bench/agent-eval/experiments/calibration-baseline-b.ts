import type { ExperimentConfig } from "@vercel/agent-eval";

// Calibration batch B: tasks 16-20, 2 runs each (10 concurrent)
export default {
  agent: "vercel-ai-gateway/claude-code",
  model: "haiku",
  runs: 2,
  earlyExit: false,
  timeout: 300,
  scripts: ["test"],
  sandbox: "docker",
  evals: [
    "task-16",
    "task-17",
    "task-18",
    "task-19",
    "task-20",
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
