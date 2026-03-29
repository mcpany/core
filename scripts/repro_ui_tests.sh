#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

echo "==> Setting up environment..."
cd "$(dirname "$0")/.."
REPO_ROOT="$(pwd)"

# Create screenshots directory
mkdir -p "${REPO_ROOT}/docs/screenshots"

echo "==> Starting backend server..."
bazelisk run //server/cmd/mcpany -- -config examples/hello_world.yaml > "${REPO_ROOT}/server.log" 2>&1 &
BACKEND_PID=$!

echo "==> Starting frontend UI..."
cd "${REPO_ROOT}/ui"
npm install

# Start Vite dev server in background
npm run dev > "${REPO_ROOT}/dev.log" 2>&1 &
FRONTEND_PID=$!

# Wait for Vite to be ready
echo "==> Waiting for servers to initialize..."
sleep 15

echo "==> Capturing UI screenshots via Playwright..."
# Install dependencies if missing
npx playwright install --with-deps chromium

# Run the screenshot generation suite
TEST_PORT=9002 npx playwright test --config=playwright.screenshots.config.ts || true

echo "==> Organizing generated screenshots..."
# Move screenshots to the centralized docs folder
mv docs/screenshots/*.png "${REPO_ROOT}/docs/screenshots/" 2>/dev/null || true

echo "==> Cleaning up..."
kill $FRONTEND_PID 2>/dev/null || true
kill $BACKEND_PID 2>/dev/null || true

echo "==> UI state capture completed successfully. Screenshots saved to docs/screenshots/"
