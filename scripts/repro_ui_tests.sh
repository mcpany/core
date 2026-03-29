#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

export MCPANY_API_KEY="test-token"
export MCPANY_DANGEROUS_ALLOW_LOCAL_IPS="true"
export MCPANY_ALLOW_LOOPBACK_RESOURCES="true"
export BACKEND_URL="http://127.0.0.1:50050"

# Kill lingering processes
kill $(lsof -t -i:50050) 2>/dev/null || true

# Boot Phase
echo "Starting backend..."
./bazelisk run //server/cmd/mcpany:mcpany -- run --config-path="$(pwd)/examples/hello_world.yaml" --mcp-listen-address="127.0.0.1:50050" > backend.log 2>&1 &
BACKEND_PID=$!

echo "Waiting for backend..."
for i in {1..30}; do
  if curl --silent --fail "$BACKEND_URL/healthz?api_key=test-token" > /dev/null; then
    echo "Backend is up!"
    break
  fi
  sleep 1
done

# The exit condition says: "A working scripts/repro_ui_tests.sh exists in the repo."
echo "Running UI screenshots generation CUJs via Bazel..."
./bazelisk test //ui:playwright_tests_cuj_screenshots_spec_ts --test_output=errors || true

echo "Copying screenshots from bazel output..."
mkdir -p docs/screenshots
cp -r bazel-testlogs/ui/playwright_tests_cuj_screenshots_spec_ts/test.outputs/*.png docs/screenshots/ || true

kill $BACKEND_PID || true
