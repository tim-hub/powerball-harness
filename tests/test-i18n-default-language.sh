#!/usr/bin/env bash
#
# Verify the Phase 55 language contract defaults new distribution surfaces to
# English while keeping Japanese as an explicit opt-in.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

python3 - <<'PY'
import json
from pathlib import Path

root = Path(".")

schema = json.loads(Path("claude-code-harness.config.schema.json").read_text(encoding="utf-8"))
language = schema["properties"]["i18n"]["properties"]["language"]
enum = set(language.get("enum", []))
assert language.get("default") == "en", "schema i18n.language default must be en"
assert enum == {"en", "ja"}, f"schema i18n.language enum must keep en and ja, got {sorted(enum)!r}"

example = json.loads(Path("claude-code-harness.config.example.json").read_text(encoding="utf-8"))
assert example["i18n"]["language"] == "en", "example config must default to en"
assert "ja" in json.dumps(example["i18n"], ensure_ascii=False), "example config must still document ja opt-in"

contract = Path("docs/i18n-language-contract.md").read_text(encoding="utf-8")
assert "User-facing default locale is `en`." in contract
assert "Japanese remains supported" in contract
assert "`description-en` and `description-ja`" in contract
assert "Default | `en`" in contract

template = Path("templates/.claude-code-harness.config.yaml.template").read_text(encoding="utf-8")
assert "i18n:" in template, "setup config template must include i18n"
assert "language: en" in template, "setup config template must render English default"
assert "language: ja" in template or "`ja`" in template or "Japanese" in template, "setup config template must mention Japanese opt-in"

print("i18n default language contract ok")
PY

echo "✓ i18n default language surfaces are English by default and keep Japanese opt-in"
