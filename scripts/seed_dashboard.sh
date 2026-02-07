#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

API_URL="http://localhost:50050/api/v1"

echo "🌱 Seeding Dashboard Data..."

# 1. Register a Weather Service (wttr.in)
echo "   -> Registering 'demo-weather' (wttr.in)..."
curl -s -X POST "$API_URL/services/registry" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "demo-weather",
    "http_service": {
        "address": "http://wttr.in",
        "calls": {
            "get_weather": {
                "endpoint_path": "/{args.city}?format=j1",
                "method": "HTTP_METHOD_GET",
                "output_transformer": {
                    "format": "JSON",
                    "extraction_rules": {
                        "temperature": "{.current_condition[0].temp_C}",
                        "humidity": "{.current_condition[0].humidity}",
                        "desc": "{.current_condition[0].weatherDesc[0].value}"
                    }
                }
            }
        }
    }
}' > /dev/null
echo "      Done."


# 2. Register an Echo Service (httpbin) - Adjusted for compatibility
echo "   -> Registering 'demo-echo' (httpbin.org)..."
# Note: Removed "args" from extraction rules as it caused schema issues in previous runs.
# Keeping it simple to ensure registration succeeds.
curl -s -X POST "$API_URL/services/registry" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "demo-echo",
    "http_service": {
        "address": "http://httpbin.org",
        "calls": {
            "echo_get": {
                "endpoint_path": "/get",
                "method": "HTTP_METHOD_GET",
                "output_transformer": {
                    "format": "JSON",
                    "extraction_rules": {
                        "url": "{.url}",
                        "origin": "{.origin}"
                    }
                }
            }
        }
    }
}' > /dev/null
echo "      Done."

# 3. Seed Traffic/Metrics (using debug endpoint if available, or simulating calls)
echo "   -> Seeding traffic history..."
# Try to use the debug seed endpoint if it exists
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/debug/seed_traffic")

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "      Seeded via debug endpoint."
else
    echo "      Debug endpoint not found. Simulating traffic..."
    # Simulate some calls to generate metrics
    for _ in {1..5}; do
        curl -s "$API_URL/services/demo-weather/tools/get_weather/call" \
            -H "Content-Type: application/json" \
            -d '{"city": "London"}' > /dev/null &
    done
    wait
    echo "      Traffic simulated."
fi

echo "✅ Dashboard Seeded!"
