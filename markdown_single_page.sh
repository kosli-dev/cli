#!/bin/bash

# Output file
OUTPUT_FILE="combined_markdown.md"

# Directory containing markdown files
INPUT_DIR="./docs.kosli.com/content"

# Remove the output file if it exists
if [ -f "$OUTPUT_FILE" ]; then
    rm "$OUTPUT_FILE"
fi

# Find all markdown files, sort them by directory and filename, and append to the output file
find "$INPUT_DIR" -type f -name "*.md" | sort | while read -r file; do
    echo "## $(basename "$file")" >> "$OUTPUT_FILE"
    cat "$file" >> "$OUTPUT_FILE"
    echo -e "\n" >> "$OUTPUT_FILE"
done

echo "All markdown files have been combined into $OUTPUT_FILE"