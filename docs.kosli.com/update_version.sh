#!/bin/bash

# This script runs in netlify before the docs site is built
# It converts any instance of "%%VERSION%%" in the content directory
# to the CLI's latest version
# It should not be run outside of netlify's build process

VERSION=$(gh release list --repo kosli-dev/cli -L 1 --exclude-pre-releases --exclude-drafts | awk '{print $1}' | cut -c2-)
TARGET_DIR="content"

find "$TARGET_DIR" -type f | while read -r file; do
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/%%VERSION%%/$VERSION/g" "$file"
  else
    sed -i "s/%%VERSION%%/$VERSION/g" "$file"
  fi
done


echo "Version replaced to $VERSION in content."
