#!/bin/bash
set -e

echo "Running E2E tests for Discovery..."

# Ensure we are in the root directory
cd "$(git rev-parse --show-toplevel)"

# 1. Start Server in Background (Ensure fresh start)
echo "Killing any existing server..."
pkill -f "bin/server" || true
sleep 2

echo "Starting server..."
make build
./build/bin/server run --config-path server/config.minimal.yaml --grpc-port 50051 &
SERVER_PID=$!
# Wait for server to start
echo "Waiting for server to be ready..."
sleep 10

# 2. Seed Data
echo "Seeding data..."
# We use the /api/v1/debug/seed endpoint to inject a service with tools
# Port might be 50050 by default in config.minimal.yaml, or 8070 if MCPANY_DEFAULT_HTTP_ADDR is used.
# The server logs showed: "HTTP server listening ... port=[::]:50050"
curl -v -X POST http://localhost:50050/api/v1/debug/seed \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-secret-key" \
  -d '{
  "upstream_services": [
    {
      "name": "weather-service",
      "http_service": {
        "address": "http://weather.api",
        "tools": [
          {
            "name": "get_weather",
            "description": "Get current weather for a city",
            "input_schema": { "type": "object", "properties": { "city": { "type": "string" } } }
          },
          {
            "name": "get_forecast",
            "description": "Get 5-day weather forecast",
            "input_schema": { "type": "object", "properties": { "city": { "type": "string" } } }
          }
        ]
      }
    },
    {
      "name": "database-service",
      "http_service": {
        "address": "http://db.api",
        "tools": [
          {
            "name": "query_users",
            "description": "Search for users in the database",
            "input_schema": { "type": "object", "properties": { "query": { "type": "string" } } }
          }
        ]
      }
    }
  ]
}'

echo "Data seeded."

# 3. Test Backend Discovery API
echo "Testing Discovery API..."

# Search for "weather"
RESPONSE=$(curl -s -X POST http://localhost:50050/v1/discovery/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer my-secret-key" \
  -d '{"query": "weather"}')

echo "Search Response: $RESPONSE"

if echo "$RESPONSE" | grep -q "get_weather"; then
    echo "✅ Search found 'get_weather'"
else
    echo "❌ Search failed to find 'get_weather'"
    exit 1
fi

# Get Index Status
STATUS=$(curl -s -X GET http://localhost:50050/v1/discovery/status \
  -H "Authorization: Bearer my-secret-key")

echo "Index Status: $STATUS"

if echo "$STATUS" | grep -q '"totalTools":3'; then
    echo "✅ Index status correct (3 tools)"
else
    echo "❌ Index status incorrect"
    exit 1
fi

echo "✅ Backend Verification Passed"

# Cleanup
if [ ! -z "$SERVER_PID" ]; then
    kill $SERVER_PID
fi
