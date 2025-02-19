#!/bin/bash

if [[ "$OSTYPE" == "darwin"* ]]; then
  SED_CMD="sed -i ''"
else
  SED_CMD="sed -i"
fi

VERSION=$(git describe --tags --abbrev=0 | cut -c2-)
TARGET_DIR="content"

find "$TARGET_DIR" -type f | while read -r file; do
  $SED_CMD "s/%%VERSION%%/$VERSION/g" "$file"
done

echo "Version replaced in content."
