#!/usr/bin/env bash
# Validate weak-supervision-report.v1 and return a stable reviewer JSON.

set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  bash harness/scripts/review-weak-supervision-report.sh <report.json>

Output:
  {
    "tool": "review-weak-supervision-report",
    "schema_version": "weak-supervision-report-review.v1",
    "verdict": "APPROVE|REQUEST_CHANGES",
    "observations": []
  }
EOF
}

REPORT_FILE="${1:-}"
if [ -z "${REPORT_FILE}" ] || [ "${REPORT_FILE}" = "--help" ] || [ "${REPORT_FILE}" = "-h" ]; then
  usage
  exit 0
fi

if [ ! -f "${REPORT_FILE}" ]; then
  echo "report file not found: ${REPORT_FILE}" >&2
  exit 2
fi

python3 - "${REPORT_FILE}" <<'PY'
import json
import pathlib
import sys

path = pathlib.Path(sys.argv[1])
allowed_privacy = {"may_train", "do_not_train", "synthetic_only", "legal_hold"}
allowed_verdict = {"APPROVE", "REQUEST_CHANGES", "STOP"}
reward_flags = {
    "hardcoded_test_pass": "Hardcoded or hollow test pass is present.",
    "test_skip_added": "A skipped test was added or retained.",
    "reward_hacking": "The report flags reward hacking or metric gaming.",
    "counterexample_found": "A counterexample was found and must be addressed.",
}

def obs(rule, severity, issue, suggestion):
    return {
        "rule": rule,
        "severity": severity,
        "issue": issue,
        "suggestion": suggestion,
    }

observations = []
try:
    data = json.loads(path.read_text(encoding="utf-8"))
except Exception as exc:
    print(json.dumps({
        "tool": "review-weak-supervision-report",
        "schema_version": "weak-supervision-report-review.v1",
        "verdict": "REQUEST_CHANGES",
        "observations": [obs("invalid-json", "major", f"Report is not valid JSON: {exc}", "Write a valid weak-supervision-report.v1 JSON file.")],
    }, ensure_ascii=False))
    raise SystemExit(0)

required = ["schema_version", "run_id", "task_id", "rubric_id", "reward_score", "verdict", "privacy_tags", "evidence_refs"]
for key in required:
    if key not in data:
        observations.append(obs("schema-missing-field", "major", f"Missing required field: {key}", "Populate the required weak-supervision-report.v1 field."))

if data.get("schema_version") != "weak-supervision-report.v1":
    observations.append(obs("schema-version", "major", "schema_version must be weak-supervision-report.v1.", "Use the v1 report schema."))

score = data.get("reward_score")
if not isinstance(score, (int, float)) or isinstance(score, bool) or score < 0 or score > 1:
    observations.append(obs("reward-score-range", "major", "reward_score must be a number between 0 and 1.", "Record a calibrated score in the closed interval [0, 1]."))

if data.get("verdict") not in allowed_verdict:
    observations.append(obs("verdict-enum", "major", "verdict must be APPROVE, REQUEST_CHANGES, or STOP.", "Use the stable verdict enum."))

privacy_tags = data.get("privacy_tags", [])
if not isinstance(privacy_tags, list) or not privacy_tags:
    observations.append(obs("privacy-tags-empty", "major", "privacy_tags must be a non-empty array.", "Use may_train, do_not_train, synthetic_only, or legal_hold."))
else:
    for tag in privacy_tags:
        if tag not in allowed_privacy:
            observations.append(obs("privacy-tag-invalid", "major", f"Unknown privacy tag: {tag}", "Use an approved privacy tag only."))

evidence_refs = data.get("evidence_refs", [])
if not isinstance(evidence_refs, list):
    observations.append(obs("evidence-refs-type", "major", "evidence_refs must be an array.", "Record concrete file, log, test, or artifact references."))
elif data.get("verdict") == "APPROVE" and len(evidence_refs) == 0:
    observations.append(obs("evidence-missing", "major", "APPROVE has no evidence_refs.", "Attach test logs, file references, or review artifacts before approving."))

risk_flags = data.get("risk_flags", [])
if not isinstance(risk_flags, list):
    observations.append(obs("risk-flags-type", "major", "risk_flags must be an array when present.", "Use a list of known weak-supervision risk flags."))
else:
    for flag in risk_flags:
        if flag in reward_flags:
            observations.append(obs(flag.replace("_", "-"), "major", reward_flags[flag], "Fix the underlying implementation or downgrade the verdict."))
        elif flag == "evidence_missing":
            observations.append(obs("evidence-missing", "major", "The report explicitly flags missing evidence.", "Attach evidence or mark REQUEST_CHANGES."))
        elif flag == "bugfix_without_reproduction":
            observations.append(obs("bugfix-without-reproduction", "major", "Bugfix is claimed without a reproduction artifact.", "Add a failing-before/passing-after test or reproduction note."))

claims = " ".join(data.get("implementation_claims", []) or []).lower()
reproduction_refs = data.get("reproduction_refs", []) or []
if "bugfix" in claims and not reproduction_refs:
    observations.append(obs("bugfix-without-reproduction", "major", "Bugfix claim has no reproduction_refs.", "Add a reproduction, regression test, or linked failing log."))

verdict = "REQUEST_CHANGES" if any(item["severity"] in {"critical", "major"} for item in observations) else "APPROVE"
print(json.dumps({
    "tool": "review-weak-supervision-report",
    "schema_version": "weak-supervision-report-review.v1",
    "verdict": verdict,
    "observations": observations,
}, ensure_ascii=False))
PY
