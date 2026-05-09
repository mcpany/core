#!/bin/bash
make build || true
# wait, how to start the backend?
cd server && go build -o mcpany ./cmd/mcpany && ./mcpany start &
PID=$!
sleep 5
cd ../ui && npx playwright test tests/network-widget.spec.ts
kill $PID
