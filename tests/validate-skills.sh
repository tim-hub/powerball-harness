#!/bin/bash
# validate-skills.sh
# Skills integrity and governance validation test
#
# Usage: ./tests/validate-skills.sh [--verbose]
#
# Validation items:
#   1. SKILL.md frontmatter required fields (description, allowed-tools)
#   2. *.md files exist in references/ directory
#   3. allowed-tools are valid Claude Code tool names
#   4. dependencies reference existing skills

set -u
set -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_ROOT="$(dirname "$SCRIPT_DIR")"
SKILLS_DIR="$PLUGIN_ROOT/skills"

VERBOSE=0
if [[ "${1:-}" == "--verbose" ]]; then
  VERBOSE=1
fi

# Color output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

pass_test() {
  echo -e "${GREEN}✓${NC} $1"
  PASS_COUNT=$((PASS_COUNT + 1))
}

fail_test() {
  echo -e "${RED}✗${NC} $1"
  FAIL_COUNT=$((FAIL_COUNT + 1))
}

warn_test() {
  echo -e "${YELLOW}⚠${NC} $1"
  WARN_COUNT=$((WARN_COUNT + 1))
}

debug_log() {
  if [ "$VERBOSE" -eq 1 ]; then
    echo "  [DEBUG] $1"
  fi
}

# Valid Claude Code tool name list
VALID_TOOLS=(
  "Read" "Write" "Edit" "Glob" "Grep" "Bash"
  "Task" "WebFetch" "WebSearch" "TodoWrite"
  "AskUserQuestion" "Skill" "EnterPlanMode" "ExitPlanMode"
  "NotebookEdit" "LSP" "MCPSearch" "Append"
)

is_valid_tool() {
  local tool="$1"
  for valid in "${VALID_TOOLS[@]}"; do
    if [[ "$valid" == "$tool" ]]; then
      return 0
    fi
  done
  return 1
}

# Extract field value from frontmatter
extract_frontmatter_field() {
  local file="$1"
  local field="$2"

  awk -v field="$field" '
    NR==1 && $0!="---" { exit 1 }
    NR>1 && $0=="---" { exit 0 }
    $0 ~ "^"field":" {
      sub("^"field": *", "")
      gsub(/^["'\'']|["'\'']$/, "")
      print
      exit 0
    }
  ' "$file"
}

echo "=========================================="
echo "Claude harness - Skills Validation Test"
echo "=========================================="
echo ""

if [ ! -d "$SKILLS_DIR" ]; then
  fail_test "skills directory not found: $SKILLS_DIR"
  exit 1
fi

# Collect skill directories
SKILL_DIRS=()
while IFS= read -r skill_md; do
  SKILL_DIRS+=("$(dirname "$skill_md")")
done < <(find "$SKILLS_DIR" -name "SKILL.md" -type f 2>/dev/null | sort)

if [ ${#SKILL_DIRS[@]} -eq 0 ]; then
  warn_test "No SKILL.md files found"
  exit 0
fi

echo "1. SKILL.md frontmatter validation"
echo "----------------------------------------"

for skill_dir in "${SKILL_DIRS[@]}"; do
  skill_name=$(basename "$skill_dir")
  skill_file="$skill_dir/SKILL.md"

  debug_log "Checking: $skill_name"

  # description required
  description=$(extract_frontmatter_field "$skill_file" "description")
  if [ -n "$description" ]; then
    pass_test "[$skill_name] description: ${description:0:50}..."
  else
    fail_test "[$skill_name] description not found"
  fi

  # allowed-tools required
  allowed_tools=$(extract_frontmatter_field "$skill_file" "allowed-tools")
  if [ -n "$allowed_tools" ]; then
    pass_test "[$skill_name] allowed-tools: $allowed_tools"
  else
    fail_test "[$skill_name] allowed-tools not found"
  fi
done

echo ""
echo "2. allowed-tools validity check"
echo "----------------------------------------"

for skill_dir in "${SKILL_DIRS[@]}"; do
  skill_name=$(basename "$skill_dir")
  skill_file="$skill_dir/SKILL.md"

  allowed_tools=$(extract_frontmatter_field "$skill_file" "allowed-tools")
  if [ -z "$allowed_tools" ]; then
    continue
  fi

  # Parse [Tool1, Tool2] or ["Tool1", "Tool2"] format
  # Remove quotes, brackets, and spaces
  tools_str=$(echo "$allowed_tools" | sed 's/^\[//' | sed 's/\]$//' | tr ',' '\n' | sed 's/^[ "]*//;s/[ "]*$//')

  invalid_found=0
  while IFS= read -r tool; do
    # Remove extra whitespace and quotes
    tool=$(echo "$tool" | tr -d ' "'\''')
    if [ -z "$tool" ]; then
      continue
    fi

    # Skip wildcard patterns (mcp__*)
    if [[ "$tool" == *"*"* ]]; then
      debug_log "[$skill_name] Wildcard pattern skipped: $tool"
      continue
    fi

    if is_valid_tool "$tool"; then
      debug_log "[$skill_name] Valid tool: $tool"
    else
      fail_test "[$skill_name] invalid tool name: $tool"
      invalid_found=1
    fi
  done <<< "$tools_str"

  if [ "$invalid_found" -eq 0 ]; then
    pass_test "[$skill_name] all tool names valid"
  fi
done

echo ""
echo "3. references/ directory validation"
echo "----------------------------------------"

for skill_dir in "${SKILL_DIRS[@]}"; do
  skill_name=$(basename "$skill_dir")
  ref_dir="$skill_dir/references"

  if [ -d "$ref_dir" ]; then
    ref_count=$(find "$ref_dir" -name "*.md" -type f | wc -l | tr -d ' ')
    if [ "$ref_count" -gt 0 ]; then
      pass_test "[$skill_name] references/: $ref_count documents"
    else
      warn_test "[$skill_name] references/ is empty"
    fi
  else
    debug_log "[$skill_name] no references/ (optional)"
  fi
done

echo ""
echo "4. Dependencies validation"
echo "----------------------------------------"

# Collect all skill names
ALL_SKILL_NAMES=()
for skill_dir in "${SKILL_DIRS[@]}"; do
  ALL_SKILL_NAMES+=("$(basename "$skill_dir")")
done

for skill_dir in "${SKILL_DIRS[@]}"; do
  skill_name=$(basename "$skill_dir")
  skill_file="$skill_dir/SKILL.md"

  dependencies=$(extract_frontmatter_field "$skill_file" "dependencies")
  if [ -z "$dependencies" ] || [ "$dependencies" == "[]" ]; then
    debug_log "[$skill_name] no dependencies"
    continue
  fi

  # Parse [dep1, dep2] format
  deps_str=$(echo "$dependencies" | sed 's/^\[//' | sed 's/\]$//' | tr ',' '\n')

  invalid_dep=0
  while IFS= read -r dep; do
    dep=$(echo "$dep" | tr -d ' ')
    if [ -z "$dep" ]; then
      continue
    fi

    found=0
    for existing in "${ALL_SKILL_NAMES[@]}"; do
      if [ "$existing" == "$dep" ]; then
        found=1
        break
      fi
    done

    if [ "$found" -eq 1 ]; then
      pass_test "[$skill_name] dependency '$dep' exists"
    else
      fail_test "[$skill_name] dependency '$dep' not found"
      invalid_dep=1
    fi
  done <<< "$deps_str"
done

echo ""
echo "=========================================="
echo "Skills Validation Results Summary"
echo "=========================================="
echo -e "${GREEN}Passed:${NC} $PASS_COUNT"
echo -e "${YELLOW}Warnings:${NC} $WARN_COUNT"
echo -e "${RED}Failed:${NC} $FAIL_COUNT"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
  echo -e "${GREEN}✓ All skill validations passed!${NC}"
  exit 0
else
  echo -e "${RED}✗ $FAIL_COUNT validation(s) failed${NC}"
  exit 1
fi
