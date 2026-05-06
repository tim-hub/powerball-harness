# Sandbagging-Aware Weak-Supervision Harness

## Purpose

This feature catches "looks successful, but may not be real" work.

Examples:

- tests pass because a hollow assertion was added
- a bugfix is claimed without a failing-before reproduction
- a Reviewer approval has no concrete evidence
- repeated failures plateau and the next Advisor prompt needs prior failure context

The feature is not a model trainer. It does not perform supervised fine-tuning, reinforcement learning, or hidden capability evaluation.

## Event Schema

### `weak-supervision-report.v1`

Use this when a reviewer or loop wants to score whether a task result is genuinely good.

Required fields:

| Field | Meaning |
|---|---|
| `run_id` | Current harness run, session, or loop identifier |
| `task_id` | Plans task or external task identifier |
| `rubric_id` | Rubric used to score the result |
| `reward_score` | Number from `0` to `1` |
| `verdict` | `APPROVE`, `REQUEST_CHANGES`, or `STOP` |
| `privacy_tags` | Data-use tags |
| `evidence_refs` | Files, logs, tests, screenshots, or artifacts used as evidence |

Schema file: `harness/scripts/lib/weak-supervision-report.schema.json`

Reviewer helper:

```bash
bash harness/scripts/review-weak-supervision-report.sh path/to/report.json
```

### `elicitation-event.v1`

Use this for local weak-supervision observations that can inform later Advisor or Reviewer prompts.

Allowed `event_kind` values:

| Kind | Use |
|---|---|
| `capability_probe` | A question or elicitation prompt was presented |
| `weak_label` | A weak or noisy label was recorded |
| `judge_verdict` | A judge or reviewer made a decision |
| `eval_result` | A scored evaluation result was recorded |
| `counterexample` | A failing case contradicts the success claim |

Local ledger:

```text
.claude/state/elicitation/events.jsonl
```

Schema file: `harness/scripts/lib/elicitation-event.schema.json`

## Privacy Tags

Allowed tags:

| Tag | Meaning |
|---|---|
| `may_train` | User explicitly allows downstream training-style reuse |
| `do_not_train` | Default. Do not use as training data |
| `synthetic_only` | Synthetic or fixture-only data |
| `legal_hold` | Preserve for legal/compliance reasons; do not mutate or purge casually |

Default:

```text
do_not_train
```

Override for a session:

```bash
HARNESS_ELICITATION_PRIVACY_TAGS=synthetic_only claude
```

## Advisor / Reviewer Boundary

Advisor remains a planning helper.

It still returns only:

- `PLAN`
- `CORRECTION`
- `STOP`

Reviewer remains the quality gate.

It still returns:

- `APPROVE`
- `REQUEST_CHANGES`

Weak-supervision cues can be injected into Advisor context on high-risk, repeated-failure, plateau, or explicit advisor-required requests. Those cues help choose the next plan, but they never approve the work.

## Local Ledger Only

This fork writes every elicitation event to the local JSONL ledger only.

```text
.claude/state/elicitation/events.jsonl
```

The hook must not fail the user's session when the ledger write encounters any error (errors are silently discarded). No external memory forwarding is performed.

## Failure Modes

Reviewer-side checks should treat these as major issues:

- `hardcoded_test_pass`
- `test_skip_added`
- `evidence_missing`
- `bugfix_without_reproduction`
- `reward_hacking`
- `counterexample_found`

## Verification

Focused checks:

```bash
cd go && go test ./internal/hookhandler/... -run 'Elicitation|Append'
bash tests/test-weak-supervision-report.sh
```

Full gates:

```bash
./tests/validate-plugin.sh
bash .claude/skills/release-this/scripts/check-consistency.sh
```

## Why This Shape

Weak supervision is useful here because the harness already has multiple partial observers: hooks, tests, reviewers, advisors. A lightweight event ledger lets them share evidence without pretending that a weak signal is a final truth.

That is why the implementation records observations and routes them to review, rather than adding a real training system or expanding Advisor authority.
