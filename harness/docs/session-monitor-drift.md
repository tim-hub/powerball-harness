# Session Monitor — Active Drift Detection

The Harness session monitor (`bin/harness hook session-monitor`) runs at every SessionStart and emits `⚠️` warnings when your session has drifted from a healthy state.

## Plans.md Drift

**Warning**: `⚠️ plans drift: WIP={n}, stale_for={h}h`

Emitted when:
- `WIP >= wip_threshold` (default: 5) — too many tasks in flight simultaneously
- `stale_for >= stale_hours` (default: 24h) — Plans.md hasn't been modified in over a day

**Tune** in `harness/.claude-code-harness.config.yaml`:
```yaml
monitor:
  plans_drift:
    wip_threshold: 5    # lower for stricter WIP limits
    stale_hours: 24     # hours without Plans.md modification before warning
```

## Advisor/Reviewer Drift

**Warning**: `⚠️ advisor drift: request_id={id}, waiting {elapsed}s`

Emitted when an `advisor-request.v1` event in `.claude/state/session.events.jsonl` has no matching `advisor-response.v1` within the TTL.

**Tune** in `harness/.claude-code-harness.config.yaml`:
```yaml
orchestration:
  advisor_ttl_seconds: 600    # seconds before an unanswered advisor request triggers warning
```

## harness-mem Health

**Warning**: `⚠️ harness-mem unhealthy: {reason}`

Run manually with `bin/harness mem health`. Reasons:
- `not-initialized`: `~/.claude-mem/` directory does not exist
- `corrupted-settings`: `~/.claude-mem/settings.json` is not valid JSON
- `daemon-unreachable`: harness-mem daemon not listening on `HARNESS_MEM_HOST:HARNESS_MEM_PORT` (default `127.0.0.1:37888`)
