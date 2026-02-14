#!/bin/bash
# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Install tools dependencies
apt-get update && apt-get install -y curl unzip

# Install protoc manually (newer version required for edition="2023")
PROTOC_VERSION="29.3"
ARCH="$(uname -m)"
if [ "$ARCH" = "x86_64" ]; then
  PROTOC_ARCH="x86_64"
elif [ "$ARCH" = "aarch64" ]; then
  PROTOC_ARCH="aarch_64"
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

echo "Downloading protoc v${PROTOC_VERSION} for ${PROTOC_ARCH}..."
curl -sSL -o "/tmp/protoc.zip" "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-${PROTOC_ARCH}.zip"
unzip -o "/tmp/protoc.zip" -d /usr/local bin/protoc
unzip -o "/tmp/protoc.zip" -d /usr/local 'include/*'
rm -f "/tmp/protoc.zip"

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
# Exclude problematic proto files that are causing import errors
find /proto -name "*.proto" -not -path "*/examples/*" -not -path "*/mcp_options/*" -exec protoc \
    --proto_path=/proto \
    --proto_path="$BUILD_DIR/grpc-gateway" \
    --proto_path="$BUILD_DIR/googleapis" \
    --proto_path=/usr/local/include \
    --plugin=protoc-gen-ts_proto=/app/node_modules/.bin/protoc-gen-ts_proto \
    --ts_proto_out=/proto \
    --ts_proto_opt=esModuleInterop=true,forceLong=long,useOptionals=messages,outputClientImpl=grpc-web \
    {} +

# Generate Standard Protos (google/api/*.proto and google/protobuf/*.proto)
echo "Generating standard protos..."
STANDARD_PROTOS=$(find "$BUILD_DIR/googleapis/google/api" -name "*.proto")
PROTOBUF_PROTOS=$(find /usr/local/include/google/protobuf -name "*.proto" 2>/dev/null || true)

ALL_PROTOS="$STANDARD_PROTOS $PROTOBUF_PROTOS"
if [ -n "$ALL_PROTOS" ]; then
    # shellcheck disable=SC2086
    protoc \
        --proto_path=/usr/local/include \
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
apt-get remove -y curl unzip
apt-get autoremove -y
