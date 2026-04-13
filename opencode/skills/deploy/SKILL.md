---
name: deploy
description: "Use when deploying to Vercel or Netlify — production pushes, health checks, post-deploy monitoring, or analytics setup. Do NOT load for: feature implementation, local dev, reviews, or project init."
allowed-tools: ["Read", "Write", "Edit", "Bash"]
disable-model-invocation: true
argument-hint: "[vercel|netlify|health]"
context: fork
---

# Deploy Skills

A collection of skills responsible for deployment and monitoring configuration.

## Feature Details

| Feature | Details |
|---------|--------|
| **Deployment Setup** | See [references/deployment-setup.md](${CLAUDE_SKILL_DIR}/references/deployment-setup.md) |
| **Analytics** | See [references/analytics.md](${CLAUDE_SKILL_DIR}/references/analytics.md) |
| **Health Checking** | See [references/health-checking.md](${CLAUDE_SKILL_DIR}/references/health-checking.md) |

## Execution Steps

1. Classify the user's request
2. Read the appropriate reference file from "Feature Details" above
3. Configure according to its contents
