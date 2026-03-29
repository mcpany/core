#!/bin/bash
set -e

# Run the backend
echo "Starting backend..."
bazelisk run //server/cmd/mcpany &
BACKEND_PID=$!

# Wait for backend to be ready
echo "Waiting for backend..."
sleep 5

# Start frontend
echo "Starting frontend dev server..."
cd ui
pnpm run dev &
FRONTEND_PID=$!

# Wait for frontend to be ready
echo "Waiting for frontend..."
sleep 5

# Run frontend tests
echo "Running UI tests to generate screenshots..."
npx playwright test --config playwright.screenshots.config.ts

# Kill frontend and backend
kill $FRONTEND_PID
kill $BACKEND_PID
echo "Done."
