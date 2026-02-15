#!/bin/bash

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
  -P \
  -e MCPANY_DANGEROUS_ALLOW_LOCAL_IPS=true \
  "$IMAGE_NAME" \
  run \
  --mcp-listen-address :50050 \
  --grpc-port :50051

echo "Waiting for container to be ready..."

# Get assigned ports
HTTP_PORT=$(docker port "$CONTAINER_NAME" 50050/tcp | head -n 1 | awk -F: '{print $2}')
GRPC_PORT=$(docker port "$CONTAINER_NAME" 50051/tcp | head -n 1 | awk -F: '{print $2}')
echo "Container mapped ports: HTTP=$HTTP_PORT, gRPC=$GRPC_PORT"

# Check HTTP Health
echo "Checking HTTP health..."
RETRIES=60
while [ $RETRIES -gt 0 ]; do
  if ! docker container inspect "$CONTAINER_NAME" >/dev/null 2>&1; then
    echo "Container $CONTAINER_NAME does not exist."
    break
  fi
  if [ "$(docker container inspect -f '{{.State.Running}}' "$CONTAINER_NAME")" != "true" ]; then
    echo "Container $CONTAINER_NAME is not running."
    break
  fi

  if curl --fail --silent http://localhost:"$HTTP_PORT"/healthz > /dev/null; then
    echo "HTTP Health Check Passed"
    break
  fi
  echo "Waiting for HTTP health check... ($RETRIES retries left)"
  sleep 1
  RETRIES=$((RETRIES-1))
done

if [ $RETRIES -eq 0 ] || [ "$(docker container inspect -f '{{.State.Running}}' "$CONTAINER_NAME")" != "true" ]; then
  echo "HTTP Health Check Failed"
  docker logs "$CONTAINER_NAME"
  docker rm -f "$CONTAINER_NAME"
  exit 1
fi

# Check gRPC Health
echo "Checking gRPC health..."
RETRIES=60
while [ $RETRIES -gt 0 ]; do
  if ! docker container inspect "$CONTAINER_NAME" >/dev/null 2>&1; then
    echo "Container $CONTAINER_NAME does not exist."
    break
  fi
  if [ "$(docker container inspect -f '{{.State.Running}}' "$CONTAINER_NAME")" != "true" ]; then
    echo "Container $CONTAINER_NAME is not running."
    break
  fi

  if docker run --network host --rm fullstorydev/grpcurl:v1.9.3 -plaintext localhost:"$GRPC_PORT" list > /dev/null 2>&1; then
    echo "gRPC Health Check Passed"
    break
  fi
  echo "Waiting for gRPC health check... ($RETRIES retries left)"
  sleep 1
  RETRIES=$((RETRIES-1))
done

if [ $RETRIES -eq 0 ] || [ "$(docker container inspect -f '{{.State.Running}}' "$CONTAINER_NAME")" != "true" ]; then
  echo "gRPC Health Check Failed"
  docker logs "$CONTAINER_NAME"
  docker rm -f "$CONTAINER_NAME"
  exit 1
fi

echo "Smoke Test Passed!"
docker rm -f "$CONTAINER_NAME"
