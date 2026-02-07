import type { ExperimentConfig } from "@vercel/agent-eval";

// Quick re-calibration for task-14 after BUG comment removal
export default {
  agent: "vercel-ai-gateway/claude-code",
  model: "haiku",
  runs: 3,
  earlyExit: false,
  timeout: 300,
  scripts: ["test"],
  sandbox: "docker",
  evals: ["task-14"],
  setup: async (sandbox) => {
    await sandbox.writeFiles({
      "CLAUDE.md": [
        "Complete the task described in PROMPT.md.",
        "Read the existing source files in src/ carefully.",
      ].join("\n"),
    });
  },
} satisfies ExperimentConfig;
