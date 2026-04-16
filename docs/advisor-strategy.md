# Advisor Strategy

The Advisor is a read-only consultation agent that gives Workers structured guidance — one of three decision types: PLAN, CORRECTION, or STOP — at specific trigger points during task execution. It operates entirely in an advisory capacity: it does not execute code, does not modify files, and cannot replace the Reviewer's final APPROVE/REQUEST_CHANGES verdict. Workers consult the Advisor when they are blocked, before high-risk operations, or after repeated failures, then decide how to act on the guidance.

## Trigger Conditions

| Trigger Name | Condition Description | When It Fires | `reason_code` Value |
|---|---|---|---|
| `high_risk_preflight` | Worker is about to touch a file or operation flagged as security-sensitive or needs-spike in the sprint contract | Before the risky operation begins | `high_risk_preflight` |
| `repeated_failure` | The same error signature has caused a build/test failure at least `retry_threshold` times | After the Nth retry when threshold is met | `repeated_failure` |
| `plateau_before_escalation` | Worker has exhausted local fix attempts and is one step away from escalating to the user | Before the escalation prompt is shown | `plateau_before_escalation` |
| `explicit_marker` | A task description or sprint contract contains an `[advisor]` annotation requesting consultation | At the start of the annotated task | `explicit_marker` |

## Decision Types

| Decision | Meaning | What the Executor Does on Receipt |
|---|---|---|
| `PLAN` | The current approach has a structural problem; a different overall strategy is recommended | Worker discards the current approach and replans from the beginning using the Advisor's suggested strategy |
| `CORRECTION` | A targeted, local fix is available for the current error without changing the overall approach | Worker applies the specific correction described, then resumes normal execution |
| `STOP` | The problem exceeds what the Worker can resolve alone; escalation is required | Worker stops the fix loop and escalates to the Reviewer (not the user directly); the Reviewer decides whether to APPROVE, REQUEST_CHANGES, or surface to the user |

## Advisor vs Reviewer Authority

The Advisor and Reviewer serve distinct roles and must not be conflated.

The **Advisor** provides mid-task guidance only. It has no final authority over whether a task is complete. Its decisions are guidance, not verdicts — the Worker may reject a CORRECTION if it contradicts task requirements, and a PLAN recommendation must still be validated against the sprint contract's DoD.

The **Reviewer** holds the final APPROVE/REQUEST_CHANGES gate. Its verdict is definitive. No Advisor decision can override it.

When the Advisor returns `STOP`, this means "escalate to the Reviewer," not "auto-reject the task." The Reviewer then evaluates the work independently and issues its own verdict.

## Duplicate Suppression

Before invoking the Advisor, the Worker computes a hash of the tuple `(task_id + reason_code + error_sig)` and checks it against `.claude/state/advisor/history.jsonl`. If an identical consultation has already occurred for this task, the request is skipped silently.

This prevents the same blocked state from triggering repeated identical Advisor calls. The history file is append-only; entries are written by `run-advisor-consultation.sh` immediately after a consultation request is accepted. The `max_consults_per_task` config field provides a hard upper bound as a secondary guard.

## Configuration Reference

All Advisor settings live under the `advisor:` block in `harness/.claude-code-harness.config.yaml`.

| Field | Type | Description |
|---|---|---|
| `enabled` | boolean | Master switch. When `false`, all consultation requests are silently skipped and `run-advisor-consultation.sh` exits 0 with the message "advisor disabled". |
| `mode` | string | Consultation mode. `on-demand` means the Advisor is only invoked when a trigger condition fires. |
| `max_consults_per_task` | integer (≥ 1) | Maximum number of Advisor consultations allowed per task ID. Once reached, further requests for that task are skipped. |
| `retry_threshold` | integer (> 0) | Number of retries after which the `repeated_failure` trigger fires. Must be a positive integer. |
| `consult_before_user_escalation` | boolean | When `true`, the `plateau_before_escalation` trigger fires automatically before surfacing a failure to the user. |
| `model_defaults.claude` | string | Model tier for the Advisor subagent. Should be `opus` for best guidance quality. |

## 4-Agent Team Model

```
Lead (harness-work --breezing)
  |
  +-- Worker (powerball-harness:worker)  ──consults──>  Advisor (powerball-harness:advisor)
  |     | Implementation                                  Read-only guidance
  |     | Preflight self-check                            Returns PLAN / CORRECTION / STOP
  |     + Worktree commit
  |
  +-- Reviewer (powerball-harness:reviewer)
        Final APPROVE / REQUEST_CHANGES verdict
```

The Advisor sits laterally to the Worker — it is not spawned by Lead and is not in the Lead → Worker chain of command. Lead is unaware of individual Advisor consultations; it only sees the Worker's eventual result or escalation. This matches the diagram in [team-composition.md](../harness/agents/team-composition.md).

## harness-loop Integration

`harness-loop` invokes the Advisor at all three runtime trigger points — `high_risk_preflight`, `repeated_failure`, and `plateau_before_escalation` — by calling `run-advisor-consultation.sh` with the appropriate `--reason-code` and `--error-sig` arguments. When the Advisor returns a `STOP` decision, `harness-loop` exits the fix loop cleanly, writes a structured summary of the failure context and the Advisor's rationale to `.claude/state/advisor/last-request.json`, and escalates to the Reviewer rather than prompting the user directly.
