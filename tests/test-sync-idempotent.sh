#!/bin/bash
# test-sync-idempotent.sh
# harness sync を 3 回連続実行しても plugin.json が安定することを検証
#
# Usage: bash tests/test-sync-idempotent.sh

set -euo pipefail

PLUGIN_JSON=".claude-plugin/plugin.json"

# 初回 sync
bash bin/harness sync > /dev/null
SHA1_INITIAL=$(shasum -a 256 "$PLUGIN_JSON" | awk '{print $1}')

# 連続 2 回 sync
bash bin/harness sync > /dev/null
SHA1_AFTER_2ND=$(shasum -a 256 "$PLUGIN_JSON" | awk '{print $1}')

bash bin/harness sync > /dev/null
SHA1_AFTER_3RD=$(shasum -a 256 "$PLUGIN_JSON" | awk '{print $1}')

# checksum 一致確認
if [ "$SHA1_INITIAL" != "$SHA1_AFTER_2ND" ] || [ "$SHA1_INITIAL" != "$SHA1_AFTER_3RD" ]; then
  echo "FAIL: plugin.json checksum changed across sync runs"
  echo "  initial: $SHA1_INITIAL"
  echo "  after 2nd: $SHA1_AFTER_2ND"
  echo "  after 3rd: $SHA1_AFTER_3RD"
  exit 1
fi

# 公式 SSOT 外 field の不在確認
if grep -q '"monitors"' "$PLUGIN_JSON"; then
  echo "FAIL: plugin.json contains monitors field (should be in monitors/monitors.json)"
  exit 1
fi
if grep -q '"agents"' "$PLUGIN_JSON"; then
  echo "FAIL: plugin.json contains agents field (should be auto-discovery from agents/)"
  exit 1
fi

echo "PASS: harness sync is idempotent and emits no phantom fields"
