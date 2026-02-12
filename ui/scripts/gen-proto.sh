#!/bin/sh
set -e

echo "=== Protobuf Generation Debug Start ==="
echo "Current Directory: $(pwd)"
echo "Protoc Path: $(which protoc)"
echo "Protoc Version: $(protoc --version)"

PLUGIN="/app/node_modules/.bin/protoc-gen-ts_proto"
echo "Checking Plugin at $PLUGIN..."
if [ ! -f "$PLUGIN" ]; then
    echo "ERROR: Plugin not found at $PLUGIN"
    ls -la /app/node_modules/.bin/
    exit 1
fi
echo "Plugin found."

echo "Finding proto files in /proto..."
FILES=$(find /proto -name "*.proto")
if [ -z "$FILES" ]; then
    echo "ERROR: No proto files found in /proto"
    ls -R /proto
    exit 1
fi
# echo "Found files: $FILES"

echo "Running protoc..."
protoc \
  --plugin=protoc-gen-ts_proto="$PLUGIN" \
  --ts_proto_out=/proto \
  --ts_proto_opt=esModuleInterop=true \
  --ts_proto_opt=outputServices=grpc-js \
  --ts_proto_opt=env=browser \
  --proto_path=/proto \
  $FILES

echo "=== Protobuf Generation Complete ==="
