#!/usr/bin/env bash
# skill-description-drift-report.sh — Detect description drift between
# skills/ and opencode/skills/
#
# Purpose:
#   For every skill that exists in both `skills/<name>/SKILL.md` and
#   `opencode/skills/<name>/SKILL.md`, compare the description: field.
#   Descriptions are expected to match (same skill -> same trigger shape).
#   Divergence indicates drift that should either be intentional or
#   reconciled.
#
#   This tool only reports. It never overwrites (Phase 44.7 DoD: "no
#   automatic overwrite").
#
# Usage:
#   bash scripts/skill-description-drift-report.sh
#
# Exit codes:
#   0 — no drift
#   1 — one or more divergent pairs
#   2 — invocation error (e.g., required directory missing)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

SKILLS_DIR="${REPO_ROOT}/skills"
OPENCODE_DIR="${REPO_ROOT}/opencode/skills"

if [ ! -d "$SKILLS_DIR" ] || [ ! -d "$OPENCODE_DIR" ]; then
  echo "Error: required directory missing ($SKILLS_DIR or $OPENCODE_DIR)" >&2
  exit 2
fi

exec python3 - "$SKILLS_DIR" "$OPENCODE_DIR" <<'PYEOF'
import re
import sys
from pathlib import Path

skills_dir = Path(sys.argv[1])
opencode_dir = Path(sys.argv[2])

DESC_RE = re.compile(r'^description:\s*"(.*)"\s*$', re.MULTILINE)

def extract_desc(path: Path):
    try:
        m = DESC_RE.search(path.read_text())
    except OSError:
        return None
    return m.group(1) if m else None

pairs = []
total = 0
for opencode_skill in sorted(p for p in opencode_dir.iterdir() if p.is_dir()):
    name = opencode_skill.name
    skill_file = skills_dir / name / "SKILL.md"
    opencode_file = opencode_skill / "SKILL.md"
    if not (skill_file.exists() and opencode_file.exists()):
        continue
    total += 1
    s_desc = extract_desc(skill_file)
    o_desc = extract_desc(opencode_file)
    if s_desc == o_desc:
        continue
    pairs.append((name, s_desc, o_desc))

print("## Drift Report - skills/ vs opencode/skills/ descriptions")
print()
print(f"Total overlapping skills scanned: {total}")
print(f"Divergent description pairs: {len(pairs)}")
print()

if not pairs:
    print("No divergence detected. All overlapping skills share the same")
    print("description in both trees.")
    sys.exit(0)

for name, s, o in pairs:
    print(f"### {name} - DIVERGENT")
    print(f"  skills/:   {s}")
    print(f"  opencode/: {o}")
    print(f"  Recommended action: reconcile (decide whether the divergence")
    print(f"  is intentional; if not, sync the two descriptions).")
    print()

sys.exit(1)
PYEOF
