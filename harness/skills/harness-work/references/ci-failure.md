# CI Failure Handling

Auto-fix loop with a 3-failure cap — classify, fix, and escalate with a structured summary when the same root cause fails three times.

---

When CI fails:

1. Check logs and identify the error
2. Implement fixes
3. Stop the auto-fix loop after 3 failures from the same cause
4. Summarize failure logs, attempted fixes, and remaining issues for escalation
