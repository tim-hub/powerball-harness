#!/bin/bash
# sync-version.sh - Sync VERSION to all release metadata files
#
# Usage:
#   ./harness/skills/harness-release/scripts/sync-version.sh check    # Check for mismatches
#   ./harness/skills/harness-release/scripts/sync-version.sh sync     # Sync all files to VERSION
#   ./harness/skills/harness-release/scripts/sync-version.sh bump     # Bump patch version for release
#
# Version sources (must stay in sync):
#   harness/VERSION                          — canonical source of truth
#   harness/harness.toml                     — read by `harness sync`
#   harness/templates/template-registry.json — templateVersion fields
#   harness/templates/CLAUDE.md.template     — _harness_version field
#   harness/templates/AGENTS.md.template     — _harness_version field
#   harness/templates/Plans.md.template      — _harness_version field
#
# Note: .claude-plugin/marketplace.json no longer carries a version field (v4.4.0+).
# Claude Code's plugin validator does not accept unknown keys in marketplace manifests.

set -euo pipefail

VERSION_FILE="harness/VERSION"                                   # project-root
HARNESS_TOML="harness/harness.toml"                             # project-root
TEMPLATE_REGISTRY="harness/templates/template-registry.json"    # project-root

# Get current canonical version
get_version() {
    cat "$VERSION_FILE" | tr -d '\n'
}

get_toml_version() {
    grep '^version' "$HARNESS_TOML" | sed 's/.*= "\([^"]*\)".*/\1/'
}

get_registry_version() {
    grep '"templateVersion"' "$TEMPLATE_REGISTRY" | head -1 | sed 's/.*"\([0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\)".*/\1/'
}

get_md_template_version() {
    local file="$1"
    grep '_harness_version:' "$file" | sed 's/.*"\([^"]*\)".*/\1/'
}

# Cross-platform sed in-place
do_sed() {
    local pattern="$1"
    local file="$2"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "$pattern" "$file"
    else
        sed -i "$pattern" "$file"
    fi
}

# Version mismatch check
check_version() {
    local v=$(get_version)
    local ok=true

    if [ -f "$HARNESS_TOML" ]; then
        local toml_ver=$(get_toml_version)
        if [ "$v" != "$toml_ver" ]; then
            echo "❌ FAIL: harness.toml: version mismatch ($toml_ver != $v)"
            ok=false
        fi
    fi

    if [ -f "$TEMPLATE_REGISTRY" ]; then
        local reg_ver=$(get_registry_version)
        if [ "$v" != "$reg_ver" ]; then
            echo "❌ FAIL: template-registry.json: version mismatch ($reg_ver != $v)"
            ok=false
        fi
    fi

    while IFS= read -r tmpl; do
        local tmpl_ver=$(get_md_template_version "$tmpl")
        local tmpl_name="${tmpl#harness/templates/}"
        if [ "$v" != "$tmpl_ver" ]; then
            echo "❌ FAIL: $tmpl_name: version mismatch ($tmpl_ver != $v)"
            ok=false
        fi
    done < <(grep -rl '_harness_version:' harness/templates/ 2>/dev/null | sort)

    if [ "$ok" = true ]; then
        echo "✅ Versions match: $v"
        return 0
    fi
    return 1
}

# Sync all metadata files to VERSION
sync_version() {
    local version=$(get_version)
    local updated=false

    # harness.toml
    if [ -f "$HARNESS_TOML" ]; then
        local toml_ver=$(get_toml_version)
        if [ "$version" != "$toml_ver" ]; then
            do_sed "s/^version = \"$toml_ver\"/version = \"$version\"/" "$HARNESS_TOML"
            echo "✅ Updated harness.toml: $toml_ver → $version"
            updated=true
        fi
    fi

    # template-registry.json — all templateVersion fields
    if [ -f "$TEMPLATE_REGISTRY" ]; then
        local reg_ver=$(get_registry_version)
        if [ "$version" != "$reg_ver" ]; then
            do_sed "s/\"templateVersion\": \"$reg_ver\"/\"templateVersion\": \"$version\"/g" "$TEMPLATE_REGISTRY"
            echo "✅ Updated template-registry.json: $reg_ver → $version"
            updated=true
        fi
    fi

    # *.md.template files — _harness_version field (glob all tracked templates)
    while IFS= read -r tmpl; do
        local tmpl_ver=$(get_md_template_version "$tmpl")
        local tmpl_name="${tmpl#harness/templates/}"
        if [ "$version" != "$tmpl_ver" ]; then
            do_sed "s/_harness_version: \"$tmpl_ver\"/_harness_version: \"$version\"/" "$tmpl"
            echo "✅ Updated $tmpl_name: $tmpl_ver → $version"
            updated=true
        fi
    done < <(grep -rl '_harness_version:' harness/templates/ 2>/dev/null | sort)

    if [ "$updated" = false ]; then
        echo "✅ All files already at $version (no changes)"
    else
        echo "✅ Sync complete: $version"
    fi
}

# Bump patch version
bump_version() {
    local current=$(get_version)
    local major=$(echo "$current" | cut -d. -f1)
    local minor=$(echo "$current" | cut -d. -f2)
    local patch=$(echo "$current" | cut -d. -f3)

    local new_patch=$((patch + 1))
    local new_version="$major.$minor.$new_patch"

    echo "$new_version" > "$VERSION_FILE"
    echo "✅ Updated VERSION: $current → $new_version"

    sync_version
}

# Main
case "${1:-check}" in
    check)
        check_version
        ;;
    sync)
        sync_version
        ;;
    bump)
        bump_version
        ;;
    *)
        echo "Usage: $0 {check|sync|bump}"
        exit 1
        ;;
esac
