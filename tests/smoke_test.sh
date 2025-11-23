#!/bin/bash

# Copyright (C) 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Default to the image name used in local dev if not provided
IMAGE_NAME=${1:-test-server}
CONTAINER_NAME="smoke-test-server"

echo "Starting container $CONTAINER_NAME using image $IMAGE_NAME..."

# Remove if exists
docker rm -f "$CONTAINER_NAME" > /dev/null 2>&1 || true

# Start container
# We pass flags to override default listen addresses
docker run -d --name "$CONTAINER_NAME" \
  -p 50050:50050 -p 50051:50051 \
  "$IMAGE_NAME" \
  --mcp-listen-address :50050 \
  --grpc-port :50051

echo "Waiting for container to be ready..."
sleep 5

# Check HTTP Health
echo "Checking HTTP health..."
if curl --fail --silent http://localhost:50050/healthz > /dev/null; then
  echo "HTTP Health Check Passed"
else
  echo "HTTP Health Check Failed"
  docker logs "$CONTAINER_NAME"
  docker rm -f "$CONTAINER_NAME"
  exit 1
fi

# Check gRPC Health
echo "Checking gRPC health..."
if docker run --network host --rm fullstorydev/grpcurl -plaintext localhost:50051 list; then
  echo "gRPC Health Check Passed"
else
  echo "gRPC Health Check Failed"
  docker logs "$CONTAINER_NAME"
  docker rm -f "$CONTAINER_NAME"
  exit 1
fi

echo "Smoke Test Passed!"
docker rm -f "$CONTAINER_NAME"
