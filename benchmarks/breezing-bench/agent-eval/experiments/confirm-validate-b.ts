import type { ExperimentConfig } from "@vercel/agent-eval";

// Confirmatory validate batch B: tasks 16-20, 5 runs each
export default {
  agent: "vercel-ai-gateway/claude-code",
  model: "haiku",
  runs: 5,
  earlyExit: false,
  timeout: 300,
  scripts: ["test"],
  sandbox: "docker",
  evals: ["task-16", "task-17", "task-18", "task-19", "task-20"],
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
