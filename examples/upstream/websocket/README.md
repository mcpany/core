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

Once both servers are running, you can interact with the tool using the `gemini` CLI.

### Using the `gemini` CLI

Now, you can call the `echo` tool by sending a `tools/call` request.

```bash
gemini --allowed-mcp-server-names mcpany-websocket -p "call the tool echo-service.echo with message 'Hello, WebSocket!'"
```

You should receive a JSON response echoing your message:

```json
{
  "message": "Hello, WebSocket!"
}
```

## Test Client

This example also includes a test client in `client.go` that demonstrates how to interact with the `mcpany` server programmatically using the Go SDK. You can run it with `go run ./client/main.go`.
