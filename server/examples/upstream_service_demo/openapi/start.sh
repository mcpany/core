#!/bin/bash

set -e

# Start the server
go run ./server/main.go &

# Wait for the server to start
sleep 1

# Run the client
go run ./client/main.go
