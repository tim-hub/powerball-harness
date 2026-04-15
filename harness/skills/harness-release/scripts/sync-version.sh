#!/bin/bash
# sync-version.sh - Sync VERSION / harness.toml release metadata
#
# Usage:
#   ./harness/skills/harness-release/scripts/sync-version.sh check    # Check for mismatches
#   ./harness/skills/harness-release/scripts/sync-version.sh sync     # Sync harness.toml to VERSION
#   ./harness/skills/harness-release/scripts/sync-version.sh bump     # Bump patch version for release
#
# Version sources (must stay in sync):
#   harness/VERSION      — canonical source of truth
#   harness/harness.toml — read by `harness sync` to regenerate plugin files
#
# Note: .claude-plugin/marketplace.json no longer carries a version field (v4.4.0+).
# Claude Code's plugin validator does not accept unknown keys in marketplace manifests.

set -euo pipefail

VERSION_FILE="harness/VERSION"
HARNESS_TOML="harness/harness.toml"

# Get current version
get_version() {
    cat "$VERSION_FILE" | tr -d '\n'
}

get_toml_version() {
    grep '^version' "$HARNESS_TOML" | sed 's/.*= "\([^"]*\)".*/\1/'
}

# Version mismatch check
check_version() {
    local v1=$(get_version)
    local ok=true

    if [ -f "$HARNESS_TOML" ]; then
        local v3=$(get_toml_version)
        if [ "$v1" != "$v3" ]; then
            echo "❌ Version mismatch:"
            echo "   VERSION:      $v1"
            echo "   harness.toml: $v3"
            ok=false
        fi
    fi

    if [ "$ok" = true ]; then
        echo "✅ Versions match: $v1"
        return 0
    fi
    return 1
}

# Sync harness.toml to VERSION
sync_version() {
    local version=$(get_version)

    if [ -f "$HARNESS_TOML" ]; then
        local toml_ver=$(get_toml_version)
        if [ "$version" != "$toml_ver" ]; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                sed -i '' "s/^version = \"$toml_ver\"/version = \"$version\"/" "$HARNESS_TOML"
            else
                sed -i "s/^version = \"$toml_ver\"/version = \"$version\"/" "$HARNESS_TOML"
            fi
            echo "✅ Updated harness.toml: $toml_ver → $version"
        fi
    fi

    echo "✅ Sync complete: $version"
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
