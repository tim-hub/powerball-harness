#!/bin/bash
# check-translations.sh
# Check that shipped commands and skills have i18n translation fields.
#
# Usage: ./scripts/i18n/check-translations.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🌐 Checking i18n translations..."
echo ""

missing_count=0
total_count=0
skill_error_count=0

extract_frontmatter_value() {
  local file="$1"
  local key="$2"

  awk -v key="$key" '
    NR == 1 {
      if ($0 != "---") {
        exit 3
      }
      next
    }
    $0 == "---" {
      exit 2
    }
    index($0, key ":") == 1 {
      value = substr($0, length(key) + 2)
      sub(/^[[:space:]]*/, "", value)
      print value
      exit 0
    }
    NR > 80 {
      exit 4
    }
  ' "$file"
}

# Check legacy commands for description-en
echo "📁 Commands (legacy):"
while IFS= read -r file; do
  total_count=$((total_count + 1))
  relative_path="${file#$PROJECT_ROOT/}"
  if ! grep -q "^description-en:" "$file"; then
    echo -e "  ${RED}✗${NC} $relative_path (missing description-en)"
    missing_count=$((missing_count + 1))
  else
    echo -e "  ${GREEN}✓${NC} $relative_path"
  fi
done < <(find "$PROJECT_ROOT/commands" -type f -name "*.md" 2>/dev/null | sort)

echo ""

# Check shipped skills for complete i18n metadata and English default.
skill_missing=0
skill_total=0

check_skill_surface() {
  local skills_dir="$1"
  local label="$2"

  if [[ ! -d "$skills_dir" ]]; then
    return
  fi

  echo "📁 ${label}:"

  while IFS= read -r file; do
    skill_total=$((skill_total + 1))
    total_count=$((total_count + 1))
    relative_path="${file#$PROJECT_ROOT/}"

    local desc desc_en desc_ja ok
    desc="$(extract_frontmatter_value "$file" "description" || true)"
    desc_en="$(extract_frontmatter_value "$file" "description-en" || true)"
    desc_ja="$(extract_frontmatter_value "$file" "description-ja" || true)"
    ok=1

    if [[ -z "$desc" ]]; then
      echo -e "  ${RED}✗${NC} $relative_path (missing description)"
      ok=0
    fi
    if [[ -z "$desc_en" ]]; then
      echo -e "  ${RED}✗${NC} $relative_path (missing description-en)"
      ok=0
    fi
    if [[ -z "$desc_ja" ]]; then
      echo -e "  ${RED}✗${NC} $relative_path (missing description-ja)"
      ok=0
    fi
    if [[ -n "$desc" && -n "$desc_en" && "$desc" != "$desc_en" ]]; then
      echo -e "  ${RED}✗${NC} $relative_path (description must equal description-en for shipped English default)"
      ok=0
    fi

    if [[ "$ok" -eq 1 ]]; then
      echo -e "  ${GREEN}✓${NC} $relative_path"
    else
      skill_missing=$((skill_missing + 1))
      skill_error_count=$((skill_error_count + 1))
    fi
  done < <(find "$skills_dir" -mindepth 2 -maxdepth 2 -type f -name "SKILL.md" 2>/dev/null | sort)

  echo ""
}

check_skill_surface "$PROJECT_ROOT/skills" "skills"
check_skill_surface "$PROJECT_ROOT/skills-codex" "skills-codex"
check_skill_surface "$PROJECT_ROOT/codex/.codex/skills" "codex/.codex/skills"
check_skill_surface "$PROJECT_ROOT/opencode/skills" "opencode/skills"

missing_count=$((missing_count + skill_error_count))

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [[ $missing_count -eq 0 ]]; then
  echo -e "${GREEN}✓ All $total_count files have translations${NC}"
  exit 0
else
  echo -e "${YELLOW}⚠ $missing_count / $total_count files have i18n errors${NC}"
  if [[ $skill_missing -gt 0 ]]; then
    echo -e "${YELLOW}  Skills with i18n errors: $skill_missing / $skill_total${NC}"
  fi
  exit 1
fi
