---
name: deploy
description: "Deploys to Vercel or Netlify, runs health checks, and monitors post-deploy. Use when deploying to production or checking deployment status."
when_to_use: "deploy to Vercel, deploy to Netlify, production push, health check, post-deploy monitoring, analytics setup"
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
