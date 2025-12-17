#!/bin/bash
# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


# This script starts the mcpany server and configures it to expose the
# tools defined in the ./config/mcpany.yaml file.

# The root directory of the mcpany repository.

MCPANY_ROOT_DIR="$(git rev-parse --show-toplevel)"

# The path to the mcpany server binary.
MCPANY_SERVER_BIN="${MCPANY_ROOT_DIR}/build/bin/server"

# The path to the configuration file for this example.
CONFIG_PATH="$(dirname "$0")/config/"

# Check if the mcpany server binary exists.
if [ ! -f "${MCPANY_SERVER_BIN}" ]; then
    echo "Error: mcpany server binary not found at '${MCPANY_SERVER_BIN}'"
    echo "Please build the server first by running 'make build' from the root directory."
    exit 1
fi

# Start the mcpany server.
echo "Starting mcpany server..."
"${MCPANY_SERVER_BIN}" --config-paths "${CONFIG_PATH}" --stdio --logfile /tmp/mcpany.log
