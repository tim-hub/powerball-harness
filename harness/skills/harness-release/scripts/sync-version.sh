#!/bin/bash
# sync-version.sh - Sync VERSION to all release metadata files
#
# Usage:
#   ./harness/skills/harness-release/scripts/sync-version.sh check    # Check for mismatches
#   ./harness/skills/harness-release/scripts/sync-version.sh sync     # Sync all files to VERSION
#   ./harness/skills/harness-release/scripts/sync-version.sh bump     # Bump patch version for release
#
# Version sources (must stay in sync):
#   harness/VERSION      — canonical source of truth
#   harness/harness.toml — read by `harness sync`
#
# Note: .claude-plugin/marketplace.json no longer carries a version field (v4.4.0+).
# Claude Code's plugin validator does not accept unknown keys in marketplace manifests.
# Templates (harness/templates/) are NOT synced here — they are scaffolded once into
# user projects and must remain backward compatible across plugin version bumps.

set -euo pipefail

VERSION_FILE="harness/VERSION"       # project-root
HARNESS_TOML="harness/harness.toml"  # project-root

# Get current canonical version
get_version() {
    cat "$VERSION_FILE" | tr -d '\n'
}

get_toml_version() {
    grep '^version' "$HARNESS_TOML" | sed 's/.*= "\([^"]*\)".*/\1/'
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
