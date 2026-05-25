# Session Init — Step Implementation Details

Full bash commands and templates for each session-init step.

## Step 0: File Status Check

```bash
# Check Plans.md line count
if [ -f "Plans.md" ]; then
  lines=$(wc -l < Plans.md)
  if [ "$lines" -gt 200 ]; then
    echo "⚠️ Plans.md has ${lines} lines. Recommend cleanup with 'clean up'"
  fi
fi

# Check session-log.md line count
if [ -f ".claude/memory/session-log.md" ]; then
  lines=$(wc -l < .claude/memory/session-log.md)
  if [ "$lines" -gt 500 ]; then
    echo "⚠️ session-log.md has ${lines} lines. Recommend cleanup with 'clean up session log'"
  fi
fi
```

If cleanup is needed, a suggestion is displayed (does not affect work).

## Step 0.7: Unified Harness Memory Resume Pack

Required call:

```text
harness_mem_resume_pack(project, session_id?, limit=5, include_private=false)
```

Operational rules:
- `project` must always specify the current project name
- `session_id` is obtained from `$CLAUDE_SESSION_ID`, falling back to `.session_id` in `.claude/state/session.json`
- Using the first result of `harness_mem_sessions_list(project, limit=1)` is limited to read-only (resume confirmation); do not use it for writes via `record_checkpoint` / `finalize_session`
- Inject retrieved results into the session start context
- On retrieval failure, check daemon status with `harness_mem_health()`, report the failure explicitly, and continue
- Recovery order: `scripts/harness-memd doctor` -> `scripts/harness-memd cleanup-stale` -> `scripts/harness-memd start`

## Step 1: Environment Check Commands

Execute the following in parallel:

```bash
# Git status
git status -sb
git log --oneline -3
```

```bash
# Plans.md
cat Plans.md 2>/dev/null || echo "Plans.md not found"
```

```bash
# Key points from AGENTS.md
head -50 AGENTS.md 2>/dev/null || echo "AGENTS.md not found"
```

## Step 3: Output Status Report Template

```markdown
## 🚀 Session Start

**Date/Time**: {{YYYY-MM-DD HH:MM}}
**Branch**: {{branch}}
**Session ID**: ${CLAUDE_SESSION_ID}

---

### 📋 Today's Tasks

**Priority Tasks**:
- {{pm:requesting or cc:WIP tasks}}

**Other Tasks**:
- {{List of cc:TODO tasks}}

---

### ⚠️ Notes

{{Important constraints and prohibitions from AGENTS.md}}

---

**Ready to start work?**
```

## Output Format

At session start, present the following information concisely:

| Item | Content |
|------|---------|
| Current branch | e.g., `staging` |
| Priority tasks | Top 1-2 most important |
| Notes | Summary of prohibitions |
| Next action | Specific suggestions |
