# Example: Exposing an HTTP Server

This example demonstrates how to expose a public API as a tool through `mcpany`.

## Overview

This example consists of two main components:

1. **`mcpany` Configuration**: A YAML file (`config/mcpxy_config.yaml`) that tells `mcpany` how to connect to the `ipinfo.io` API and what tools to expose.
2. **`mcpany` Server**: The `mcpany` instance that acts as a bridge between the AI assistant and the `ipinfo.io` API.

## Running the Example

### 1. Build the `mcpany` Binary

First, ensure the `mcpany` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Run the `mcpany` Server

In another terminal, start the `mcpany` server using the provided shell script. This script points `mcpany` to the correct configuration file.

```bash
./start.sh
```

The `mcpany` server will start and listen for JSON-RPC requests on port `8080`.

## Interacting with the Tool

Once the server is running, you can interact with the tool using `curl`.

### Using `curl`

1. **Initialize a session:**
   First, send an `initialize` request to the server to establish a session. The server will respond with a session ID in the `Mcp-Session-Id` header.

   ```bash
   SESSION_ID=$(curl -i -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "initialize", "params": {"client_name": "curl-client", "client_version": "v0.0.1"}, "id": 1}' http://localhost:8080 2>/dev/null | grep -i "Mcp-Session-Id" | awk '{print $2}' | tr -d '\r')
   ```

2. **Call the tool:**
   Now, you can call the `get_time_by_ip` tool by sending a `tools/call` request with the session ID you received in the previous step.

   ```bash
   curl -X POST -H "Content-Type: application/json" -H "Mcp-Session-Id: $SESSION_ID" -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "ip-info-service.get_time_by_ip", "arguments": {"ip": "8.8.8.8"}}, "id": 2}' http://localhost:8080
   ```

This example showcases how `mcpany` can make any HTTP API available to an AI assistant with minimal configuration.
