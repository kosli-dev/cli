#!/usr/bin/env bash
# Interactive step after suggest-version-ai: show version and release notes,
# let user edit notes, then confirm before creating tag and pushing.
# Called from Make when running `make release` (no tag).
# Requires: dist/suggested_version and dist/release_notes.md exist.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

SUGGESTED_VERSION_FILE="dist/suggested_version"
RELEASE_NOTES_FILE="dist/release_notes.md"

if [ ! -f "$SUGGESTED_VERSION_FILE" ] || [ ! -f "$RELEASE_NOTES_FILE" ]; then
  echo "Missing $SUGGESTED_VERSION_FILE or $RELEASE_NOTES_FILE. Run suggest-version-ai first." >&2
  exit 1
fi

VER=$(cat "$SUGGESTED_VERSION_FILE")
if [ -z "$VER" ]; then
  echo "Suggested version is empty. Run suggest-version-ai or use: make release tag=vX.Y.Z" >&2
  exit 1
fi

echo "Suggested tag: $VER"
echo ""
echo "Release notes ($RELEASE_NOTES_FILE):"
echo "---"
cat "$RELEASE_NOTES_FILE"
echo "---"
echo ""

# Let user edit release notes
read -r -p "Edit release notes? [y/N] " edit_notes
case "$edit_notes" in
  y|Y) "${EDITOR:-vi}" "$RELEASE_NOTES_FILE" ;;
  *) ;;
esac

echo ""
read -r -p "Create tag $VER and push? [y/N] " confirm
case "$confirm" in
  y|Y) ;;
  *) echo "Aborted. To release later run: make release tag=$VER"; exit 0 ;;
esac

git remote update
if ! git status -uno | grep -q "Your branch is up to date"; then
  echo "ERROR: your branch is NOT up to date with remote" >&2
  exit 1
fi

git tag -a "$VER" -F "$RELEASE_NOTES_FILE"
git push origin "$VER"
echo "Pushed tag $VER. Release workflow will run on GitHub."
# # Clean up temporary files to avoid reusing them in subsequent releases
# rm "$RELEASE_NOTES_FILE" || true
# rm "$SUGGESTED_VERSION_FILE" || true
