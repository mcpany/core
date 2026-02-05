#!/bin/bash
set -e

API_URL="http://localhost:50050/api/v1"

echo "Checking if server is running..."
if ! curl -s --fail "$API_URL/health" > /dev/null; then
    # try just /health
    if ! curl -s --fail "http://localhost:50050/health" > /dev/null; then
        echo "Server is not running on port 50050. Please start it."
        exit 1
    fi
fi

echo "Server is up."

# 1. Register demo-weather (if not exists)
echo "Registering Demo Service 1 (wttr.in)..."
curl -s -X POST "$API_URL/services" -H "Content-Type: application/json" -d '{
  "name": "demo-weather",
  "version": "1.0.0",
  "http_service": {
    "address": "https://wttr.in",
    "tools": [
       {
         "name": "get_weather",
         "description": "Get weather for a location",
         "input_schema": {
            "type": "object",
            "properties": {
              "location": { "type": "string" }
            },
            "required": ["location"]
          }
       }
    ]
  },
  "tags": ["demo", "weather", "external"]
}' || echo "Failed to register demo-weather (might already exist)"

# 2. Register demo-echo
echo "Registering Demo Service 2 (echo)..."
curl -s -X POST "$API_URL/services" -H "Content-Type: application/json" -d '{
  "name": "demo-echo",
  "version": "1.0.0",
  "command_line_service": {
    "command": "echo",
    "args": ["hello world"]
  },
  "tags": ["demo", "internal"]
}' || echo "Failed to register demo-echo (might already exist)"

# 3. Seed Traffic
echo "Seeding Traffic Data..."
# Generate a JSON array of traffic points
POINTS="["
NOW=$(date +%s)
for i in {0..60}; do
    # Time format: 2024-05-20T10:00:00Z
    # Adjust for Linux/Mac date command differences?
    # sandbox likely has GNU date
    TIME=$(date -d "@$((NOW - (60-i)*60))" -u +"%Y-%m-%dT%H:%M:%SZ")

    # Random requests between 5 and 50
    REQUESTS=$((5 + RANDOM % 45))

    if [ "$i" -ne 0 ]; then POINTS="$POINTS,"; fi
    POINTS="$POINTS{\"time\": \"$TIME\", \"requests\": $REQUESTS, \"service_id\": \"demo-weather\"}"
done
POINTS="$POINTS]"

curl -s -X POST "$API_URL/debug/seed_traffic" -H "Content-Type: application/json" -d "$POINTS" || echo "Failed to seed traffic"

echo "Done seeding."
