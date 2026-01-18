#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Get the root directory
ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Configuration
TEST_DB_PATH="$ROOT_DIR/server/data/test-mcpany.db"
TEST_SKILLS_DIR="$ROOT_DIR/server/data/test-skills"
export BACKEND_PORT=19999
export GRPC_PORT=19998
export MCPANY_ALLOW_UNSAFE_CONFIG=true

# Cleanup
rm -f "$TEST_DB_PATH"
rm -rf "$TEST_SKILLS_DIR"
mkdir -p "$(dirname "$TEST_DB_PATH")"

# Seed
# seed-db was built in server/build/bin
"$ROOT_DIR/server/build/bin/seed-db" --db-path="$TEST_DB_PATH" --skills-dir="$TEST_SKILLS_DIR"

# Start Server
# server was built in build/bin (root)
export MCPANY_DB_PATH="$TEST_DB_PATH"
export SKILL_ROOT="$TEST_SKILLS_DIR"

"$ROOT_DIR/build/bin/server" run \
    --config-path="$ROOT_DIR/server/config.minimal.yaml" \
    --mcp-listen-address=:$BACKEND_PORT \
    --grpc-port=$GRPC_PORT
