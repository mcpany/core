#!/bin/bash
# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Install tools
apt-get update && apt-get install -y protobuf-compiler curl unzip

# Check protoc version
protoc --version

BUILD_DIR="/tmp/build"
mkdir -p "$BUILD_DIR/googleapis" "$BUILD_DIR/grpc-gateway"

# Download googleapis
echo "Downloading googleapis..."
curl -sSL -o "$BUILD_DIR/googleapis.zip" https://github.com/googleapis/googleapis/archive/refs/heads/master.zip
unzip -q "$BUILD_DIR/googleapis.zip" -d "$BUILD_DIR"
cp -r "$BUILD_DIR/googleapis-master/"* "$BUILD_DIR/googleapis/"

# Download grpc-gateway
echo "Downloading grpc-gateway..."
curl -sSL -o "$BUILD_DIR/grpc-gateway.zip" https://github.com/grpc-ecosystem/grpc-gateway/archive/refs/tags/v2.27.3.zip
unzip -q "$BUILD_DIR/grpc-gateway.zip" -d "$BUILD_DIR"
GATEWAY_DIR=$(find "$BUILD_DIR" -maxdepth 1 -name "grpc-gateway-*" -type d | head -n 1)
cp -r "$GATEWAY_DIR/"* "$BUILD_DIR/grpc-gateway/"

# Generate TypeScript protos (Source)
echo "Generating TypeScript protos (Source)..."
find /proto -name "*.proto" -exec protoc \
    --proto_path=/proto \
    --proto_path="$BUILD_DIR/grpc-gateway" \
    --proto_path="$BUILD_DIR/googleapis" \
    --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=/proto \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

# Generate Standard Protos (google/api/*.proto and google/protobuf/*.proto)
echo "Generating standard protos..."
STANDARD_PROTOS=$(find "$BUILD_DIR/googleapis/google/api" -name "*.proto")
PROTOBUF_PROTOS=$(find /usr/include/google/protobuf -name "*.proto" 2>/dev/null || true)

ALL_PROTOS="$STANDARD_PROTOS $PROTOBUF_PROTOS"
if [ -n "$ALL_PROTOS" ]; then
    # shellcheck disable=SC2086
    protoc \
        --proto_path=/usr/include \
        --proto_path="$BUILD_DIR/googleapis" \
        --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
        --ts_proto_out=/proto \
        --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
        $ALL_PROTOS
fi

# Fix struct.ts
if [ -f "/proto/google/protobuf/struct.ts" ]; then
    sed -i 's/map((e) => e)/map((e: any) => e)/g' /proto/google/protobuf/struct.ts
fi

# Fix imports path
find /proto -name "*.ts" -exec sed -i 's|\.\./\.\./\.\./google|\.\./\.\./google|g' {} +

# Clean up
rm -rf "$BUILD_DIR"
apt-get remove -y protobuf-compiler curl unzip
apt-get autoremove -y
