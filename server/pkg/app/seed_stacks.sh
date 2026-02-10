#!/bin/bash

# Seed script for creating a demo stack (Collection)

PORT=${MCPANY_MCP_LISTEN_ADDRESS:-8070}
# Extract port if it's host:port
if [[ "$PORT" == *":"* ]]; then
  PORT=${PORT##*:}
fi

echo "Seeding stacks to http://localhost:$PORT..."

# Create a demo collection payload
cat <<EOF > /tmp/demo-stack.json
{
  "name": "demo-data-stack",
  "description": "A demo data engineering stack with SQLite and Filesystem.",
  "author": "Jules (Architect)",
  "version": "1.0.0",
  "services": [
    {
      "name": "demo-sqlite",
      "commandLineService": {
        "command": "npx -y @modelcontextprotocol/server-sqlite",
        "env": {
            "DB_PATH": { "plainText": ":memory:" }
        }
      },
      "tags": ["database", "demo"]
    },
    {
      "name": "demo-fs",
      "commandLineService": {
        "command": "npx -y @modelcontextprotocol/server-filesystem",
        "env": {}
      },
      "tags": ["filesystem", "demo"]
    }
  ]
}
EOF

# Post to API
# Using -v to debug if it fails
curl -v -X POST "http://localhost:$PORT/api/v1/collections" \
  -H "Content-Type: application/json" \
  -d @/tmp/demo-stack.json

echo -e "\nSeed complete."
rm /tmp/demo-stack.json
