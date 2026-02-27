#!/bin/bash
set -e

VERSION="$1"
if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  exit 1
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

# Publish platform packages first (wrapper depends on them)
find npm -name package.json ! -path "npm/wrapper/*" | while read -r f; do
  echo "Publishing $(basename "$(dirname "$f")")..."
  (cd "$(dirname "$f")" && npm publish --tag "$NPM_TAG")
done

# Publish wrapper last
echo "Publishing wrapper..."
(cd npm/wrapper && npm publish --tag "$NPM_TAG")
