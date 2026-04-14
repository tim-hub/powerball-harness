#!/bin/bash
# plans-format-migrate.sh
# Migrate Plans.md from old format to new format

set -uo pipefail

PLANS_FILE="${1:-Plans.md}"
DRY_RUN="${2:-false}"

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}Plans.md Format Migration${NC}"
echo "=========================================="
echo ""

# If Plans.md does not exist
if [ ! -f "$PLANS_FILE" ]; then
  echo -e "${RED}Error: $PLANS_FILE not found${NC}"
  exit 1
fi

# Create backup
BACKUP_DIR=".claude-code-harness/backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp "$PLANS_FILE" "$BACKUP_DIR/Plans.md.backup"
echo -e "${GREEN}✓${NC} Backup created: $BACKUP_DIR/Plans.md.backup"

# Change count
CHANGES=0

# 1. cursor:WIP → pm:pending (interpreted as PM review pending state)
# Note: cursor:WIP typically means "PM (Cursor) is reviewing"
# In the new format this corresponds to pm:pending (implementation done, awaiting PM review)
if grep -qE 'cursor:WIP' "$PLANS_FILE" 2>/dev/null; then
  echo -e "${YELLOW}→${NC} cursor:WIP detected"
  if [ "$DRY_RUN" = "false" ]; then
    sed -i '' 's/cursor:WIP/pm:pending/g' "$PLANS_FILE" 2>/dev/null || \
    sed -i 's/cursor:WIP/pm:pending/g' "$PLANS_FILE"
    echo -e "  ${GREEN}✓${NC} cursor:WIP → pm:pending converted"
  else
    echo -e "  [DRY RUN] cursor:WIP → pm:pending will be converted"
  fi
  ((CHANGES++))
fi

# 2. cursor:done → pm:confirmed
if grep -qE 'cursor:done' "$PLANS_FILE" 2>/dev/null; then
  echo -e "${YELLOW}→${NC} cursor:done detected"
  if [ "$DRY_RUN" = "false" ]; then
    sed -i '' 's/cursor:done/pm:confirmed/g' "$PLANS_FILE" 2>/dev/null || \
    sed -i 's/cursor:done/pm:confirmed/g' "$PLANS_FILE"
    echo -e "  ${GREEN}✓${NC} cursor:done → pm:confirmed converted"
  else
    echo -e "  [DRY RUN] cursor:done → pm:confirmed will be converted"
  fi
  ((CHANGES++))
fi

# 3. Check for marker legend section update
if ! grep -qE '## Marker Legend' "$PLANS_FILE" 2>/dev/null; then
  echo -e "${YELLOW}→${NC} Marker legend section is missing"
  echo -e "  ${YELLOW}!${NC} Recommended to add manually"
fi

# Show result
echo ""
echo "=========================================="
if [ $CHANGES -gt 0 ]; then
  if [ "$DRY_RUN" = "false" ]; then
    echo -e "${GREEN}✓ Migration complete: $CHANGES change(s)${NC}"
    echo ""
    echo "Please review the changes:"
    echo "  git diff $PLANS_FILE"
  else
    echo -e "${YELLOW}DRY RUN: $CHANGES change(s) planned${NC}"
    echo ""
    echo "To actually convert:"
    echo "  ./scripts/plans-format-migrate.sh $PLANS_FILE false"
  fi
else
  echo -e "${GREEN}✓ No changes needed. Format is up to date.${NC}"
fi
