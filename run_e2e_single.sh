#!/bin/bash
bazelisk run //server/cmd/mcpany -- -config server/config.minimal.yaml &
PID=$!
sleep 5
cd ui
export PATH="$PATH:$HOME/.npm-global/bin"
TEST_PORT=9002 playwright test $1
kill $PID
