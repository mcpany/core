#!/bin/bash
# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


set -e

# Start the server
go run ./server/main.go > server.log 2>&1 &

# Wait for the server to start
sleep 3

# Run the client
go run ./client/main.go
