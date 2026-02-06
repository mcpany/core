#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Usage: SHARD=1/4 ./run_shard.sh [additional go test args]

if [ -z "$SHARD" ]; then
    echo "SHARD env var not set, running all tests..."
    SHARD_INDEX=1
    SHARD_TOTAL=1
else
    # Parse SHARD=index/total
    IFS='/' read -r SHARD_INDEX SHARD_TOTAL <<< "$SHARD"
    echo "Running shard $SHARD_INDEX of $SHARD_TOTAL"
fi

# Validate inputs
if [ -z "$SHARD_INDEX" ] || [ -z "$SHARD_TOTAL" ]; then
    echo "Error: Invalid SHARD format. Expected index/total (e.g., 1/4)"
    exit 1
fi

# Determine package pattern based on current directory
PKG_PATTERN="./..."
if [ -d "server" ] && [ -f "go.mod" ]; then
    # We are in project root
    PKG_PATTERN="./server/..."
elif [ -f "go.mod" ] && [ -d "cmd" ] && [ -d "pkg" ]; then
    # We are likely in server directory
    PKG_PATTERN="./..."
fi

# Get list of packages, excluding unwanted ones
PACKAGES=$(go list $PKG_PATTERN | \
    grep -v /tests/public_api | \
    grep -v /pkg/command | \
    grep -v /build | \
    grep -v /server/tools | \
    grep -v /tests/e2e_sequential | \
    sort)

# Select packages for this shard
SELECTED_PACKAGES=""
COUNT=0
for PKG in $PACKAGES; do
    # 1-based index for shard logic
    if [ $(( (COUNT % SHARD_TOTAL) + 1 )) -eq "$SHARD_INDEX" ]; then
        SELECTED_PACKAGES="$SELECTED_PACKAGES $PKG"
    fi
    COUNT=$((COUNT + 1))
done

if [ -z "$SELECTED_PACKAGES" ]; then
    echo "No packages selected for shard $SHARD_INDEX/$SHARD_TOTAL"
    exit 0
fi

echo "Selected packages for shard $SHARD_INDEX/$SHARD_TOTAL:"
echo "$SELECTED_PACKAGES" | tr ' ' '\n' | head -n 5
echo "... (and more)"

# Run tests
# shellcheck disable=SC2086
go test $SELECTED_PACKAGES "$@"
