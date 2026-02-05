#!/bin/bash
# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0


set -e

# Start the server
go run ./server/main.go &

# Wait for the server to start
sleep 1

# Run the client
go run ./client/main.go
