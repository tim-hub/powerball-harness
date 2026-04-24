#!/usr/bin/env bash
#
# Verify setup-facing language templates render the English default without
# dropping the Japanese opt-in path.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

rendered_config="$tmpdir/.claude-code-harness.config.yaml"
cp templates/.claude-code-harness.config.yaml.template "$rendered_config"

if grep -q '{{[A-Z_]*}}' "$rendered_config"; then
  echo "config template rendered with unresolved placeholders" >&2
  exit 1
fi

python3 - "$rendered_config" <<'PY'
import re
import sys
from pathlib import Path

text = Path(sys.argv[1]).read_text(encoding="utf-8")
assert re.search(r"(?m)^i18n:\s*$", text), "rendered config must include i18n section"
assert re.search(r"(?m)^\s+language:\s+en\s*$", text), "rendered config must default i18n.language to en"
assert "ja" in text and "Japanese" in text, "rendered config must keep Japanese opt-in guidance"
PY

for template in templates/AGENTS.md.template templates/CLAUDE.md.template templates/Plans.md.template; do
  rendered="$tmpdir/$(basename "$template" .template)"
  sed \
    -e "s|{{PROJECT_NAME}}|sample-project|g" \
    -e "s|{{DATE}}|2026-04-24|g" \
    -e "s|{{LANGUAGE}}|en|g" \
    "$template" > "$rendered"
  if grep -q '{{[A-Z_]*}}' "$rendered"; then
    echo "$template rendered with unresolved placeholders" >&2
    exit 1
  fi
done

test -f templates/modes/harness--ja.json

echo "✓ setup language rendering keeps English default config and Japanese opt-in assets"
