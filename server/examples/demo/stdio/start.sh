#!/bin/bash

# This script builds and starts the MCP server with the stdio example configuration.

# The path to the MCP server binary.

MCP_SERVER_BIN="./build/bin/server"

# Check if the MCP server binary exists.
if [ ! -f "$MCP_SERVER_BIN" ]; then
  echo "MCP server binary not found. Building the server..."
  make build
fi

# Run the MCP server with the stdio example configuration.
"$MCP_SERVER_BIN" --config-paths ./examples/demo/stdio/config.yaml
