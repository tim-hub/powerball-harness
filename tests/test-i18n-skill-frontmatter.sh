#!/usr/bin/env bash
#
# Verify every shipped skill surface keeps bilingual metadata and ships with
# description set to the English default.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

bash scripts/i18n/check-translations.sh

python3 - <<'PY'
from pathlib import Path

SURFACES = [
    Path("skills"),
    Path("skills-codex"),
    Path("codex/.codex/skills"),
    Path("opencode/skills"),
]


def frontmatter(path: Path) -> dict[str, str]:
    lines = path.read_text(encoding="utf-8").splitlines()
    if not lines or lines[0] != "---":
        raise AssertionError(f"{path}: missing frontmatter")
    data: dict[str, str] = {}
    for line in lines[1:]:
        if line == "---":
            return data
        if ":" not in line:
            continue
        key, value = line.split(":", 1)
        data[key] = value.strip()
    raise AssertionError(f"{path}: unterminated frontmatter")


skill_count = 0
all_text = []
for surface in SURFACES:
    files = sorted(surface.glob("*/SKILL.md"))
    assert files, f"{surface}: no SKILL.md files found"
    for path in files:
        skill_count += 1
        meta = frontmatter(path)
        for key in ("description", "description-en", "description-ja"):
            assert meta.get(key), f"{path}: missing or empty {key}"
        assert meta["description"] == meta["description-en"], f"{path}: description must equal description-en"
        all_text.append(path.read_text(encoding="utf-8"))

joined = "\n".join(all_text)
for phrase in ("実装して", "レビューして", "計画作って"):
    assert phrase in joined, f"Japanese trigger phrase disappeared: {phrase}"

print(f"validated {skill_count} shipped skill files")
PY

echo "✓ shipped skill frontmatter preserves English default and Japanese routing"
