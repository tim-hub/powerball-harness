#!/bin/bash
# setup-existing-project.sh
# Setup script to apply claude-code-harness to an existing project
#
# Usage: ./scripts/setup-existing-project.sh [project_path]
#
# Cross-platform: Supports Windows (Git Bash/MSYS2/Cygwin/WSL), macOS, Linux

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HARNESS_ROOT="$(dirname "$SCRIPT_DIR")"

# Load cross-platform path utilities
if [ -f "$SCRIPT_DIR/path-utils.sh" ]; then
  # shellcheck source=./path-utils.sh
  source "$SCRIPT_DIR/path-utils.sh"
fi

PROJECT_PATH="${1:-.}"
# Normalize project path for cross-platform compatibility
if type normalize_path &>/dev/null; then
  PROJECT_PATH="$(normalize_path "$PROJECT_PATH")"
fi

# Color output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Claude harness - Apply to Existing Project${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# ================================
# Step 1: Prerequisite check
# ================================

echo -e "${BLUE}[1/6] Prerequisite check${NC}"
echo "----------------------------------------"

# Verify project directory exists
if [ ! -d "$PROJECT_PATH" ]; then
    echo -e "${RED}✗ Project directory not found: $PROJECT_PATH${NC}"
    exit 1
fi

cd "$PROJECT_PATH" || {
    echo -e "${RED}✗ Cannot change to directory: $PROJECT_PATH${NC}"
    exit 1
}
PROJECT_PATH=$(pwd)
echo -e "${GREEN}✓${NC} Project directory: $PROJECT_PATH"

# Setup metadata
PROJECT_NAME="$(basename "$PROJECT_PATH")"
SETUP_DATE_ISO="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
SETUP_DATE_SHORT="$(date +"%Y-%m-%d")"
HARNESS_VERSION="unknown"
if [ -f "$HARNESS_ROOT/VERSION" ]; then
    HARNESS_VERSION="$(cat "$HARNESS_ROOT/VERSION" | tr -d ' \n\r')"
fi

# Template placeholder (may be overwritten by analyze-project result)
LANGUAGE="unknown"

# Check if this is a Git repository
if [ ! -d ".git" ]; then
    echo -e "${YELLOW}⚠${NC}  Not a Git repository"
    read -p "Initialize a Git repository? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git init
        echo -e "${GREEN}✓${NC} Git repository initialized"
    fi
else
    echo -e "${GREEN}✓${NC} Git repository detected"
fi

# Check for uncommitted changes
if [ -d ".git" ]; then
    if ! git diff-index --quiet HEAD -- 2>/dev/null; then
        echo -e "${YELLOW}⚠${NC}  Uncommitted changes detected"
        echo ""
        echo -e "${YELLOW}Recommended: commit before running setup${NC}"
        echo ""
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Setup cancelled"
            exit 0
        fi
    else
        echo -e "${GREEN}✓${NC} Working tree is clean"
    fi
fi

echo ""

# ================================
# Step 2: Discover existing specs / documents
# ================================

echo -e "${BLUE}[2/6] Discovering existing documents${NC}"
echo "----------------------------------------"

FOUND_DOCS=()
DOC_PATTERNS=(
    "README.md"
    "SPEC.md"
    "SPECIFICATION.md"
    "specs.md"
    "requirements.md"
    "docs/spec.md"
    "docs/specification.md"
    "docs/requirements.md"
    "docs/proposal.md"
    "docs/proposal-doc.md"
    "Plans.md"
    "PLAN.md"
    "plan.md"
)

for pattern in "${DOC_PATTERNS[@]}"; do
    if [ -f "$pattern" ]; then
        FOUND_DOCS+=("$pattern")
        echo -e "${GREEN}✓${NC} Found: $pattern"
    fi
done

if [ ${#FOUND_DOCS[@]} -eq 0 ]; then
    echo -e "${YELLOW}⚠${NC}  No existing specification documents found"
else
    echo ""
    echo -e "${GREEN}${#FOUND_DOCS[@]} document(s) found${NC}"
fi

echo ""

# ================================
# Step 3: Project analysis
# ================================

echo -e "${BLUE}[3/6] Project analysis${NC}"
echo "----------------------------------------"

# Run analyze-project.sh
if [ -f "$HARNESS_ROOT/scripts/analyze-project.sh" ]; then
    ANALYSIS_RESULT=$("$HARNESS_ROOT/scripts/analyze-project.sh" "$PROJECT_PATH" 2>/dev/null || echo "{}")

    # Display detected tech stack (analyze-project.sh output: technologies/frameworks/testing)
    if command -v jq &> /dev/null; then
        TECHNOLOGIES=$(echo "$ANALYSIS_RESULT" | jq -r '.technologies[]?' 2>/dev/null || true)
        FRAMEWORKS=$(echo "$ANALYSIS_RESULT" | jq -r '.frameworks[]?' 2>/dev/null || true)
        TESTING=$(echo "$ANALYSIS_RESULT" | jq -r '.testing[]?' 2>/dev/null || true)

        # Simple LANGUAGE estimation (for template placeholder {{LANGUAGE}})
        LANGUAGE=$(echo "$ANALYSIS_RESULT" | jq -r '.technologies[0] // "unknown"' 2>/dev/null || echo "unknown")

        if [ -n "${TECHNOLOGIES}${FRAMEWORKS}${TESTING}" ]; then
            echo "Detection results:"
            if [ -n "$TECHNOLOGIES" ]; then
                echo "  technologies:"
                echo "$TECHNOLOGIES" | while read -r tech; do
                    [ -n "$tech" ] && echo -e "    ${GREEN}•${NC} $tech"
                done
            fi
            if [ -n "$FRAMEWORKS" ]; then
                echo "  frameworks:"
                echo "$FRAMEWORKS" | while read -r fw; do
                    [ -n "$fw" ] && echo -e "    ${GREEN}•${NC} $fw"
                done
            fi
            if [ -n "$TESTING" ]; then
                echo "  testing:"
                echo "$TESTING" | while read -r t; do
                    [ -n "$t" ] && echo -e "    ${GREEN}•${NC} $t"
                done
            fi
        fi
    fi
else
    echo -e "${YELLOW}⚠${NC}  Project analysis script not found"
fi

echo ""

# ================================
# Step 4: Create harness configuration files
# ================================

echo -e "${BLUE}[4/6] Creating harness configuration files${NC}"
echo "----------------------------------------"

# Create .claude-code-harness directory
mkdir -p .claude-code-harness

# Create config file with references to existing documents (skip if already exists)
CONFIG_PATH=".claude-code-harness/config.json"
if [ -f "$CONFIG_PATH" ]; then
    echo -e "${YELLOW}⚠${NC}  Config file already exists (skipping): $CONFIG_PATH"
else
    cat > "$CONFIG_PATH" << EOF
{
  "version": "$HARNESS_VERSION",
  "setup_date": "$SETUP_DATE_ISO",
  "project_type": "existing",
  "existing_documents": [
$(
    for doc in "${FOUND_DOCS[@]}"; do
        echo "    \"$doc\","
    done | sed '$ s/,$//'
)
  ],
  "harness_path": "$HARNESS_ROOT"
}
EOF

    echo -e "${GREEN}✓${NC} Config file created: $CONFIG_PATH"
fi

# Create existing documents summary (skip if already exists)
if [ ${#FOUND_DOCS[@]} -gt 0 ]; then
    SUMMARY_PATH=".claude-code-harness/existing-docs-summary.md"
    if [ -f "$SUMMARY_PATH" ]; then
        echo -e "${YELLOW}⚠${NC}  Existing docs summary already exists (skipping): $SUMMARY_PATH"
    else
        cat > "$SUMMARY_PATH" << EOF
# Existing Documents

This project has the following existing documents:

EOF

        for doc in "${FOUND_DOCS[@]}"; do
            echo "## $doc" >> "$SUMMARY_PATH"
            echo "" >> "$SUMMARY_PATH"
            echo '```' >> "$SUMMARY_PATH"
            head -20 "$doc" >> "$SUMMARY_PATH"
            echo '```' >> "$SUMMARY_PATH"
            echo "" >> "$SUMMARY_PATH"
        done

        echo -e "${GREEN}✓${NC} Existing docs summary created: $SUMMARY_PATH"
    fi
fi

echo ""

# ================================
# Step 5: Create Project Rules
# ================================

echo -e "${BLUE}[5/6] Creating Project Rules / workflow files${NC}"
echo "----------------------------------------"

# Create .claude/rules directory
mkdir -p .claude/rules

# Simple template renderer ({{PROJECT_NAME}}/{{DATE}}/{{LANGUAGE}})
escape_sed_repl() {
    # Make safe as a sed replacement string (escape \, /, &, |)
    # Escape backslash first, then other characters
    printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e 's/[\/&|]/\\&/g'
}

render_template_if_missing() {
    local template_path="$1"
    local dest_path="$2"
    local label="$3"

    if [ -f "$dest_path" ]; then
        echo -e "${GREEN}✓${NC} ${label}: exists (skipped)"
        return 0
    fi
    if [ ! -f "$template_path" ]; then
        echo -e "${YELLOW}⚠${NC} ${label}: template not found: $template_path"
        return 0
    fi
    # Support nested paths
    mkdir -p "$(dirname "$dest_path")" 2>/dev/null || true

    local project_esc date_esc lang_esc
    project_esc=$(escape_sed_repl "$PROJECT_NAME")
    date_esc=$(escape_sed_repl "$SETUP_DATE_SHORT")
    lang_esc=$(escape_sed_repl "$LANGUAGE")

    sed \
        -e "s|{{PROJECT_NAME}}|$project_esc|g" \
        -e "s|{{DATE}}|$date_esc|g" \
        -e "s|{{LANGUAGE}}|$lang_esc|g" \
        "$template_path" > "$dest_path"

    echo -e "${GREEN}✓${NC} ${label} created: $dest_path"
}

# Create Project Rules for existing project (skip if already exists)
RULES_PATH=".claude/rules/harness.md"
if [ -f "$RULES_PATH" ]; then
    echo -e "${YELLOW}⚠${NC}  Project Rules already exist (skipping): $RULES_PATH"
else
    cat > "$RULES_PATH" << EOF
# Claude harness - Project Rules

This project uses **claude-code-harness**.

## Applying to an Existing Project

This project applies claude-code-harness to an existing codebase.

### Respect Existing Assets

1. **Prioritize existing documents**
   - When existing specs, README, or plans are present, refer to them first
   - A list of existing documents is available in \`.claude-code-harness/existing-docs-summary.md\`

2. **Maintain existing code style**
   - Respect existing coding conventions and formatting
   - New code should match the existing code style

3. **Incremental improvement**
   - Do not rewrite everything at once
   - Be careful not to break existing behavior

## Available Commands

### Core (Plan → Work → Review)
- \`/plan-with-agent\` - Create/update project plan (considering existing documents)
- \`/work\` - Implement features (parallel execution, maintain consistency with existing code)
- \`/harness-review\` - Code review

### Quality / Operations
- \`/validate\` - Pre-delivery validation
- \`/cleanup\` - Auto-organize Plans.md etc.
- \`/sync-status\` - Check progress → suggest next action
- \`/refactor\` - Safe refactoring

### Implementation Support
- \`/crud\` - Generate CRUD functionality
- \`/ci-setup\` - CI/CD configuration

### Skills (auto-triggered in conversation)
- \`component\` - "Create a hero" → UI component implementation
- \`auth\` - "Add login" → authentication implementation
- \`payments\` - "Stripe payments" → payment integration
- \`deploy-setup\` - "Deploy to Vercel" → deployment configuration
- \`analytics\` - "Add analytics" → analytics integration
- \`auto-fix\` - "Fix the issues" → auto-fix

## Notes for Existing Projects

1. **Always check existing specs first**
   - Read existing documents before running commands
   - Ask for confirmation if there are any contradictions

2. **Incremental application**
   - Start with small features
   - Verify behavior frequently

3. **Version management**
   - Commit often
   - Create a branch before large changes

## Setup Information

- Setup date: $SETUP_DATE_SHORT
- Harness version: $HARNESS_VERSION
- Config file: \`.claude-code-harness/config.json\`
EOF

    echo -e "${GREEN}✓${NC} Project Rules created: $RULES_PATH"
fi

echo ""

# Create workflow files (AGENTS/CLAUDE/Plans) as needed (skip if already exist)
TEMPLATE_DIR="$HARNESS_ROOT/templates"
render_template_if_missing "$TEMPLATE_DIR/AGENTS.md.template" "AGENTS.md" "AGENTS.md"
render_template_if_missing "$TEMPLATE_DIR/CLAUDE.md.template" "CLAUDE.md" "CLAUDE.md"
render_template_if_missing "$TEMPLATE_DIR/Plans.md.template" "Plans.md" "Plans.md"

echo ""

# ================================
# Step 5.5: Initialize project memory (SSOT)
# ================================
echo -e "${BLUE}[5.5/6] Initializing project memory (SSOT)${NC}"
echo "----------------------------------------"

# decisions/patterns are recommended as shared SSOT. session-log is for local use.
mkdir -p .claude/memory
render_template_if_missing "$TEMPLATE_DIR/memory/decisions.md.template" ".claude/memory/decisions.md" "decisions.md (SSOT)"
render_template_if_missing "$TEMPLATE_DIR/memory/patterns.md.template" ".claude/memory/patterns.md" "patterns.md (SSOT)"
render_template_if_missing "$TEMPLATE_DIR/memory/session-log.md.template" ".claude/memory/session-log.md" "session-log.md"

echo ""

# ================================
# Step 6: Setup complete
# ================================

echo -e "${BLUE}[6/6] Setup complete${NC}"
echo "----------------------------------------"

echo ""
echo -e "${GREEN}✅ Setup completed successfully!${NC}"
echo ""
echo "Next steps:"
echo ""
echo "1. Review existing documents:"
echo -e "   ${BLUE}cat .claude-code-harness/existing-docs-summary.md${NC}"
echo ""
echo "2. Open the project in Claude Code:"
echo -e "   ${BLUE}cd $PROJECT_PATH${NC}"
echo -e "   ${BLUE}claude${NC}"
echo -e "   ${YELLOW}(If loading the harness locally without installing the plugin)${NC}"
echo -e "   ${BLUE}claude --plugin-dir \"$HARNESS_ROOT\"${NC}"
echo ""
echo "3. Review existing specs and update the plan:"
echo -e "   ${BLUE}/plan${NC}"
echo ""
echo "4. Start implementing from small features:"
echo -e "   ${BLUE}/work${NC}"
echo ""
echo "5. Review frequently:"
echo -e "   ${BLUE}/harness-review${NC}"
echo ""
echo "6. (Optional) Enable Cursor integration:"
echo -e "   ${BLUE}/setup-cursor${NC}"
echo ""

# Add to .gitignore
if [ -f ".gitignore" ]; then
    if ! grep -q ".claude-code-harness" .gitignore; then
        echo "" >> .gitignore
        echo "# Claude harness" >> .gitignore
        echo ".claude-code-harness/" >> .gitignore
        echo -e "${GREEN}✓${NC} Added to .gitignore"
    fi

    # Recommend memory policy (no duplicate entries)
    if ! grep -q "Claude Memory Policy" .gitignore; then
        echo "" >> .gitignore
        echo "# Claude Memory Policy (recommended)" >> .gitignore
        echo "# - Keep (shared SSOT): .claude/memory/decisions.md, .claude/memory/patterns.md" >> .gitignore
        echo "# - Ignore (local): .claude/state/, session-log.md, context.json, archives" >> .gitignore
        echo ".claude/state/" >> .gitignore
        echo ".claude/memory/session-log.md" >> .gitignore
        echo ".claude/memory/context.json" >> .gitignore
        echo ".claude/memory/archive/" >> .gitignore
        echo -e "${GREEN}✓${NC} Added memory policy recommendation to .gitignore (adjust as needed)"
    fi
fi

echo ""
echo -e "${YELLOW}⚠${NC}  Important: It is recommended to commit the changes"
echo ""
