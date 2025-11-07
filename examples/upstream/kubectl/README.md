# Example: Wrapping `kubectl`

This example demonstrates how to wrap the `kubectl` command-line tool and expose its functionality as tools through `mcpany`. This allows an AI assistant to interact with a Kubernetes cluster.

## Overview

This example consists of two main components:

1. **`mcpany` Configuration**: A YAML file (`config/mcpany.yaml`) that defines how to translate `mcpany` tool calls into `kubectl` commands.
2. **`mcpany` Server**: The `mcpany` instance that executes the `kubectl` commands.

## Prerequisites

- A running Kubernetes cluster.
- `kubectl` installed and configured to connect to your cluster.

> [!IMPORTANT]
> This example will not work without `kubectl` installed and configured.

## Running the Example

### 1. Build the `mcpany` Binary

Ensure the `mcpany` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Run the `mcpany` Server

Start the `mcpany` server using the provided shell script.

```bash
./start.sh
```

The `mcpany` server will start and listen for JSON-RPC requests on port `50050`.

## Interacting with the Tool

Once the server is running, you can interact with the tools using `curl`.

### Using `curl`

1. **Initialize a session:**
   First, send an `initialize` request to the server to establish a session. The server will respond with a session ID in the `Mcp-Session-Id` header.

   ```bash
   SESSION_ID=$(curl -i -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "initialize", "params": {"client_name": "curl-client", "client_version": "v0.0.1"}, "id": 1}' http://localhost:50050 2>/dev/null | grep -i "Mcp-Session-Id" | awk '{print $2}' | tr -d '\r')
   ```

2. **Call a tool:**
   Now, you can call the `get-pods` tool by sending a `tools/call` request with the session ID you received in the previous step.

   ```bash
   curl -X POST -H "Content-Type: application/json" -H "Mcp-Session-Id: $SESSION_ID" -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "kubectl.get-pods", "arguments": {"namespace": "default"}}, "id": 2}' http://localhost:50050
   ```

   You can also call the `get-deployments` tool:

   ```bash
   curl -X POST -H "Content-Type: application/json" -H "Mcp-Session-Id: $SESSION_ID" -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "kubectl.get-deployments", "arguments": {"namespace": "default"}}, "id": 3}' http://localhost:50050
   ```

This example showcases how `mcpany` can be used to create powerful integrations with existing command-line tools, enabling AI assistants to perform complex tasks like managing a Kubernetes cluster.
