#!/usr/bin/env bash
# propose-skill-variants.sh — Generate 3 SKILL.md variants using a Claude subagent
#
# Usage:
#   bash local-scripts/propose-skill-variants.sh <skill-dir>
#   bash local-scripts/eval-skill.sh <skill-dir> <suite> | bash local-scripts/propose-skill-variants.sh <skill-dir>
#
# Arguments:
#   <skill-dir>   Path to a skill directory containing SKILL.md
#
# Stdin (optional):
#   JSON from eval-skill.sh — used to inform what aspects to improve
#
# Output:
#   3 SKILL.md variants saved to /tmp/skill-variants/<skill-basename>-v{1,2,3}/SKILL.md
#   Summary of what each variant changed printed to stdout
#
# Exit codes:
#   0 — all 3 variants generated successfully
#   1 — SKILL.md not found or other error
#   2 — invocation error

set -euo pipefail

# ---- arg parsing ----
if [[ $# -lt 1 ]]; then
  echo "Usage: $0 <skill-dir>" >&2
  exit 2
fi

SKILL_DIR="$1"
SKILL_BASENAME=$(basename "$SKILL_DIR")
SKILL_FILE="${SKILL_DIR}/SKILL.md"

if [[ ! -f "${SKILL_FILE}" ]]; then
  echo "Error: ${SKILL_FILE} not found" >&2
  exit 1
fi

# ---- extract YAML frontmatter (lines between first and second ---) ----
# FRONTMATTER captures the lines between the two --- delimiters (not including the --- lines)
FRONTMATTER=$(awk 'BEGIN{f=0} /^---$/{f++; if(f==2){exit}} {if(f>=1)print}' "${SKILL_FILE}")

# ---- extract skill body (strip YAML frontmatter) ----
# Skip everything up to and including the second --- delimiter
SKILL_BODY=$(awk 'BEGIN{fence=0} /^---$/{fence++; next} fence>=2{print}' "${SKILL_FILE}")

if [[ -z "${SKILL_BODY}" ]]; then
  # No frontmatter found — use the whole file
  SKILL_BODY=$(cat "${SKILL_FILE}")
  FRONTMATTER="---"
fi

# ---- read eval output from stdin if available (not a terminal) ----
EVAL_JSON=""
if [ ! -t 0 ]; then
  EVAL_JSON=$(cat)
fi

# ---- build eval context section for the prompt ----
EVAL_CONTEXT=""
if [[ -n "${EVAL_JSON}" ]]; then
  EVAL_CONTEXT="
## Evaluation Results (use these to inform your changes)

The current skill was evaluated against a test suite. Here are the results:

\`\`\`json
${EVAL_JSON}
\`\`\`

Use the eval results to understand where the skill underperforms. Focus your variant's change on improving the skill's behavior in failing cases.
"
fi

echo "==> Generating 3 SKILL.md variants for: ${SKILL_BASENAME}"
echo ""

# ---- generate 3 variants ----
for N in 1 2 3; do
  OUT_DIR="/tmp/skill-variants/${SKILL_BASENAME}-v${N}"
  mkdir -p "${OUT_DIR}"

  echo "--> Generating variant ${N}/3..."

  # Build a description of what kind of variation to make for this N
  case "${N}" in
    1) VARIANT_HINT="Focus on strictness level: adjust the threshold for when to approve vs request changes. You may make the skill stricter (lower tolerance for minor issues) or more lenient (only block on serious issues). Rewrite the verdict criteria section to clearly reflect this change in strictness." ;;
    2) VARIANT_HINT="Focus on output format and verdict instruction phrasing: rewrite how the skill instructs the reviewer to structure their response. Change the ordering or phrasing of verdict instructions, rationale requirements, or output format requirements. The review logic should stay similar but the output instructions should be meaningfully different." ;;
    3) VARIANT_HINT="Focus on review priority and emphasis: reorder the priority of review concerns (e.g., promote security above correctness, or emphasize test coverage more than style). The change should meaningfully alter what gets flagged first and why." ;;
  esac

  PROMPT="You are helping improve a Claude Code skill (SKILL.md) by generating a variant.

## Current SKILL.md body (no frontmatter)

\`\`\`
${SKILL_BODY}
\`\`\`
${EVAL_CONTEXT}
## Your Task

Generate variant ${N} of 3 for this skill.

**Variation axis for this variant**: ${VARIANT_HINT}

## Rules

1. Output ONLY the complete skill body — no YAML frontmatter, no explanation, no markdown fences wrapping the entire output.
2. Make exactly ONE meaningful change aligned with the variation axis above.
3. The output must be a complete, valid skill body (not truncated).
4. Keep all other sections intact; only change what's needed for the variation axis.
5. After the skill body, append a single HTML comment on its own line summarizing what changed:
   <!-- variant-${N}-change: <one sentence description of the change made> -->

Output the complete skill body now:"

  # Run claude subagent to generate the variant body
  variant_body=$(printf '%s' "${PROMPT}" | \
    claude --print \
      --model haiku \
      --output-format text \
      --exclude-dynamic-system-prompt-sections \
      2>/dev/null || true)

  if [[ -z "${variant_body}" ]]; then
    echo "Warning: claude returned empty output for variant ${N}" >&2
    variant_body="# ${SKILL_BASENAME} (variant ${N})

Review diffs. Output APPROVE or REQUEST_CHANGES.

<!-- variant-${N}-change: fallback variant — claude returned empty output -->"
  fi

  # Extract the change summary from the HTML comment (last occurrence)
  change_summary=$(printf '%s' "${variant_body}" | \
    grep -oE '<!-- variant-[0-9]+-change: .* -->' | tail -1 || true)
  if [[ -z "${change_summary}" ]]; then
    change_summary="(no change summary found in output)"
  fi

  # Reassemble: frontmatter + --- + body
  # FRONTMATTER already starts with the opening ---, so we need to close it
  printf '%s\n---\n%s\n' "${FRONTMATTER}" "${variant_body}" > "${OUT_DIR}/SKILL.md"

  echo "    Saved: ${OUT_DIR}/SKILL.md"
  echo "    Change: ${change_summary}"
  echo ""
done

# ---- print summary ----
echo "==> Done. 3 variants generated:"
for N in 1 2 3; do
  echo "    /tmp/skill-variants/${SKILL_BASENAME}-v${N}/SKILL.md"
done
