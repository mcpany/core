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

Once the server is running, you can interact with the tool using the `gemini` CLI.

### Using the `gemini` CLI

Now, you can call the `get_time_by_ip` tool by sending a `tools/call` request.

```bash
gemini --allowed-mcp-server-names mcpany-http -p "call the tool ip-info-service.get_time_by_ip with ip 8.8.8.8"
```

This example showcases how `mcpany` can make any HTTP API available to an AI assistant with minimal configuration.
