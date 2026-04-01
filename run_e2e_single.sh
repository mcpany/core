#!/bin/bash
cd server && go build -o mcpany ./cmd/server && ./mcpany start &
PID=$!
sleep 5
cd ../ui && npx playwright test inspector.spec.ts
kill $PID
