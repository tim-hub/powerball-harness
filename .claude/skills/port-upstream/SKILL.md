---
name: port-upstream
description: "Check upstream PR changes and port them here. Use when check upstream PR, port upstream"
user-invocable: true
argument-hint: "[pr_url]"
---

## Step 1. Spawn an Explore agent to analyze the upstream PR

- Understand the changes whether there are new features, bug fixes, comment updates, scripts updated, or go source code update
- Investigate and get what we cran bring to local branch

## Step 2. Build the harness plan from the finding

- Base on what we found on Step 1 to write a plan through `/harness-plan`
- If there is nothing worth porting, just tell user.


## Rules

- If changes are in Japanese or non English, translate to English
