#!/bin/bash
set -e

TOOL_INSTALL_DIR="$(pwd)/build/env/bin"
export PATH="$TOOL_INSTALL_DIR:$PATH"

echo "Using protoc: $(protoc --version)"

mkdir -p ./build

# Basic check for plugins
if ! command -v protoc-gen-go >/dev/null; then
    echo "protoc-gen-go not found in $TOOL_INSTALL_DIR"
    exit 1
fi

echo "Generating protobuf files..."
find proto -name "*.proto" -exec protoc \
    --proto_path=. \
    --proto_path=build/grpc-gateway \
    --proto_path=build/googleapis \
    --descriptor_set_out=build/all.protoset \
    --include_imports \
    --go_out=. \
    --go_opt=module=github.com/mcpany/core,default_api_level=API_HYBRID \
    --go-grpc_out=. \
    --go-grpc_opt=module=github.com/mcpany/core \
    --grpc-gateway_out=. \
    --grpc-gateway_opt=module=github.com/mcpany/core \
    {} +

echo "Protobuf generation complete."
