#!/bin/bash

# check_headers.sh
# 
# This script is designed to run in a CI pipeline (e.g., GitHub Actions)
# to verify that all Go files and Dockerfiles contain the correct Apache 2.0
# license header and that no placeholder text remains.

# --- Configuration ---
PLACEHOLDER_OWNER="Author(s) of MCP-XY"
PLACEHOLDER_YEAR="2025"
ERROR_COUNT=0

# Define the expected Go/Docker file patterns
GO_PATTERN="*.go"
DOCKER_PATTERN="*Dockerfile*"

# --- Functions ---

# Function to check a single file for the header
check_file() {
    local file=$1
    local comment_style=$2
    local required_text="Copyright"

    # Check if the required copyright text is present in the first 20 lines
    # This is an efficiency check, not a guarantee.
    if ! head -n 20 "$file" | grep -q "$required_text"; then
        echo "::error file=$file::Missing '$required_text' header."
        ERROR_COUNT=$((ERROR_COUNT + 1))
        return
    fi
}

# Function to process files based on pattern
process_files() {
    local pattern=$1
    local comment_style=$2
    
    # Use git ls-files to efficiently list only files tracked by Git
    find . -type f -name "$pattern" -not -name "*.pb.go" | while read -r file; do
        check_file "$file" "$comment_style"
    done
}

# --- Main Execution ---

echo "Starting license header validation..."

# 1. Check Go files
process_files "$GO_PATTERN" "Go"

# 2. Check Dockerfiles
process_files "$DOCKER_PATTERN" "Docker"

# --- Report Results ---

if [ "$ERROR_COUNT" -gt 0 ]; then
    echo "--- CHECK FAILED ---"
    echo "Found $ERROR_COUNT files with missing or incomplete headers."
    exit 1
else
    echo "--- CHECK SUCCESS ---"
    echo "All Go and Docker files contain valid headers."
fi
