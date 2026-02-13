#!/bin/sh
# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e
set -x

echo "=== Protobuf Generation Debug Start ==="
echo "Current Directory: $(pwd)"
echo "Protoc Path: $(which protoc)"
echo "Protoc Version: $(protoc --version)"

# Check node modules
echo "Listing /app/node_modules/.bin/:"
ls -la /app/node_modules/.bin/

PLUGIN="/app/node_modules/.bin/protoc-gen-ts_proto"
echo "Checking Plugin at $PLUGIN..."
if [ ! -f "$PLUGIN" ]; then
    echo "ERROR: Plugin not found at $PLUGIN"
    exit 1
fi
echo "Plugin found."

# Ensure plugin is executable
chmod +x "$PLUGIN"

echo "Finding proto files in /proto..."
FILES=$(find /proto -name "*.proto")
if [ -z "$FILES" ]; then
    echo "ERROR: No proto files found in /proto"
    ls -R /proto
    exit 1
fi
# echo "Found files: $FILES"

echo "Running protoc..."
# shellcheck disable=SC2086
protoc \
  --plugin=protoc-gen-ts_proto="$PLUGIN" \
  --ts_proto_out=/proto \
  --ts_proto_opt=esModuleInterop=true \
  --ts_proto_opt=outputServices=grpc-js \
  --ts_proto_opt=env=browser \
  --proto_path=/proto \
  $FILES

echo "=== Protobuf Generation Complete ==="
