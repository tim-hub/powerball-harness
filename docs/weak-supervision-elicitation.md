# Weak-Supervision Elicitation Ledger — Reader's Guide

## What is the ledger?

The elicitation ledger is an append-only JSONL file at:

```text
.claude/state/elicitation/events.jsonl
```

Every `Elicitation` and `ElicitationResult` hook call writes one JSON line to this file. The ledger accumulates observations from the current and past sessions, and the weak-supervision pipeline reads it for review context.

## Event kinds

Each line is an `elicitation-event.v1` JSON object with an `event_kind` field:

| Kind | When it is written |
|---|---|
| `capability_probe` | An `Elicitation` hook fires — MCP server asked the user a question |
| `eval_result` | An `ElicitationResult` hook fires — user answered (or the result was skipped) |
| `weak_label` | Written externally to record a noisy reviewer label |
| `judge_verdict` | Written externally to record a judge or reviewer verdict |
| `counterexample` | Written externally to flag a failing case that contradicts a success claim |

The `capability_probe` and `eval_result` kinds are written automatically by the Go guardrail engine. The others can be appended manually or by scripts.

## Schema

See `harness/scripts/lib/elicitation-event.schema.json` for the full JSON Schema.

Required fields in every event:

| Field | Example |
|---|---|
| `schema_version` | `"elicitation-event.v1"` |
| `event_kind` | `"capability_probe"` |
| `run_id` | `"session-abc123"` |
| `privacy_tags` | `["do_not_train"]` |
| `evidence_refs` | `[]` |
| `source` | `"claude-code-hook:elicitation"` |
| `timestamp` | `"2026-05-06T12:00:00Z"` |

## Privacy tags

The default privacy tag is `do_not_train`. Override for a session:

```bash
HARNESS_ELICITATION_PRIVACY_TAGS=synthetic_only claude
```

Allowed values: `may_train`, `do_not_train`, `synthetic_only`, `legal_hold`.

Events with unrecognized tags are rejected and not written to the ledger.

## Using the scripts

### Validate a weak-supervision report

After a reviewer or loop produces a `weak-supervision-report.v1` payload:

```bash
bash harness/scripts/review-weak-supervision-report.sh report.json
```

Outputs structured JSON with `verdict: APPROVE` or `verdict: REQUEST_CHANGES` and an `observations` array.

Schema reference: `harness/scripts/lib/weak-supervision-report.schema.json`

## Running the tests

```bash
bash tests/test-weak-supervision-report.sh
```

This script validates both schema files and runs the reviewer against fixture payloads (one valid, several invalid).
