#!/bin/bash
# Copyright 2025 Author(s) of MCP Any
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
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
if docker run --network host --rm fullstorydev/grpcurl:v1.9.3 -plaintext localhost:50051 list; then
  echo "gRPC Health Check Passed"
else
  echo "gRPC Health Check Failed"
  docker logs "$CONTAINER_NAME"
  docker rm -f "$CONTAINER_NAME"
  exit 1
fi

echo "Smoke Test Passed!"
docker rm -f "$CONTAINER_NAME"
