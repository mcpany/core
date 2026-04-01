#!/bin/bash
cd server && go build -o mcpany ./cmd/server && ./mcpany start &
PID=$!
sleep 5
cd ../ui && npx playwright test tests/inspector.spec.ts
kill $PID
