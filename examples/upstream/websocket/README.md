# Example: Exposing a WebSocket Service

This example demonstrates how to expose a WebSocket service as a tool through `mcpxy`.

## Overview

This example consists of three main components:

1.  **Upstream WebSocket Server**: A simple Go-based WebSocket server (`echo_server/`) that echoes back any message it receives.
2.  **`mcpxy` Configuration**: A YAML file (`config/mcpxy.yaml`) that tells `mcpxy` how to connect to the WebSocket server.
3.  **`mcpxy` Server**: The `mcpxy` instance that acts as a proxy between the AI assistant and the WebSocket server.

## Running the Example

### 1. Build the `mcpxy` Binary

Ensure the `mcpxy` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Run the Upstream WebSocket Server

In a separate terminal, start the upstream WebSocket server. From this directory (`examples/upstream/websocket`), run:

```bash
go run ./echo_server/main.go
```

The server will start and listen on port `8082`.

### 3. Run the `mcpxy` Server

In another terminal, start the `mcpxy` server using the provided script.

```bash
./start.sh
```

The `mcpxy` server will start and listen for JSON-RPC requests on port `50050`.

## Interacting with the Tool

Once both servers are running, you can connect your AI assistant to `mcpxy`.

### Using Gemini CLI

1.  **Add `mcpxy` as an MCP Server:**
    Register the running `mcpxy` process with the Gemini CLI.

    ```bash
    gemini mcp add mcpxy-websocket-echo --address http://localhost:50050 --command "sleep" "infinity"
    ```

2.  **List Available Tools:**
    Ask Gemini to list the tools.

    ```bash
    gemini list tools
    ```

    You should see the `websocket-echo-server/-/echo` tool in the list.

3.  **Call the Tool:**
    Call the `echo` tool with a message.

    ```bash
    gemini call tool websocket-echo-server/-/echo '{"message": "Hello, WebSocket!"}'
    ```

    You should receive a JSON response echoing your message:

    ```json
    {
      "response": "Hello, WebSocket!"
    }
    ```

This example demonstrates how `mcpxy` can expose real-time, stateful services like WebSockets to AI assistants.
