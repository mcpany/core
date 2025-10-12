#!/bin/bash

# This script starts the mcpxy server and configures it to expose the
# tools defined in the ./config/mcpxy.yaml file.

# The root directory of the mcpxy repository.
MCPXY_ROOT_DIR="$(git rev-parse --show-toplevel)"

# The path to the mcpxy server binary.
MCPXY_SERVER_BIN="${MCPXY_ROOT_DIR}/bin/server"

# The path to the configuration file for this example.
CONFIG_PATH="./config/"

# Check if the mcpxy server binary exists.
if [ ! -f "${MCPXY_SERVER_BIN}" ]; then
    echo "Error: mcpxy server binary not found at '${MCPXY_SERVER_BIN}'"
    echo "Please build the server first by running 'make build' from the root directory."
    exit 1
fi

# Start the mcpxy server.
echo "Starting mcpxy server..."
"${MCPXY_SERVER_BIN}" --config-paths "${CONFIG_PATH}"