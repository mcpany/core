# Example: Exposing an HTTP Server

This example demonstrates how to expose a public API as a tool through `MCP Any`.

## Overview

This example consists of two main components:

1. **`MCP Any` Configuration**: A YAML file (`config/mcp_any_config.yaml`) that tells `MCP Any` how to connect to the `ipinfo.io` API and what tools to expose.
2. **`MCP Any` Server**: The `MCP Any` instance that acts as a bridge between the AI assistant and the `ipinfo.io` API.

## Running the Example

### 1. Build the `MCP Any` Binary

First, ensure the `MCP Any` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Run the `MCP Any` Server

In another terminal, start the `MCP Any` server using the provided shell script. This script points `MCP Any` to the correct configuration file.

```bash
./start.sh
```

The `MCP Any` server will start and listen for JSON-RPC requests on port `8080`.

## Interacting with the Tool

Once the server is running, you can interact with the tool using the `gemini` CLI.

### Using the `gemini` CLI

Now, you can call the `get_time_by_ip` tool by sending a `tools/call` request.

```bash
gemini --allowed-mcp-server-names mcpany-http -p "call the tool ip-info-service.get_time_by_ip with ip 8.8.8.8"
```

This example showcases how `MCP Any` can make any HTTP API available to an AI assistant with minimal configuration.
