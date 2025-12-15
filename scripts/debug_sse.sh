#!/bin/bash
# ... (same as before but log in cwd)
SERVER_BIN="./build/bin/server"
CONFIG_FILE="/tmp/mcp_debug_config.json"
LOG_FILE="./server_debug.log"

cat > "$CONFIG_FILE" <<EOF
{
  "global_settings": {
    "mcp_listen_address": "127.0.0.1:0"
  },
  "upstream_services": []
}
EOF

# Start server
"$SERVER_BIN" run --config-path "$CONFIG_FILE" > "$LOG_FILE" 2>&1 &
SERVER_PID=$!
echo "Server started with PID $SERVER_PID"

# Wait for server to be ready
sleep 5
grep "HTTP server listening" "$LOG_FILE"
PORT=$(grep "HTTP server listening" "$LOG_FILE" | grep -oE "port=127.0.0.1:[0-9]+" | cut -d: -f2)
echo "Server Port: $PORT"

if [ -z "$PORT" ]; then
  echo "Failed to get port"
  cat "$LOG_FILE"
  kill "$SERVER_PID"
  exit 1
fi

# Function to test endpoint
test_endpoint() {
  URL="http://127.0.0.1:$PORT$1"
  echo "Testing $URL with GET and Accept: text/event-stream"
  curl -v -H "Accept: text/event-stream" "$URL" 2>&1 | head -n 20
  echo "----------------------------------------"
}

test_get_dual() {
  URL="http://127.0.0.1:$PORT$1"
  echo "Testing GET $URL with dual Accept header"
  curl -v -H "Accept: application/json, text/event-stream" "$URL" 2>&1 | head -n 20
  echo "----------------------------------------"
}

test_get_dual "/"
test_get_dual "/sse"

test_post() {
  URL="http://127.0.0.1:$PORT/mcp"
  echo "Testing POST $URL with ping"
  curl -v -X POST -H "Content-Type: application/json" -H "Accept: text/event-stream" -d '{"jsonrpc": "2.0", "method": "ping", "id": 0}' "$URL" 2>&1 | head -n 30
  echo "----------------------------------------"
}

test_post

# Clean up
kill "$SERVER_PID"
rm "$CONFIG_FILE"
rm "$LOG_FILE"
