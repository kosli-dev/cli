#!/bin/bash

echo "Running in: $(pwd)"
ls -la

VERSION=$(git describe --tags --abbrev=0 | cut -c2-)
TARGET_DIR="content"

find "$TARGET_DIR" -type f | while read -r file; do
  sed -i '' "s/%%VERSION%%/$VERSION/g" "$file"
done

echo "Replacement complete."
