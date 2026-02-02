#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

# Pre-commit hook to validate mcpany configuration files

# Check if mcpctl is available
MCPCTL="mcpctl"
if ! command -v $MCPCTL &> /dev/null; then
    # Try to find it in likely locations
    if [ -f "./mcpctl" ]; then
        MCPCTL="./mcpctl"
    elif [ -f "./build/bin/mcpctl" ]; then
        MCPCTL="./build/bin/mcpctl"
    else
        echo "Warning: mcpctl not found. Skipping config validation."
        exit 0
    fi
fi

# Find staged files that look like config
STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(yaml|yml|json)$')

if [ -z "$STAGED_FILES" ]; then
    exit 0
fi

has_errors=0

for file in $STAGED_FILES; do
    # Check if it looks like an mcpany config (heuristic)
    if grep -q "upstream_services" "$file" || grep -q "global_settings" "$file"; then
        echo "Validating $file..."
        if ! $MCPCTL validate --config-path "$file"; then
            echo "Error: Validation failed for $file"
            has_errors=1
        fi
    fi
done

if [ $has_errors -ne 0 ]; then
    echo "Commit aborted due to invalid configuration."
    exit 1
fi

exit 0
