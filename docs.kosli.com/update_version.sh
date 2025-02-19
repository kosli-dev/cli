#!/bin/bash

VERSION=$(git describe --tags --abbrev=0 | cut -c2-)
TARGET_DIR="content"

find "$TARGET_DIR" -type f | while read -r file; do
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s/%%VERSION%%/$VERSION/g" "$file"
  else
    sed -i "s/%%VERSION%%/$VERSION/g" "$file"
  fi
done


echo "Version replaced to $VERSION in content."
