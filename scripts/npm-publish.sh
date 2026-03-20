#!/bin/bash
set -e

VERSION="$1"
if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

# When called from goreleaser, $2 is "true" if snapshot build
DRY_RUN=false
if [ "$2" == "true" ] || [ "$2" == "--dry-run" ]; then
  echo "Running in DRY-RUN mode. Packages will be created but not published."
  DRY_RUN=true
fi

# Regex for stable: X.Y.Z (where X, Y, Z are numbers)
STABLE_REGEX="^[0-9]+\.[0-9]+\.[0-9]+$"

# Regex for pre-release: X.Y.Z-TAG (where TAG starts with a hyphen)
PRE_REGEX="^[0-9]+\.[0-9]+\.[0-9]+-.*$"

# Determine npm dist-tag: pre-release versions must not go to "latest"
if [[ $VERSION =~ $STABLE_REGEX ]]; then
    echo "✅ '$VERSION' is a STABLE release."
    NPM_TAG="latest"
elif [[ $VERSION =~ $PRE_REGEX ]]; then
    echo "🧪 '$VERSION' is a PRE-RELEASE version."
    NPM_TAG="snapshot"
else
    echo "❌ '$VERSION' is not a valid SemVer format."
    exit 1
fi

# Inject version into all platform package.json files
find npm -name package.json -exec sed -i "s/\"version\": \"[^\"]*\"/\"version\": \"${VERSION}\"/" {} \;

# Also update the optionalDependencies version references in the wrapper
sed -i "s/\(\"@jbrejner\/cli-[^\"]*\": \)\"[^\"]*\"/\1\"${VERSION}\"/g" npm/wrapper/package.json

# Pack and optionally publish platform packages first (wrapper depends on them)
find npm -name package.json ! -path "npm/wrapper/*" | while read -r f; do
  PKG_DIR="$(dirname "$f")"
  PKG_NAME="$(basename "$PKG_DIR")"
  echo "Packing ${PKG_NAME}..."
  (cd "$PKG_DIR" && npm pack)
  if [ "$DRY_RUN" = false ]; then
    echo "Publishing ${PKG_NAME}..."
    (cd "$PKG_DIR" && npm publish --tag "$NPM_TAG")
  fi
done

# Pack and optionally publish wrapper last
echo "Packing wrapper..."
(cd npm/wrapper && npm pack)
if [ "$DRY_RUN" = false ]; then
  echo "Publishing wrapper..."
  (cd npm/wrapper && npm publish --tag "$NPM_TAG")
fi
