#!/bin/bash
# sync-version.sh - Sync release metadata VERSION / plugin.json
#
# Usage:
#   ./scripts/sync-version.sh check    # Check for version mismatch
#   ./scripts/sync-version.sh sync     # Sync plugin.json to match VERSION
#   ./scripts/sync-version.sh bump     # Bump patch version for release

set -euo pipefail

VERSION_FILE="VERSION"
PLUGIN_JSON=".claude-plugin/plugin.json"

# Get current version
get_version() {
    cat "$VERSION_FILE" | tr -d '\n'
}

get_plugin_version() {
    grep '"version"' "$PLUGIN_JSON" | sed 's/.*"version": "\([^"]*\)".*/\1/'
}

# Check for version mismatch
check_version() {
    local v1=$(get_version)
    local v2=$(get_plugin_version)

    if [ "$v1" != "$v2" ]; then
        echo "❌ Version mismatch:"
        echo "   VERSION:     $v1"
        echo "   plugin.json: $v2"
        return 1
    else
        echo "✅ Version match: $v1"
        return 0
    fi
}

# Sync plugin.json to VERSION
sync_version() {
    local version=$(get_version)
    local current=$(get_plugin_version)

    if [ "$version" = "$current" ]; then
        echo "✅ Already in sync: $version"
        return 0
    fi

    # Compatible with both macOS and Linux
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s/\"version\": \"$current\"/\"version\": \"$version\"/" "$PLUGIN_JSON"
    else
        sed -i "s/\"version\": \"$current\"/\"version\": \"$version\"/" "$PLUGIN_JSON"
    fi

    echo "✅ Updated plugin.json: $current → $version"
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
