#!/bin/bash

# Check if VERSION file exists
if [ ! -f VERSION ]; then
    echo "Error: VERSION file not found."
    exit 1
fi

# Check for uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
    echo "Error: Working directory is not clean. Please commit or stash changes first."
    exit 1
fi

CURRENT_VERSION=$(cat VERSION | tr -d '[:space:]')

# Split version into array based on . delimiter
IFS='.' read -r -a parts <<< "$CURRENT_VERSION"

MAJOR=${parts[0]}
MINOR=${parts[1]}
PATCH=${parts[2]}

case "$1" in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "Usage: $0 {major|minor|patch}"
        exit 1
        ;;
esac

NEW_VERSION="$MAJOR.$MINOR.$PATCH"
echo "$NEW_VERSION" > VERSION
echo "Bumped version from $CURRENT_VERSION to $NEW_VERSION"

# Git operations
git add VERSION
git commit -m "Bump version to $NEW_VERSION"
git tag "v$NEW_VERSION"
echo "Pushing changes and tag..."
git push
git push origin "v$NEW_VERSION"