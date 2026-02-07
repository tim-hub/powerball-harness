import type { ExperimentConfig } from "@vercel/agent-eval";

// Confirmatory study: baseline condition (no validate instruction)
export default {
  agent: "vercel-ai-gateway/claude-code",
  model: "haiku",
  runs: 5,
  earlyExit: false,
  timeout: 300,
  scripts: ["test"],
  sandbox: "docker",
  evals: [
    "task-11",
    "task-12",
    "task-13",
    "task-14",
    "task-15",
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
