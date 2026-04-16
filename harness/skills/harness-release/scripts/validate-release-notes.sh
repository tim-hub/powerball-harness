#!/bin/bash
# validate-release-notes.sh
# GitHub Release notes format validation script
# Usage: ./scripts/validate-release-notes.sh [tag]
# Example: ./scripts/validate-release-notes.sh v2.10.0

set -e

TAG="${1:-}"
ERRORS=0

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_error() {
    echo -e "${RED}❌ $1${NC}"
    ERRORS=$((ERRORS + 1))
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_ok() {
    echo -e "${GREEN}✅ $1${NC}"
}

# If no tag is specified, check the latest release
if [ -z "$TAG" ]; then
    TAG=$(gh release list --limit 1 --json tagName -q '.[0].tagName')
    echo "📋 Checking latest release: $TAG"
fi

# Get release notes
NOTES=$(gh release view "$TAG" --json body -q '.body' 2>/dev/null)

if [ -z "$NOTES" ]; then
    log_error "Release not found: $TAG"
    exit 1
fi

echo ""
echo "🔍 Release notes validation: $TAG"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 1. Heading check
if echo "$NOTES" | grep -qE "^## (What's Changed|What's New|Summary)"; then
    log_ok "Heading: present"
else
    log_error "Missing heading: ## What's Changed"
fi

# 2. Before / After table check
if echo "$NOTES" | grep -qE "Before.*(→|/|\\|).*After"; then
    log_ok "Before / After table: present"
else
    log_error "Before / After table is missing"
fi

# 3. Footer check
if echo "$NOTES" | grep -q "Generated with \[Claude Code\]"; then
    log_ok "Footer: present"
else
    log_warn "Footer not found (recommended: '🤖 Generated with [Claude Code](...)')"
fi

# 4. Section check
for section in "Added" "Changed" "Fixed" "Security"; do
    if echo "$NOTES" | grep -q "^## $section"; then
        log_ok "Section: $section present"
    fi
done

# 6. Bold summary check
if echo "$NOTES" | head -10 | grep -qE "^\*\*.+\*\*$"; then
    log_ok "Bold summary: present"
else
    log_warn "Bold summary not found (describe the value of the change in one line)"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}Validation result: $ERRORS error(s)${NC}"
    echo ""
    echo "📖 Reference: .claude/rules/github-release.md"
    exit 1
else
    echo -e "${GREEN}Validation result: all checks passed${NC}"
    exit 0
fi
