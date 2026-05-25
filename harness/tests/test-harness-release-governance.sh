#!/bin/bash
# Verify harness-release SKILL.md contains the bare-invocation governance contract:
# review gate before release, Work Commit Gate, AskUserQuestion for unreviewed work.
# Also verifies P27 AUTO-START contract: RELEASE_AUTOSTART: marker, ARGUMENTS == "" condition,
# and forbidden-action literals (prevents silent stalls on bare invocation).

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

skill_file="$ROOT_DIR/skills/harness-release/SKILL.md"

required_terms=(
  "AskUserQuestion"
  "commit and release the work done so far"
  "Bare invocation contract"
  "Review Gate"
  "Work Commit Gate"
  "Start with review"
  "harness-review"
  "APPROVE"
  "REQUEST_CHANGES"
  "harness-work"
  "repeat until APPROVE"
  "Do not treat REQUEST_CHANGES alone as a terminal stop"
  "release dry-run"
  "working tree clean check"
  "RELEASE_AUTOSTART:"
  'ARGUMENTS == ""'
  "no tasks found"
)

failures=0

if [ ! -f "$skill_file" ]; then
  echo "missing release skill file: ${skill_file#$ROOT_DIR/}" >&2
  exit 1
fi

for term in "${required_terms[@]}"; do
  if ! grep -Fq "$term" "$skill_file"; then
    echo "missing required term in ${skill_file#$ROOT_DIR/}: $term" >&2
    failures=$((failures + 1))
  fi
done

if ! grep -Eq '^allowed-tools: .*AskUserQuestion' "$skill_file"; then
  echo "AskUserQuestion is not in allowed-tools: ${skill_file#$ROOT_DIR/}" >&2
  failures=$((failures + 1))
fi

if ! grep -Eq '^allowed-tools: .*Skill' "$skill_file"; then
  echo "Skill tool is not in allowed-tools (needed for harness-review handoff): ${skill_file#$ROOT_DIR/}" >&2
  failures=$((failures + 1))
fi

if python3 -c "import sys, re; data=open(sys.argv[1]).read(); sys.exit(0 if re.search(r'[぀-ヿ一-鿿]', data) else 1)" "$skill_file" 2>/dev/null; then
  echo "Japanese characters found in ${skill_file#$ROOT_DIR/}: port must be English-only" >&2
  failures=$((failures + 1))
fi

dropped_docs=(
  "docs/hokage-spin-off-readiness"
  "docs/agent-view-policy"
)
for doc in "${dropped_docs[@]}"; do
  if grep -Fq "$doc" "$skill_file"; then
    echo "reference to dropped doc found: $doc in ${skill_file#$ROOT_DIR/}" >&2
    failures=$((failures + 1))
  fi
done

# P35 Layer 2: instruction line literal must appear in both harness-release and harness-review
layer2_literal="Claude will summarize this result"
layer2_files=(
  "$ROOT_DIR/skills/harness-release/SKILL.md"
  "$ROOT_DIR/skills/harness-review/SKILL.md"
)
for layer2_file in "${layer2_files[@]}"; do
  if [ ! -f "$layer2_file" ]; then
    echo "missing skill file for Layer 2 literal check: ${layer2_file#$ROOT_DIR/}" >&2
    failures=$((failures + 1))
    continue
  fi
  if ! grep -Fq "$layer2_literal" "$layer2_file"; then
    echo "missing P35 Layer 2 instruction literal in ${layer2_file#$ROOT_DIR/}: '$layer2_literal'" >&2
    failures=$((failures + 1))
  fi
done

if [ "$failures" -gt 0 ]; then
  exit 1
fi

echo "test-harness-release-governance: ok"
