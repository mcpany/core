#!/bin/bash
set -e

# Start server
echo "Starting server..."
go run server/cmd/server/main.go run --config-path server/config.minimal.yaml > server.log 2>&1 &
SERVER_PID=$!
sleep 5

# Seed service
echo "Seeding service..."
curl -v -X POST http://localhost:50050/api/v1/services \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-token" \
  -d '{
    "id": "svc_01",
    "name": "Payment Gateway",
    "version": "v1.2.0",
    "http_service": {
        "address": "https://stripe.com",
        "tools": [
            { "name": "process_payment", "description": "Process a payment" }
        ]
    }
}'

# List services
echo "Listing services..."
curl -v http://localhost:50050/api/v1/services -H "X-API-Key: test-token"

# List tools
echo "Listing tools..."
curl -v http://localhost:50050/api/v1/tools -H "X-API-Key: test-token"

# Clean up
kill $SERVER_PID
