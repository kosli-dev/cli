#!/bin/bash
set -e

VERSION="$1"
if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

# When called from goreleaser, $2 is "true" if snapshot build
DRY_RUN=false
if [ "$2" = "true" ] || [ "$2" = "--dry-run" ]; then
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
while IFS= read -r f; do
  tmp="$(mktemp)"
  jq --arg v "$VERSION" '.version = $v' "$f" > "$tmp" && mv "$tmp" "$f" || { rm -f "$tmp"; exit 1; }
done < <(find npm -name package.json)

# Also update the optionalDependencies version references in the wrapper
tmp="$(mktemp)"
jq --arg v "$VERSION" '.optionalDependencies = (.optionalDependencies | with_entries(.value = $v))' \
  npm/wrapper/package.json > "$tmp" && mv "$tmp" npm/wrapper/package.json

# Build ordered package list: platform packages first, wrapper last
PACKAGES=()
while IFS= read -r f; do
  PACKAGES+=("$(dirname "$f")")
done < <(find npm -name package.json ! -path "npm/wrapper/*" | sort)
PACKAGES+=("npm/wrapper")

# Phase 1: pack all packages — exit immediately on any failure
for PKG_DIR in "${PACKAGES[@]}"; do
  PKG_NAME="$(basename "$PKG_DIR")"
  echo "Packing ${PKG_NAME}..."
  (cd "$PKG_DIR" && npm pack) || { echo "❌ Failed to pack ${PKG_NAME}"; exit 1; }
done

# Phase 2: publish all packages if not a dry run — exit immediately on any failure
npm_publish_with_retry() {
  local pkg_dir="$1"
  local tag="$2"
  local max_attempts=3
  local delay=5

  for attempt in $(seq 1 "$max_attempts"); do
    local provenance_flag=""
    [ "${GITHUB_ACTIONS:-false}" = "true" ] && provenance_flag="--provenance"
    if (cd "$pkg_dir" && npm publish --tag "$tag" $provenance_flag); then
      return 0
    fi
    if [ "$attempt" -lt "$max_attempts" ]; then
      echo "⚠️  Attempt ${attempt}/${max_attempts} failed. Retrying in ${delay}s..."
      sleep "$delay"
      delay=$(( delay * 2 ))
    fi
  done
  return 1
}

if [ "$DRY_RUN" = false ]; then
  for PKG_DIR in "${PACKAGES[@]}"; do
    PKG_NAME="$(basename "$PKG_DIR")"
    echo "Publishing ${PKG_NAME}..."
    npm_publish_with_retry "$PKG_DIR" "$NPM_TAG" || { echo "❌ Failed to publish ${PKG_NAME} after retries"; exit 1; }
  done
fi
