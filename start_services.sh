#!/bin/bash
kill $(lsof -t -i :9002) 2>/dev/null || true
kill $(lsof -t -i :8080) 2>/dev/null || true
kill $(lsof -t -i :50050) 2>/dev/null || true
cd server
go run cmd/server/main.go > backend.log 2>&1 &
echo "Waiting for backend..."
sleep 15
cd ../ui
npm run dev > frontend.log 2>&1 &
echo "Waiting for frontend..."
sleep 15
