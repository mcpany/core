#!/bin/bash
set -e

# Build server
echo "Building server..."
make build -C server

# Clean up previous runs
rm -f data/mcpany.db
mkdir -p data

# Start server in background
echo "Starting server..."
# Ensure build/bin exists
mkdir -p build/bin
# Check where the binary is. Makefile says build/bin/server.
export MCPANY_GRPC_PORT=50051
./build/bin/server run > server.log 2>&1 &
SERVER_PID=$!
echo "Server PID: $SERVER_PID"

# Function to clean up
cleanup() {
    echo "Stopping server..."
    kill $SERVER_PID || true
}
trap cleanup EXIT

# Wait for server to be ready
echo "Waiting for server to be ready..."
# Loop checking health
for i in {1..30}; do
    if curl -s http://localhost:50050/health > /dev/null; then
        echo "Server is ready!"
        break
    fi
    echo "Waiting for server... ($i/30)"
    sleep 1
done

# Run seeder
echo "Seeding database..."
go run server/cmd/seeder/main.go

# Run UI tests
echo "Running UI tests..."
# Ensure UI deps are installed
if [ ! -d "ui/node_modules" ]; then
    echo "Installing UI dependencies..."
    cd ui && npm install && cd ..
fi

export BACKEND_URL="http://localhost:50050"
# We rely on playwright.config.ts to start Next.js
make test-local -C ui
