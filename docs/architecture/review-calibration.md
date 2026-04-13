# Review Calibration

Storage format and operational rules for reducing review drift.

## Storage Locations

- `.claude/state/review-result.json`
- `.claude/state/review-calibration.jsonl`
- `.claude/state/review-few-shot-bank.json`

## Recording Rules

When `review-result.json` includes a `calibration` entry, `record-review-calibration.sh`
appends one line to `review-calibration.jsonl`.

`calibration.label` is limited to one of the following:

- `false_positive`
- `false_negative`
- `missed_bug`
- `overstrict_rule`

## Few-shot Updates

`build-review-few-shot-bank.sh` extracts the latest samples from the calibration log
and regenerates the few-shot JSON bank.

## Quality Stance

- Only use `REQUEST_CHANGES` for critical bugs
- Keep evidence-free hunches at `minor` or `recommendation` level
- Write findings short and specific enough to be reused as few-shot examples
