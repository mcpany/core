#!/bin/bash

# Configuration
HEADER_FILE="scripts/dev/copyright/copyright_header.txt"
COMMENT_START="# "
FILE_PATTERN="*Dockerfile*" # Matches Dockerfile, Dockerfile.dev, etc.

echo "--- Applying header to Dockerfiles ($FILE_PATTERN) ---"

# Check if header file exists
if [ ! -f "$HEADER_FILE" ]; then
    echo "Error: Header file '$HEADER_FILE' not found."
    exit 1
fi

# Create a Dockerfile-style header template
# 1. Use sed to strip any existing leading whitespace (s/^[[:space:]]*//).
# 2. Use sed to prepend the comment prefix ("# ").
DOCKER_HEADER=$(sed -e 's/^[[:space:]]*//' -e 's/^/'"$COMMENT_START"'/' "$HEADER_FILE")

# Apply the header to all files matching the pattern
find . -type f -name "$FILE_PATTERN" | while read -r file; do
    # Check if the file already contains the copyright line to prevent duplication
    if grep -q "Copyright 2025 Author(s) of MCPXY" "$file"; then
        echo "Skipping $file: Header already present."
    else
        echo "Updating $file"
        
        # Prepend the generated header followed by a newline
        (echo "$DOCKER_HEADER"; echo ""; cat "$file") > "$file.tmp" && mv "$file.tmp" "$file"
    fi
done

echo "Dockerfile header application complete."
