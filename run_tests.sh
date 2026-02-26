#!/bin/bash
./build/bin/server run --config-path server/config.minimal.yaml > server.log 2>&1 &
SERVER_PID=$!
echo "Server started with PID $SERVER_PID"
sleep 5
cd ui && npm test
EXIT_CODE=$?
kill $SERVER_PID
exit $EXIT_CODE
