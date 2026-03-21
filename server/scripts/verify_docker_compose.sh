#!/bin/bash
# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%dT%H:%M:%S%z')] $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%dT%H:%M:%S%z')] ERROR: $1${NC}"
    exit 1
}

cleanup() {
    log "Cleaning up..."
    docker logs core-mcpany-server-1 || true
    docker logs docker-compose-demo-mcpany-server-1 || true
    docker compose down -v 2>/dev/null || true
    docker compose -f examples/docker-compose-demo/docker-compose.yml down -v 2>/dev/null || true
}
trap cleanup EXIT

# 1. Build the mcpany/server docker image
log "Step 1: Building mcpany/server docker image..."
docker build -t ghcr.io/mcpany/server:latest -f docker/Dockerfile.server . || error "Failed to build docker image"

# 2. Start root docker-compose (Production)
log "Step 2: Starting root docker-compose (Production)..."
docker compose up -d --wait || error "Failed to start root docker-compose"

# 3. Verify mcpany-server health
log "Step 3: Verifying mcpany-server health..."
if curl --fail --silent http://localhost:50050/healthz > /dev/null; then
    log "mcpany-server is healthy."
else
    error "mcpany-server is unhealthy."
fi

# 4. Verify Prometheus metrics scraping
log "Step 4: Verifying Prometheus metrics..."
# Wait a bit for Prometheus to scrape
sleep 15
# Query Prometheus for up metric
if curl --fail --silent "http://localhost:9099/api/v1/query?query=up" | grep '"result":\[{"metric":{.*},"value":\[.*,"1"\]}\]'; then
    log "Prometheus is successfully scraping targets."
else
     # Dump prometheus targets for debugging
    log "Prometheus targets:"
    curl -s http://localhost:9099/api/v1/targets | grep "up" || true
    error "Prometheus failed to scrape targets or query failed."
fi

# 5. Start example docker-compose (Demo)
log "Step 5: Starting example docker-compose (Demo)..."
# We need to stop the root one first as they might share ports if we didn't change them?
# checks ports...
# Root: 50050, 50051, 6379, 9090
# Example: 50050, 50051, 8080, 8081
# Conflict on 50050, 50051.
log "Stopping root docker-compose to avoid port conflicts..."
docker compose down

log "Starting examples/docker-compose-demo/docker-compose.yml..."
cd examples/docker-compose-demo
docker compose up -d --wait || error "Failed to start example docker-compose"

# 6. Verify example functionality
log "Step 6: Verifying example functionality..."
if curl --fail --silent http://localhost:50050/healthz > /dev/null; then
    log "Example mcpany-server is healthy."
else
    error "Example mcpany-server is unhealthy."
fi

# Verify Echo Server connectivity via MCP (Mocking a client call or checking logs?)
# For now, just checking if echo server itself is up (it has its own port 8080)
if curl --fail --silent http://localhost:8080/health > /dev/null; then
    log "Example http-echo-server is healthy."
else
    error "Example http-echo-server is unhealthy."
fi

log "Validation passed successfully!"
