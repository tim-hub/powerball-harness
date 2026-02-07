import type { ExperimentConfig } from "@vercel/agent-eval";

// GLM calibration: breezing condition via Z.AI Anthropic-compatible endpoint
// Requires: AI_GATEWAY_API_KEY and AI_GATEWAY_BASE_URL in .env
export default {
  agent: "vercel-ai-gateway/claude-code",
  model: "haiku",
  runs: 1,
  earlyExit: false,
  timeout: 300,
  scripts: ["test"],
  sandbox: "docker",
  evals: [
    "task-02",
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
