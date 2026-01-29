#!/bin/bash

set -e

# Default to the image name used in local dev if not provided
IMAGE_NAME=${1:-mcpany-ui}
CONTAINER_NAME="smoke-test-ui"
# Default to port 3000 unless overridden
HOST_PORT=${2:-3000}

echo "Starting container $CONTAINER_NAME using image $IMAGE_NAME..."

# Remove if exists
docker rm -f "$CONTAINER_NAME" > /dev/null 2>&1 || true

# Start container
docker run -d --name "$CONTAINER_NAME" \
  -p "$HOST_PORT":3000 \
  -e NEXT_PUBLIC_API_URL=http://localhost:50050 \
  "$IMAGE_NAME"

echo "Waiting for container to be ready..."

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

  if curl --fail --silent http://localhost:"$HOST_PORT" > /dev/null; then
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

echo "Smoke Test Passed!"
docker rm -f "$CONTAINER_NAME"
