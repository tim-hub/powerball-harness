---
name: port-upstream
description: "Analyzes upstream fork PR changes and ports selected items into the local branch. Use when checking upstream PRs, porting upstream changes, or planning selective upstream integration."
when_to_use: "port upstream, upstream PR, check upstream, upstream changes, upstream fork, bring changes from upstream"
user-invocable: true
argument-hint: "[pr_url]"
allowed-tools: ["Agent", "Bash", "Read", "Glob", "Grep"]
---

## Step 1. Spawn an Explore agent to analyze the upstream PR

- Understand the changes whether there are new features, bug fixes, comment updates, scripts updated, or go source code update
- Investigate and get what we can bring to local branch

## Step 2. Build the harness plan from the finding

- Based on what we found in Step 1, write a plan through `/harness-plan`
- If there is nothing worth porting, just tell user.


## Rules

- If changes are in Japanese or non English, translate to English
