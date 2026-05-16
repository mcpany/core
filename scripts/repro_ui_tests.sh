#!/usr/bin/env bash
# Copyright 2026 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

set -e

# Phase 0: Setup environment
cd "$(dirname "$0")/.."
echo "--- Phase 0: Environment Setup ---"

# Ensure protocol buffers are built
./bazelisk build //proto:ts_proto

# Phase 1: Boot
echo "--- Phase 1: Booting Server Stack ---"
# Start the backend in ephemeral mode (blank DB)
./bazelisk run //server:server -- --ephemeral > server.log 2>&1 &
SERVER_PID=$!

# Cleanup on exit
trap "kill $SERVER_PID 2>/dev/null || true; rm -f server.log" EXIT

echo "Waiting for backend (http://localhost:50050) to be ready..."
MAX_RETRIES=30
COUNT=0
until curl -s http://localhost:50050/healthz > /dev/null || [ $COUNT -eq $MAX_RETRIES ]; do
  sleep 2
  COUNT=$((COUNT + 1))
done

if [ $COUNT -eq $MAX_RETRIES ]; then
  echo "Error: Backend failed to start. See server.log:"
  cat server.log
  exit 1
fi
echo "Backend is up."

# Phase 2 & 3: Drive & Snapshot
echo "--- Phase 2 & 3: Driving UI and Taking Snapshots ---"
cd ui
# Ensure protos are linked for the frontend build (idempotent)
rm -rf proto && ln -s ../bazel-bin/proto proto

# Run Playwright screenshot suite
# Playwright config will handle starting the frontend via Vite
npx playwright test --config playwright.screenshots.config.ts --project=chromium --reporter=list
cd ..

# Phase 4: Relate
echo "--- Phase 4: Documentation Sync ---"
mkdir -p docs/screenshots
cp ui/docs/screenshots/*.png docs/screenshots/
# Ensure ui/docs/screenshots is also populated for localized docs
mkdir -p ui/docs/screenshots
cp docs/screenshots/*.png ui/docs/screenshots/

# Update markdown references to use correct relative paths
# Root docs: keep docs/screenshots/ or screenshots/
find docs/ ui/docs/ -maxdepth 1 -name "*.md" | xargs -r sed -i 's|docs/screenshots/|screenshots/|g'
# Subdirectory docs: use ../screenshots/
find docs/*/ ui/docs/*/ -name "*.md" | xargs -r sed -i 's|docs/screenshots/|../screenshots/|g; s|\](screenshots/|](../screenshots/|g'
# Special case for project README.md
[ -f README.md ] && sed -i 's|\](screenshots/|](docs/screenshots/|g' README.md

# Phase 5: Audit
echo "--- Phase 5: Consistency Audit ---"
MISSING_LINKS=$(grep -r "screenshots/" docs/*.md | grep -v "\[" | grep "!" || true)
if [ -n "$MISSING_LINKS" ]; then
    echo "Warning: Found potential malformed image links:"
    echo "$MISSING_LINKS"
fi

# Fix permissions (no executable bit on PNGs)
find docs/screenshots ui/docs/screenshots -name "*.png" -exec chmod 644 {} +

echo "Success! Screenshots generated and documentation updated."
echo "Visual Truth stored in docs/screenshots/"
