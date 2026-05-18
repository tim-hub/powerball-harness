# TeamAgent Debate

## Summary

TeamAgent Debate is a read-only review pass where multiple agents read the same change from different perspectives to reduce blind spots.

## When required

Run when any of the following applies:

- Changes span multiple modules
- Change touches security / auth / release / distribution / mirror
- Alignment with the spec source-of-truth or `Plans.md` is ambiguous
- Regression risk is high
- Claude and Codex verdicts diverge
- Reviewers disagree on perspective-based ratings
- Same issue fails two consecutive post-fix re-reviews

## Agents

| Agent | Primary question |
|---|---|
| Spec Agent | Find contradictions between the spec source-of-truth and the implementation diff |
| Plans Agent | Verify the diff aligns with `Plans.md` task / DoD / Depends |
| Regression Agent | Find regressions in existing behavior, tests, distribution mirrors, CLI/skill UX |
| Skeptic Agent | Find major risks being missed because of an implicit assumption that the change should pass |

Minimum 2 perspectives; up to 4 when needed.
All agents are read-only.

## Codex fallback

Do not skip TeamAgent Debate when native TeamAgents are unavailable in a Codex environment.

Available fallbacks:

- `codex-companion.sh review`
- reviewer subagent
- explicit manual-pass (each perspective reviewed separately)

Record one of the following in `team_agent_mode`:

- `native`
- `codex-companion`
- `manual-pass`
- `unavailable`

If `unavailable` and manual-pass is also impossible, stop with `decision_needed`.

## Output

```json
{
  "team_debate": {
    "required": true,
    "mode": "manual-pass",
    "team_agent_mode": "manual-pass",
    "agents": ["Spec Agent", "Plans Agent", "Regression Agent"],
    "disagreements": [],
    "acceptance_bar": {
      "spec_alignment": "pass",
      "plans_alignment": "pass",
      "regression_safety": "pass"
    }
  }
}
```

## Pass criteria

If a TeamAgent Debate disagreement is critical / major → `REQUEST_CHANGES`.
If downgrading to minor / recommendation, document the reason with evidence.
