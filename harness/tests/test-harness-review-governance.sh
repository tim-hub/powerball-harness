#!/bin/bash
# Verify harness-review SKILL.md retains the TeamAgent Debate and acceptance-gate
# contract introduced in Phase 100 (task 100.6).

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

skill_file="$ROOT_DIR/skills/harness-review/SKILL.md"
refs_dir="$ROOT_DIR/skills/harness-review/references"

failures=0

err() { echo "$*" >&2; failures=$((failures + 1)); }

# --- SKILL.md existence and core markers ---

[ -f "$skill_file" ] || { echo "missing skill file: ${skill_file#$ROOT_DIR/}" >&2; exit 1; }

required_skill_terms=(
  "AskUserQuestion"
  "## Operating Modes"
  "--quick"
  "--codex-closeout"
  "--team-debate"
  "team-debate.md"
  "governance.md"
  "code-review.md"
  "codex-closeout.md"
  "REVIEW_AUTOSTART"
)

for term in "${required_skill_terms[@]}"; do
  grep -Fq -- "$term" "$skill_file" || err "missing required term in SKILL.md: $term"
done

grep -Eq '^allowed-tools: .*AskUserQuestion' "$skill_file" \
  || err "AskUserQuestion not in allowed-tools of SKILL.md"

line_count="$(wc -l < "$skill_file" | tr -d ' ')"
[ "$line_count" -le 350 ] || err "SKILL.md too large: ${line_count} lines (max 350)"

# --- Reference files existence check ---

reference_names=(
  "governance.md"
  "code-review.md"
  "codex-closeout.md"
  "team-debate.md"
  "plan-review.md"
  "scope-review.md"
  "dual-review.md"
)

for ref in "${reference_names[@]}"; do
  [ -f "$refs_dir/$ref" ] || err "missing reference file: references/$ref"
done

# --- Link-resolution: reference files named in Operating Modes table must exist ---
# Check that each .md filename mentioned in the Operating Modes section resolves to an
# actual file under references/ (catches broken ${CLAUDE_SKILL_DIR} link targets).
operating_modes_refs=("team-debate.md" "governance.md" "code-review.md" "codex-closeout.md")
for linked_ref in "${operating_modes_refs[@]}"; do
  grep -Fq "$linked_ref" "$skill_file" \
    || err "SKILL.md Operating Modes does not reference: $linked_ref"
  [ -f "$refs_dir/$linked_ref" ] \
    || err "Operating Modes links to nonexistent file: references/$linked_ref"
done

# --- Key terms across reference files ---

required_reference_terms=(
  "TeamAgent Debate"
  "spec source-of-truth"
  "Plans.md"
  "accepted findings"
  "rejected findings"
  "Stop-on-clean"
  "decision_needed"
  "team_agent_mode"
  "manual-pass"
  "Review default read-only boundary"
  "Do not push just to review"
  "Target selection"
  "Spec Agent"
  "Plans Agent"
  "Regression Agent"
  "Skeptic Agent"
)

for term in "${required_reference_terms[@]}"; do
  found=0
  for ref in "${reference_names[@]}"; do
    [ -f "$refs_dir/$ref" ] || continue
    grep -Fq -- "$term" "$refs_dir/$ref" && found=1 && break
  done
  [ "$found" -eq 1 ] || err "missing required reference term across all references: $term"
done

# --- No Japanese in any of the 4 new reference files ---

for ref in team-debate.md governance.md code-review.md codex-closeout.md; do
  f="$refs_dir/$ref"
  [ -f "$f" ] || continue
  if python3 -c "import sys, re; data=open(sys.argv[1]).read(); sys.exit(0 if re.search(r'[぀-ヿ一-鿿]', data) else 1)" "$f" 2>/dev/null; then
    err "Japanese characters found in references/$ref"
  fi
done

if [ "$failures" -gt 0 ]; then
  exit 1
fi

echo "test-harness-review-governance: ok"
