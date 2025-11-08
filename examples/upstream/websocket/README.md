# Example: Exposing a WebSocket Service

This example demonstrates how to expose a WebSocket service as a tool through `mcpany`.

## Overview

This example consists of three main components:

1. **Upstream WebSocket Server**: A simple Go-based WebSocket server (`echo_server/`) that echoes back any message it receives.
2. **`mcpany` Configuration**: A YAML file (`config/mcpxy_config.yaml`) that tells `mcpany` how to connect to the WebSocket server.
3. **`mcpany` Server**: The `mcpany` instance that acts as a proxy between the AI assistant and the WebSocket server.

## Running the Example

### 1. Build the `mcpany` Binary

Ensure the `mcpany` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Run the Upstream WebSocket Server

In a separate terminal, start the upstream WebSocket server. From this directory (`examples/upstream/websocket`), run:

```bash
go run ./echo_server/server/main.go
```

The server will start and listen on port `8082`.

### 3. Run the `mcpany` Server

In another terminal, start the `mcpany` server using the provided script. Note that this example is configured to run the `mcpany` server on port `8081` to avoid conflicts with other examples.

```bash
./start_mcpany.sh
```

## Interacting with the Tool

Once both servers are running, you can interact with the tool using `curl`.

### Using `curl`

1. **Initialize a session:**
   First, send an `initialize` request to the server to establish a session. The server will respond with a session ID in the `Mcp-Session-Id` header.

   ```bash
   SESSION_ID=$(curl -i -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "initialize", "params": {"client_name": "curl-client", "client_version": "v0.0.1"}, "id": 1}' http://localhost:8081 2>/dev/null | grep -i "Mcp-Session-Id" | awk '{print $2}' | tr -d '\r')
   ```

2. **Call the tool:**
   Now, you can call the `echo` tool by sending a `tools/call` request with the session ID you received in the previous step.

   ```bash
   curl -X POST -H "Content-Type: application/json" -H "Mcp-Session-Id: $SESSION_ID" -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "echo-service.echo", "arguments": {"message": "Hello, WebSocket!"}}, "id": 2}' http://localhost:8081
   ```

   You should receive a JSON response echoing your message:

   ```json
   {
     "message": "Hello, WebSocket!"
   }
   ```

## Test Client

This example also includes a test client in `client.go` that demonstrates how to interact with the `mcpany` server programmatically using the Go SDK. You can run it with `go run ./client/main.go`.
