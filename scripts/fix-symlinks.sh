#!/bin/bash
# fix-symlinks.sh
# Detect broken symlinks / plain-text link projections on Windows and auto-repair with actual copies
#
# Purpose: Called from session-init.sh
# Behavior:
#   - When public harness-* skills in skills/ are plain files (old Windows checkout)
#   - Replace with actual copies from skills/
#   - Output repair count to stdout (JSON format)
#
# Output:
#   {"fixed": N, "checked": M, "details": ["harness-work", ...]}

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SKILLS_DIR="$PLUGIN_ROOT/skills"
SKILLS_V3_DIR="$PLUGIN_ROOT/skills"

# Public 5 skills list (skills/ mirror bundle)
V3_SKILLS=("harness-plan" "harness-work" "harness-review" "harness-setup" "harness-release")

FIXED=0
CHECKED=0
FIXED_NAMES=()

for skill in "${V3_SKILLS[@]}"; do
  CHECKED=$((CHECKED + 1))
  skill_path="$SKILLS_DIR/$skill"
  source_path="$SKILLS_V3_DIR/$skill"

  # Normal: exists as symlink or directory -> skip
  if [ -d "$skill_path" ]; then
    continue
  fi

  # Broken plain-text link: exists as regular file (occurs on Windows git clone)
  if [ -f "$skill_path" ]; then
    # Check if repair source exists
    if [ -d "$source_path" ]; then
      rm -f "$skill_path"
      cp -r "$source_path" "$skill_path"
      FIXED=$((FIXED + 1))
      FIXED_NAMES+=("$skill")
    fi
  fi

  # Attempt repair even if not found
  if [ ! -e "$skill_path" ] && [ -d "$source_path" ]; then
    cp -r "$source_path" "$skill_path"
    FIXED=$((FIXED + 1))
    FIXED_NAMES+=("$skill")
  fi
done

# Check symlinks in extensions/ similarly
EXTENSIONS_DIR="$SKILLS_V3_DIR/extensions"
if [ -d "$EXTENSIONS_DIR" ]; then
  for ext_path in "$EXTENSIONS_DIR"/*; do
    [ -e "$ext_path" ] || continue
    ext_name="$(basename "$ext_path")"
    CHECKED=$((CHECKED + 1))

    # When regular file (broken symlink)
    if [ -f "$ext_path" ] && [ ! -d "$ext_path" ]; then
      # Read link target (file content is path)
      target=$(cat "$ext_path" 2>/dev/null || true)
      # Resolve relative path
      resolved="$(cd "$EXTENSIONS_DIR" && cd "$(dirname "$target")" 2>/dev/null && pwd)/$(basename "$target")" 2>/dev/null || true
      if [ -d "$resolved" ]; then
        rm -f "$ext_path"
        cp -r "$resolved" "$ext_path"
        FIXED=$((FIXED + 1))
        FIXED_NAMES+=("extensions/$ext_name")
      fi
    fi
  done
fi

# JSON output
NAMES_JSON="[]"
if [ ${#FIXED_NAMES[@]} -gt 0 ]; then
  NAMES_JSON="["
  for i in "${!FIXED_NAMES[@]}"; do
    [ "$i" -gt 0 ] && NAMES_JSON+=","
    NAMES_JSON+="\"${FIXED_NAMES[$i]}\""
  done
  NAMES_JSON+="]"
fi

echo "{\"fixed\":${FIXED},\"checked\":${CHECKED},\"details\":${NAMES_JSON}}"
