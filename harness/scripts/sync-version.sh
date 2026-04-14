#!/bin/bash
# sync-version.sh - Sync VERSION / plugin.json release metadata
#
# Usage:
#   ./scripts/sync-version.sh check    # Check for mismatches
#   ./scripts/sync-version.sh sync     # Sync plugin.json to VERSION
#   ./scripts/sync-version.sh bump     # Bump patch version for release

set -euo pipefail

VERSION_FILE="VERSION"
PLUGIN_JSON=".claude-plugin/marketplace.json"
HARNESS_TOML="harness.toml"

# Get current version
get_version() {
    cat "$VERSION_FILE" | tr -d '\n'
}

get_plugin_version() {
    grep '"version"' "$PLUGIN_JSON" | sed 's/.*"version": "\([^"]*\)".*/\1/'
}

get_toml_version() {
    grep '^version' "$HARNESS_TOML" | sed 's/.*= "\([^"]*\)".*/\1/'
}

# Version mismatch check
check_version() {
    local v1=$(get_version)
    local v2=$(get_plugin_version)
    local v3=""
    if [ -f "$HARNESS_TOML" ]; then
        v3=$(get_toml_version)
    fi

    local ok=true
    if [ "$v1" != "$v2" ]; then
        echo "❌ Version mismatch:"
        echo "   VERSION:      $v1"
        echo "   plugin.json:  $v2"
        ok=false
    fi
    if [ -n "$v3" ] && [ "$v1" != "$v3" ]; then
        echo "❌ Version mismatch:"
        echo "   VERSION:      $v1"
        echo "   harness.toml: $v3"
        ok=false
    fi

    if [ "$ok" = true ]; then
        echo "✅ Versions match: $v1"
        return 0
    fi
    return 1
}

# Sync plugin.json + harness.toml to VERSION
sync_version() {
    local version=$(get_version)
    local current=$(get_plugin_version)

    if [ "$version" != "$current" ]; then
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/\"version\": \"$current\"/\"version\": \"$version\"/" "$PLUGIN_JSON"
        else
            sed -i "s/\"version\": \"$current\"/\"version\": \"$version\"/" "$PLUGIN_JSON"
        fi
        echo "✅ Updated plugin marketplace.json: $current → $version"
    fi

    # Sync harness.toml
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
