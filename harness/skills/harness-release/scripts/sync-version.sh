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
#   (optional) extra manifest files configured via HARNESS_RELEASE_EXTRA_VERSION_FILES
#
# For this plugin, harness/harness.toml carries the version and is synced by setting:
#   HARNESS_RELEASE_EXTRA_VERSION_FILES="harness/harness.toml"
#
# Generic projects leave HARNESS_RELEASE_EXTRA_VERSION_FILES unset (or set it to their own manifests).
#
# Note: .claude-plugin/marketplace.json no longer carries a version field (v4.4.0+).
# Claude Code's plugin validator does not accept unknown keys in marketplace manifests.
# Templates (harness/templates/) are NOT synced here — they are scaffolded once into
# user projects and must remain backward compatible across plugin version bumps.

set -euo pipefail

VERSION_FILE="harness/VERSION"       # project-root

# Optional: space-separated list of extra files to sync VERSION into.
# For this plugin: HARNESS_RELEASE_EXTRA_VERSION_FILES="harness/harness.toml"
# Generic projects: leave unset or set to their own manifest files.
EXTRA_VERSION_FILES="${HARNESS_RELEASE_EXTRA_VERSION_FILES:-}"

# Get current canonical version
get_version() {
    cat "$VERSION_FILE" | tr -d '\n'
}

# Get version from a TOML-style file (accepts a file path argument)
get_toml_version() {
    local file="${1:-harness/harness.toml}"  # default for backward compat
    grep '^version' "$file" | head -1 | sed 's/.*= "\([^"]*\)".*/\1/'
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
# Checks VERSION against any files listed in HARNESS_RELEASE_EXTRA_VERSION_FILES.
check_version() {
    local v
    v="$(get_version)"

    if [ -z "${EXTRA_VERSION_FILES:-}" ]; then
        echo "✅ Version: $v (no extra manifest files configured)"
        return 0
    fi

    local mismatch=0
    for file in $EXTRA_VERSION_FILES; do
        if [ ! -f "$file" ]; then
            echo "⚠️  Warning: configured file '$file' not found"
            mismatch=1
            continue
        fi
        local file_version
        file_version="$(get_toml_version "$file" 2>/dev/null || echo 'UNKNOWN')"
        if [ "$file_version" != "$v" ]; then
            echo "❌ MISMATCH: VERSION=$v but $file version=$file_version"
            mismatch=1
        else
            echo "✅ OK: VERSION=$v matches $file"
        fi
    done

    return $mismatch
}

# sync: Apply VERSION to CHANGELOG compare links + any HARNESS_RELEASE_EXTRA_VERSION_FILES.
# Example for this plugin:
#   HARNESS_RELEASE_EXTRA_VERSION_FILES="harness/harness.toml" bash sync-version.sh sync
sync_version() {
    local version
    version="$(get_version)"
    local updated=false

    # Sync to extra manifest files (only if configured)
    if [ -n "${EXTRA_VERSION_FILES:-}" ]; then
        for file in $EXTRA_VERSION_FILES; do
            if [ ! -f "$file" ]; then
                echo "⚠️  Warning: EXTRA_VERSION_FILES includes '$file' but file not found — skipping"
                continue
            fi
            local file_version
            file_version="$(get_toml_version "$file" 2>/dev/null || echo 'UNKNOWN')"
            if [ "$version" != "$file_version" ]; then
                do_sed "s/^version = \"$file_version\"/version = \"$version\"/" "$file"
                echo "✅ Updated $file: $file_version → $version"
                updated=true
            fi
        done
    fi

    if [ "$updated" = false ]; then
        echo "✅ All files already at $version (no changes)"
    else
        echo "✅ Sync complete: $version"
    fi
}

# Update CHANGELOG.md compare links (replace Unreleased link + insert new version line)
update_changelog_compare_links() {
    local current="$1"
    local new="$2"
    local changelog
    changelog="$(git rev-parse --show-toplevel)/CHANGELOG.md"  # project-root

    if [ ! -f "$changelog" ]; then
        return 0
    fi

    python3 - "$changelog" "$current" "$new" <<'PY'
import re
import sys

changelog, current, new = sys.argv[1], sys.argv[2], sys.argv[3]
with open(changelog, "r", encoding="utf-8") as fh:
    lines = fh.readlines()

pattern = re.compile(
    rf"^\[Unreleased\]: (https://github\.com/[^/]+/[^/]+)/compare/v{re.escape(current)}\.\.\.HEAD\s*$"
)

new_lines = []
inserted = False
for line in lines:
    match = pattern.match(line)
    if match and not inserted:
        repo = match.group(1)
        new_lines.append(f"[Unreleased]: {repo}/compare/v{new}...HEAD\n")
        new_lines.append(f"[{new}]: {repo}/compare/v{current}...v{new}\n")
        inserted = True
        continue
    new_lines.append(line)

if not inserted:
    print(
        f"  CHANGELOG.md [Unreleased] compare link (v{current}...HEAD) not found. Add manually.",
        file=sys.stderr,
    )
    sys.exit(0)

with open(changelog, "w", encoding="utf-8") as fh:
    fh.writelines(new_lines)

print(f"  Updated CHANGELOG.md compare link: [{new}]")
PY
}

# Bump version (default: patch)
bump_version() {
    local level="${1:-patch}"
    local current
    current="$(get_version)"
    local major
    major="$(echo "$current" | cut -d. -f1)"
    local minor
    minor="$(echo "$current" | cut -d. -f2)"
    local patch
    patch="$(echo "$current" | cut -d. -f3)"

    local new_version=""
    case "$level" in
        patch)
            local new_patch=$((patch + 1))
            new_version="$major.$minor.$new_patch"
            ;;
        minor)
            local new_minor=$((minor + 1))
            new_version="$major.$new_minor.0"
            ;;
        major)
            local new_major=$((major + 1))
            new_version="$new_major.0.0"
            ;;
        *)
            echo "Unsupported bump level: $level" >&2
            echo "  Available: patch | minor | major" >&2
            exit 1
            ;;
    esac

    echo "$new_version" > "$VERSION_FILE"
    echo "✅ Updated VERSION ($level): $current → $new_version"

    sync_version
    update_changelog_compare_links "$current" "$new_version"
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
        bump_version "${2:-patch}"
        ;;
    *)
        echo "Usage: $0 {check|sync|bump [patch|minor|major]}"
        exit 1
        ;;
esac
