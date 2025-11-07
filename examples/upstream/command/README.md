# Example: Wrapping a Command-Line Tool

This example demonstrates how to wrap the `date` command-line tool and expose its functionality as tools through `mcpany`. This powerful feature allows you to integrate any command-line tool into your AI assistant's workflow.

## Overview

This example consists of two main components:

1. **`mcpany` Configuration**: A YAML file (`config/mcpxy_config.yaml`) that defines how to translate `mcpany` tool calls into `date` commands.
2. **`mcpany` Server**: The `mcpany` instance that executes the `date` commands.

## Running the Example

### 1. Build the `mcpany` Binary

First, ensure the `mcpany` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Run the `mcpany` Server

Start the `mcpany` server using the provided shell script.

```bash
./start.sh
```

The `mcpany` server will start and listen for JSON-RPC requests on port `8080`.

## Interacting with the Tool

Once the server is running, you can interact with the tools using `curl`.

### Using `curl`

1. **Initialize a session:**
   First, send an `initialize` request to the server to establish a session. The server will respond with a session ID in the `Mcp-Session-Id` header.

   ```bash
   SESSION_ID=$(curl -i -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "initialize", "params": {"client_name": "curl-client", "client_version": "v0.0.1"}, "id": 1}' http://localhost:8080 2>/dev/null | grep -i "Mcp-Session-Id" | awk '{print $2}' | tr -d '\r')
   ```

2. **Call a tool:**
   Now, you can call the `get_current_date` tool by sending a `tools/call` request with the session ID you received in the previous step.

   ```bash
   curl -X POST -H "Content-Type: application/json" -H "Mcp-Session-Id: $SESSION_ID" -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "datetime-service.get_current_date", "arguments": {}}, "id": 2}' http://localhost:8080
   ```

   You can also call the `get_current_date_iso` tool to get the date in ISO 8601 format:

   ```bash
   curl -X POST -H "Content-Type: application/json" -H "Mcp-Session-Id: $SESSION_ID" -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "datetime-service.get_current_date_iso", "arguments": {}}, "id": 3}' http://localhost:8080
   ```

This example shows how easily you can extend your AI assistant with any command-line tool, opening up endless possibilities for automation and integration.
