#!/bin/bash

# This script checks if the mcpany-server service in docker-compose.yml
# is configured to be built from source instead of using a pre-built image.

set -e

# The docker-compose.yml file to check
DOCKER_COMPOSE_FILE="docker-compose.yml"

# Check if the mcpany-server service has a 'build' key and not an 'image' key
if ! grep -q 'mcpany-server:' "$DOCKER_COMPOSE_FILE"; then
  echo "Error: The mcpany-server service is not defined in $DOCKER_COMPOSE_FILE."
  exit 1
fi

# This is a bit brittle, but it's the most reliable way to check without a yaml parser
if ! grep -A 5 'mcpany-server:' "$DOCKER_COMPOSE_FILE" | grep -q 'build:'; then
    echo "Error: The mcpany-server service in $DOCKER_COMPOSE_FILE is not configured to be built from source."
    exit 1
fi

if grep -A 5 'mcpany-server:' "$DOCKER_COMPOSE_FILE" | grep -q 'image:'; then
    echo "Error: The mcpany-server service in $DOCKER_COMPOSE_FILE is configured to use a pre-built image."
    exit 1
fi


echo "Success: The mcpany-server service in $DOCKER_COMPOSE_FILE is correctly configured to be built from source."
exit 0
