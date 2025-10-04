#!/bin/bash

# Configuration
HEADER_FILE="scripts/dev/copyright/copyright_header.txt"
COMMENT_START=" * " # Adjusted to include space before and after asterisk
FILE_PATTERN="*.go"

echo "--- Applying header to Go files ($FILE_PATTERN) ---"

# Check if header file exists
if [ ! -f "$HEADER_FILE" ]; then
    echo "Error: Header file '$HEADER_FILE' not found."
    exit 1
fi

# Create a Go-style header template
# 1. Add "/*" at the start (first line of block comment)
# 2. Use sed to strip existing leading whitespace and prepend the comment prefix
# 3. Add " */" at the end (closing block comment)
GO_HEADER=$(
    echo "/*" 
    # Use two sed expressions: 
    # 1. Remove all leading spaces/tabs from the source line.
    # 2. Prepend the block comment prefix (COMMENT_START).
    sed -e 's/^[[:space:]]*//' -e 's/^/'"$COMMENT_START"'/' "$HEADER_FILE"
    echo
    echo " */"
)

# Apply the header to all files matching the pattern
# EXCLUSION ADDED: Use -not -name "*.pb.go" to skip protocol buffer generated files.
find . -type f -name "$FILE_PATTERN" -not -name "*.pb.go" | while read -r file; do
    # Check if the file already contains the copyright line to prevent duplication
    if grep -q "Copyright 2025 Author(s) of MCP-XY" "$file"; then
        echo "Skipping $file: Header already present."
    else
        echo "Updating $file"
        
        # Prepend the generated header followed by a newline
        (echo "$GO_HEADER"; echo ""; cat "$file") > "$file.tmp" && mv "$file.tmp" "$file"
    fi
done

echo "Go file header application complete."
