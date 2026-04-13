#!/usr/bin/env bash
# audit-skill-descriptions.sh — SKILL.md description format auditor
#
# Purpose:
#   Scan SKILL.md files for description-field violations defined in
#   .claude/rules/skill-description.md:
#     - Missing "Use when " prefix
#     - Contains forbidden phrases ("the user mentions", "the user asks",
#       "Use this skill")
#     - Length > 300 characters
#
#   Outputs one tab-separated line per violation on stdout so the CI
#   wrapper (Phase 44.6) can route each into its own fail_test line:
#     <file>\t<kind>\t<snippet>
#
#   Summary is printed to stderr and does not affect parseable stdout.
#
# Usage:
#   bash scripts/audit-skill-descriptions.sh [target-dir]
#
#   With no argument, scans skills/, opencode/skills/, and skills-codex/.
#   With one argument, scans only that directory.
#
# Exit codes:
#   0 — no violations
#   1 — one or more violations found
#   2 — invocation error (bad args, target directory missing)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

REQUIRED_PREFIX="Use when "
MAX_LEN=300
SNIPPET_MAX=120

FORBIDDEN_PHRASES=(
  "the user mentions"
  "the user asks"
  "Use this skill"
)

DEFAULT_TARGETS=("skills" "opencode/skills" "skills-codex")

# ---- arg parsing ----
if [ "$#" -gt 1 ]; then
  echo "Usage: $0 [target-dir]" >&2
  exit 2
fi

if [ "$#" -eq 1 ]; then
  TARGETS=("$1")
else
  TARGETS=("${DEFAULT_TARGETS[@]}")
fi

# ---- resolve and validate targets ----
RESOLVED_TARGETS=()
for t in "${TARGETS[@]}"; do
  if [[ "$t" = /* ]]; then
    abs="$t"
  else
    abs="${REPO_ROOT}/${t}"
  fi
  if [ ! -d "$abs" ]; then
    echo "Error: target directory not found: $abs" >&2
    exit 2
  fi
  RESOLVED_TARGETS+=("$abs")
done

# ---- collect SKILL.md files ----
SKILL_FILES=()
while IFS= read -r f; do
  [ -n "$f" ] && SKILL_FILES+=("$f")
done < <(find "${RESOLVED_TARGETS[@]}" -type f -name "SKILL.md" 2>/dev/null | sort)

# ---- helpers ----
extract_description() {
  # Print the `description:` value from a SKILL.md frontmatter block.
  # Assumes the project convention: single-line, double-quoted value.
  local file="$1"
  awk '
    /^---$/ { fm++; if (fm == 2) exit; next }
    fm == 1 && /^description:/ { print; exit }
  ' "$file" | sed -E 's/^description:[[:space:]]*//; s/^"//; s/"[[:space:]]*$//'
}

truncate_snippet() {
  local s="$1"
  if [ "${#s}" -gt "$SNIPPET_MAX" ]; then
    printf '%s...' "${s:0:$SNIPPET_MAX}"
  else
    printf '%s' "$s"
  fi
}

VIOLATION_COUNT=0
FILES_WITH_VIOLATIONS=0

report_violation() {
  local file="$1"
  local kind="$2"
  local snippet="$3"
  local rel="${file#${REPO_ROOT}/}"
  printf '%s\t%s\t%s\n' "$rel" "$kind" "$(truncate_snippet "$snippet")"
  VIOLATION_COUNT=$((VIOLATION_COUNT + 1))
}

# ---- main scan ----
# Guard: bash 3 on macOS treats "${arr[@]}" as unbound when the array is empty
# under `set -u`. Skip the loop when there are no files.
if [ "${#SKILL_FILES[@]}" -gt 0 ]; then
for file in "${SKILL_FILES[@]}"; do
  desc="$(extract_description "$file" || true)"
  had_violation=0

  if [ -z "$desc" ]; then
    report_violation "$file" "no-description" "<missing or unparseable>"
    FILES_WITH_VIOLATIONS=$((FILES_WITH_VIOLATIONS + 1))
    continue
  fi

  # 1. Required prefix check
  if [[ "$desc" != "${REQUIRED_PREFIX}"* ]]; then
    report_violation "$file" "missing-use-when-prefix" "$desc"
    had_violation=1
  fi

  # 2. Forbidden phrase checks (case-sensitive; rule specifies exact casing)
  for phrase in "${FORBIDDEN_PHRASES[@]}"; do
    if [[ "$desc" == *"$phrase"* ]]; then
      report_violation "$file" "forbidden-phrase:${phrase}" "$desc"
      had_violation=1
    fi
  done

  # 3. Length check
  len="${#desc}"
  if [ "$len" -gt "$MAX_LEN" ]; then
    report_violation "$file" "over-length:${len}" "$desc"
    had_violation=1
  fi

  if [ "$had_violation" -eq 1 ]; then
    FILES_WITH_VIOLATIONS=$((FILES_WITH_VIOLATIONS + 1))
  fi
done
fi

# ---- summary to stderr ----
{
  echo ""
  echo "Scanned ${#SKILL_FILES[@]} SKILL.md file(s)."
  if [ "$VIOLATION_COUNT" -eq 0 ]; then
    echo "All descriptions conform to .claude/rules/skill-description.md."
  else
    echo "${VIOLATION_COUNT} violation(s) across ${FILES_WITH_VIOLATIONS} file(s)."
    echo "See .claude/rules/skill-description.md for the authoritative format."
  fi
} >&2

if [ "$VIOLATION_COUNT" -eq 0 ]; then
  exit 0
else
  exit 1
fi
