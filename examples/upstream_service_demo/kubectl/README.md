# Example: Wrapping `kubectl`

This example demonstrates how to wrap the `kubectl` command-line tool and expose its functionality as tools through `MCP Any`. This allows an AI assistant to interact with a Kubernetes cluster.

## Overview

This example consists of two main components:

1. **`MCP Any` Configuration**: A YAML file (`config/mcp_any_config.yaml`) that defines how to translate `MCP Any` tool calls into `kubectl` commands.
2. **`MCP Any` Server**: The `MCP Any` instance that executes the `kubectl` commands.

## Prerequisites

- A running Kubernetes cluster.
- `kubectl` installed and configured to connect to your cluster.

> [!IMPORTANT]
> This example will not work without `kubectl` installed and configured.

## Running the Example

### 1. Build the `MCP Any` Binary

Ensure the `MCP Any` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Run the `MCP Any` Server

Start the `MCP Any` server using the provided shell script.

```bash
./start.sh
```

The `MCP Any` server will start and listen for JSON-RPC requests on port `8080`.

## Interacting with the Tool

Once the server is running, you can interact with the tools using the `gemini` CLI.

### Using the `gemini` CLI

Now, you can call the `get-pods` tool by sending a `tools/call` request.

```bash
gemini --allowed-mcp-server-names mcpany-kubectl -p "call the tool kubectl.get-pods with namespace default"
```

You can also call the `get-deployments` tool:

```bash
gemini --allowed-mcp-server-names mcpany-kubectl -p "call the tool kubectl.get-deployments with namespace default"
```

This example showcases how `MCP Any` can be used to create powerful integrations with existing command-line tools, enabling AI assistants to perform complex tasks like managing a Kubernetes cluster.
