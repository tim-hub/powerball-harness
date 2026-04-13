---
name: deploy
description: "Use this skill whenever the user mentions deploying, pushing to production, Vercel setup, Netlify config, deployment monitoring, health check endpoints, or production analytics. Also use when the user wants to verify a deployment is healthy or set up post-deploy monitoring. Do NOT load for: feature implementation, local development, code reviews, or project initialization. Configures and executes deployments to Vercel or Netlify, including analytics setup and health checks."
allowed-tools: ["Read", "Write", "Edit", "Bash"]
disable-model-invocation: true
argument-hint: "[vercel|netlify|health]"
context: fork
model: sonnet
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
