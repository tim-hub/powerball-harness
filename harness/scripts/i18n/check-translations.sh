#!/bin/bash
# check-translations.sh
# Check that all commands and skills have i18n translation fields
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

# Check legacy commands for description-en
echo "📁 Commands (legacy):"
for file in "$PROJECT_ROOT"/commands/**/*.md; do
  if [[ -f "$file" ]]; then
    total_count=$((total_count + 1))
    relative_path="${file#$PROJECT_ROOT/}"
    if ! grep -q "description-en:" "$file"; then
      echo -e "  ${RED}✗${NC} $relative_path (missing description-en)"
      missing_count=$((missing_count + 1))
    else
      echo -e "  ${GREEN}✓${NC} $relative_path"
    fi
  fi
done

echo ""

# Check skills for description-ja
skill_missing=0
skill_total=0

echo "📁 Skills:"
for file in "$PROJECT_ROOT"/skills/*/SKILL.md; do
  if [[ -f "$file" ]]; then
    skill_total=$((skill_total + 1))
    total_count=$((total_count + 1))
    relative_path="${file#$PROJECT_ROOT/}"
    if ! grep -q "description-ja:" "$file"; then
      echo -e "  ${RED}✗${NC} $relative_path (missing description-ja)"
      skill_missing=$((skill_missing + 1))
      missing_count=$((missing_count + 1))
    else
      echo -e "  ${GREEN}✓${NC} $relative_path"
    fi
  fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [[ $missing_count -eq 0 ]]; then
  echo -e "${GREEN}✓ All $total_count files have translations${NC}"
  exit 0
else
  echo -e "${YELLOW}⚠ $missing_count / $total_count files missing translations${NC}"
  if [[ $skill_missing -gt 0 ]]; then
    echo -e "${YELLOW}  Skills missing description-ja: $skill_missing / $skill_total${NC}"
  fi
  exit 1
fi
