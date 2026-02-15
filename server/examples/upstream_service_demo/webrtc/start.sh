#!/bin/bash

set -e

# Start the server
go run ./server/main.go > server.log 2>&1 &

# Wait for the server to start
sleep 3

# Run the client
go run ./client/main.go
