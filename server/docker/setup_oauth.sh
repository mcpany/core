#!/bin/bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Wait for Hydra to be ready
echo "Waiting for Hydra to be ready..."
for _ in {1..30}; do
    if curl -s http://localhost:5444/health/ready > /dev/null; then
        echo "Hydra is ready!"
        break
    fi
    echo "Waiting for Hydra..."
    sleep 2
done

# Create OAuth2 client
echo "Creating OAuth2 client..."
# We use the hydra binary inside the container to avoid needing it on the host
docker compose -f server/docker/docker-compose.oauth.yml exec -T hydra \
    hydra create client \
    --endpoint http://127.0.0.1:4445 \
    --name "test-client" \
    --grant-type client_credentials \
    --response-type token,code,id_token \
    --scope "openid offline" \
    --format json \
    > client_info.json

CLIENT_ID=$(grep -o '"client_id":"[^"]*"' client_info.json | cut -d'"' -f4)
CLIENT_SECRET=$(grep -o '"client_secret":"[^"]*"' client_info.json | cut -d'"' -f4)

echo "Created Client ID: $CLIENT_ID"
echo "Created Client Secret: $CLIENT_SECRET"

# Export these for the test to use (if sourcing) or write to a .env file
{
    echo "TEST_OAUTH_SERVER_URL=http://localhost:5444"
    echo "TEST_OAUTH_CLIENT_ID=$CLIENT_ID"
    echo "TEST_OAUTH_CLIENT_SECRET=$CLIENT_SECRET"
    echo "TEST_OAUTH_TOKEN_URL=http://localhost:5444/oauth2/token"
} > oauth_env.sh

echo "OAuth setup complete. Environment variables written to oauth_env.sh"
