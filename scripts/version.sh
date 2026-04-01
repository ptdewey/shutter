#!/usr/bin/env bash
set -euo pipefail

# Determine next semver based on conventional commits since main.
# Usage: ./scripts/version.sh [--dry-run]

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
    DRY_RUN=true
fi

# Get the latest root module tag (ignore cmd/shutter/ prefixed tags)
LATEST_TAG=$(jj tag list | grep -E '^v[0-9]' | sort -V -t: -k1,1 | tail -1 | awk '{print $1}' | tr -d ':')

if [[ -z "$LATEST_TAG" ]]; then
    echo "error: no existing version tags found"
    exit 1
fi

echo "Current version: $LATEST_TAG"

# Parse current version
VERSION="${LATEST_TAG#v}"
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

# Get commit messages since main
COMMITS=$(jj log -r 'main..@' --no-graph -T 'description ++ "---\n"' 2>/dev/null)

if [[ -z "$COMMITS" || "$COMMITS" == $'---\n' ]]; then
    echo "No commits since main."
    exit 0
fi

echo ""
echo "Commits since main:"
echo "$COMMITS" | grep -v '^---$' | grep -v '^$' | sed 's/^/  /'
echo ""

# Determine bump type from conventional commits
BUMP="patch"

while IFS= read -r line; do
    # Skip empty lines and delimiters
    [[ -z "$line" || "$line" == "---" ]] && continue

    # Check for breaking changes
    if echo "$line" | grep -qiE '^[a-z]+(\(.+\))?!:|BREAKING CHANGE'; then
        BUMP="major"
        break
    fi

    # Check for feat -> minor
    if echo "$line" | grep -qE '^feat(\(.+\))?:'; then
        BUMP="minor"
    fi
done <<< "$COMMITS"

# Calculate new version
case "$BUMP" in
    major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
    minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
    patch) PATCH=$((PATCH + 1)) ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

echo "Bump type: $BUMP"
echo "New version: $NEW_VERSION"
echo "Tags: $NEW_VERSION, cmd/shutter/$NEW_VERSION"

if $DRY_RUN; then
    echo ""
    echo "(dry run — no tags created)"
    exit 0
fi

echo ""
read -p "Create tags and push? [y/N] " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    jj tag set "$NEW_VERSION" "cmd/shutter/$NEW_VERSION"
    jj git push --tags
    echo "Done. Tagged and pushed $NEW_VERSION"
else
    echo "Aborted."
fi
