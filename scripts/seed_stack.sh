#!/bin/bash
# Seed a stack into the MCP Any server

API_URL="${1:-http://localhost:50050}"
API_KEY="${2:-}"

echo "Seeding Stack to $API_URL..."

# Define a sample stack (collection)
# Note: Use snake_case for proto fields as per json_name annotations
STACK_JSON='{
  "name": "seed-stack",
  "description": "A seeded stack for testing",
  "version": "1.0.0",
  "services": [
    {
      "name": "seed-echo-service",
      "command_line_service": {
        "command": "echo \"hello\"",
        "env": {}
      },
      "disable": true
    }
  ]
}'

# Create the collection
curl -X POST "$API_URL/api/v1/collections" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "$STACK_JSON"

echo ""
echo "Seed complete."
