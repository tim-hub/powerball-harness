#!/usr/bin/env bash
#
# Verify locale switching is idempotent on a temp copy and never dirties this
# repository while checking the ja -> en roundtrip.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

snapshot_file() {
  local output="$1"
  find skills skills-codex codex/.codex/skills opencode/skills -mindepth 2 -maxdepth 2 -type f -name "SKILL.md" \
    | sort \
    | while IFS= read -r file; do
        cksum "$file"
      done > "$output"
}

verify_locale_copy() {
  local locale="$1"
  local root="$2"

  python3 - "$locale" "$root" <<'PY'
import sys
from pathlib import Path

locale = sys.argv[1]
root = Path(sys.argv[2])
target_key = "description-ja" if locale == "ja" else "description-en"
surfaces = [
    root / "skills",
    root / "skills-codex",
    root / "codex/.codex/skills",
    root / "opencode/skills",
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


count = 0
for surface in surfaces:
    for path in sorted(surface.glob("*/SKILL.md")):
        count += 1
        meta = frontmatter(path)
        assert meta.get("description") == meta.get(target_key), (
            f"{path}: description must equal {target_key} after locale {locale}"
        )
        assert meta.get("description-en"), f"{path}: description-en was lost"
        assert meta.get("description-ja"), f"{path}: description-ja was lost"

assert count > 0, "no skill files were checked"
print(f"{locale}: checked {count} skills")
PY
}

before="$(mktemp)"
after="$(mktemp)"
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir" "$before" "$after"' EXIT

snapshot_file "$before"

mkdir -p "$tmpdir/repo/scripts/i18n" "$tmpdir/repo/codex/.codex"
cp scripts/i18n/set-locale.sh "$tmpdir/repo/scripts/i18n/set-locale.sh"
cp -R skills "$tmpdir/repo/skills"
cp -R skills-codex "$tmpdir/repo/skills-codex"
cp -R codex/.codex/skills "$tmpdir/repo/codex/.codex/skills"
mkdir -p "$tmpdir/repo/opencode"
cp -R opencode/skills "$tmpdir/repo/opencode/skills"

(
  cd "$tmpdir/repo"
  bash scripts/i18n/set-locale.sh ja >/tmp/i18n-roundtrip-ja.$$ 2>&1
  verify_locale_copy ja "$tmpdir/repo"
  bash scripts/i18n/set-locale.sh en >/tmp/i18n-roundtrip-en.$$ 2>&1
  verify_locale_copy en "$tmpdir/repo"
  bash scripts/i18n/set-locale.sh en >/tmp/i18n-roundtrip-en2.$$ 2>&1
  verify_locale_copy en "$tmpdir/repo"
)
rm -f /tmp/i18n-roundtrip-ja.$$ /tmp/i18n-roundtrip-en.$$ /tmp/i18n-roundtrip-en2.$$

snapshot_file "$after"
if ! diff -u "$before" "$after"; then
  echo "real repository skill files changed during temp-copy roundtrip" >&2
  exit 1
fi

echo "✓ locale roundtrip is idempotent on temp copy and leaves repo files untouched"
