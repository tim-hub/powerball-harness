#!/bin/bash
#
# setup-opencode.sh
#
# Set up Harness for opencode.ai without Claude Code
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/Chachamaru127/claude-code-harness/main/scripts/setup-opencode.sh | bash
#
# Or:
#   ./setup-opencode.sh
#

set -e

# Colored output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logo
echo -e "${BLUE}"
echo '  ___ _                 _        _  _                              '
echo ' / __| |__ _ _  _ _____|_)___   | || |__ _ _ _ _ _  ___ ________ '
echo '| (__| / _` | || / _` / / -_)  | __ / _` | `_| ` \/ -_|_-<_-<_-<'
echo ' \___|_\__,_|\_,_\__,_/_\___|  |_||_\__,_|_| |_||_\___/__/__/__/'
echo ''
echo '                    for opencode.ai'
echo -e "${NC}"

# Variables
HARNESS_REPO="https://github.com/Chachamaru127/claude-code-harness.git"
HARNESS_BRANCH="main"
TEMP_DIR=$(mktemp -d)
PROJECT_DIR=$(pwd)

# Cleanup function
cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# Function: error display
error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

# Function: success display
success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function: info display
info() {
    echo -e "${BLUE}→ $1${NC}"
}

# Function: warning display
warn() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Prerequisites check
check_requirements() {
    info "Checking requirements..."

    if ! command -v git &> /dev/null; then
        error "git is required but not installed"
    fi

    success "All requirements met"
}

# Clone Harness
clone_harness() {
    info "Downloading Harness..."

    git clone --depth 1 --branch "$HARNESS_BRANCH" "$HARNESS_REPO" "$TEMP_DIR/harness" 2>/dev/null || \
        error "Failed to clone Harness repository"

    success "Harness downloaded"
}

# Copy opencode directory
copy_opencode_files() {
    info "Setting up opencode files..."

    # Create .opencode/commands/
    mkdir -p "$PROJECT_DIR/.opencode/commands"

    # Copy commands
    if [ -d "$TEMP_DIR/harness/opencode/commands" ]; then
        cp -r "$TEMP_DIR/harness/opencode/commands/"* "$PROJECT_DIR/.opencode/commands/"
        success "Commands copied to .opencode/commands/"
    else
        error "opencode/commands not found in Harness"
    fi

    # Create .claude/skills/ and copy skills
    mkdir -p "$PROJECT_DIR/.claude/skills"

    if [ -d "$PROJECT_DIR/.claude/skills" ] && [ "$(ls -A "$PROJECT_DIR/.claude/skills" 2>/dev/null)" ]; then
        warn ".claude/skills/ already has content, creating backup"
        mv "$PROJECT_DIR/.claude/skills" "$PROJECT_DIR/.claude/skills.backup.$(date +%Y%m%d%H%M%S)"
        mkdir -p "$PROJECT_DIR/.claude/skills"
    fi

    if [ -d "$TEMP_DIR/harness/opencode/skills" ]; then
        cp -r "$TEMP_DIR/harness/opencode/skills/"* "$PROJECT_DIR/.claude/skills/"
        success "Skills copied to .claude/skills/"
    else
        warn "opencode/skills not found in Harness (optional)"
    fi

    # Copy AGENTS.md (backup if existing)
    if [ -f "$PROJECT_DIR/AGENTS.md" ]; then
        warn "AGENTS.md already exists, creating backup"
        mv "$PROJECT_DIR/AGENTS.md" "$PROJECT_DIR/AGENTS.md.backup.$(date +%Y%m%d%H%M%S)"
    fi

    if [ -f "$TEMP_DIR/harness/opencode/AGENTS.md" ]; then
        cp "$TEMP_DIR/harness/opencode/AGENTS.md" "$PROJECT_DIR/AGENTS.md"
        success "AGENTS.md created (from CLAUDE.md)"
    fi
}

# Generate opencode.json (optional)
setup_mcp() {
    echo ""
    echo -e "${YELLOW}Do you want to setup MCP server? (for advanced workflow tools)${NC}"
    echo "This requires Node.js and allows using harness_workflow_* tools"
    read -p "Setup MCP? (y/N): " setup_mcp_answer

    if [[ "$setup_mcp_answer" =~ ^[Yy]$ ]]; then
        if [ -f "$PROJECT_DIR/opencode.json" ]; then
            warn "opencode.json already exists, skipping"
            return
        fi

        # Copy opencode.json
        if [ -f "$TEMP_DIR/harness/opencode/opencode.json" ]; then
            cp "$TEMP_DIR/harness/opencode/opencode.json" "$PROJECT_DIR/opencode.json"
            success "opencode.json created"

            warn "You need to:"
            echo "  1. Clone Harness: git clone $HARNESS_REPO"
            echo "  2. Build MCP server: cd claude-code-harness/mcp-server && npm install && npm run build"
            echo "  3. Update path in opencode.json"
        fi
    fi
}

# Done message
print_success() {
    echo ""
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  ✅ Harness for OpenCode setup complete!${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "Created files:"
    echo "  📁 .opencode/commands/  - Harness commands"
    echo "  📁 .claude/skills/      - Harness skills (notebookLM, harness-review, impl, etc.)"
    echo "  📄 AGENTS.md            - Rules file (from CLAUDE.md)"
    [ -f "$PROJECT_DIR/opencode.json" ] && echo "  📄 opencode.json        - MCP configuration"
    echo ""
    echo "Available skills:"
    echo "  • notebookLM - Documentation (NotebookLM, slides)"
    echo "  • impl    - Feature implementation"
    echo "  • harness-review - Code review"
    echo "  • verify  - Build verification"
    echo "  • auth    - Authentication (Clerk, Stripe)"
    echo "  • deploy  - Deployment (Vercel, Netlify)"
    echo ""
    echo "Next steps:"
    echo "  1. Start opencode: ${BLUE}opencode${NC}"
    echo "  2. Run commands:   ${BLUE}/plan-with-agent${NC}, ${BLUE}/work${NC}, ${BLUE}/harness-review${NC}"
    echo ""
    echo "Documentation: https://github.com/Chachamaru127/claude-code-harness"
    echo ""
}

# Main processing
main() {
    echo ""
    info "Setting up Harness for OpenCode in: $PROJECT_DIR"
    echo ""

    check_requirements
    clone_harness
    copy_opencode_files
    setup_mcp
    print_success
}

main "$@"
