# Example: Exposing an HTTP Server

This example demonstrates how to expose a simple, Go-based HTTP server as a tool through `mcpxy`.

## Overview

This example consists of two main components:

1. **`mcpxy` Configuration**: A YAML file (`config/mcpxy.yaml`) that tells `mcpxy` how to connect to the upstream server and what tools to expose.
2. **`mcpxy` Server**: The `mcpxy` instance that acts as a bridge between the AI assistant and the upstream server.

## Running the Example

### 1. Build the `mcpxy` Binary

First, ensure the `mcpxy` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Run the `mcpxy` Server

In another terminal, start the `mcpxy` server using the provided shell script. This script points `mcpxy` to the correct configuration file.

```bash
examples/upstream/http/start.sh
```

The `mcpxy` server will start and listen for JSON-RPC requests on port `50050`.

## Interacting with the Tool

Once the server is running, you can connect your AI assistant to `mcpxy`.

### Using Gemini CLI

1. **Add `mcpxy` as an MCP Server:**
   Use the `gemini mcp add` command to register the running `mcpxy` process. Note that the `start.sh` script must be running in another terminal.

   ```bash
   gemini mcp add --transport http mcpxy-server http://localhost:50050
   ```

   Confirm the addition is successful:

   ```bash
   gemini mcp list
   ```

2. **Call the Tools:**
   Now, you can ask Gemini for the current time for a specific IP address.

   ```bash
   gemini -m gemini-2.5-flash -p 'what is the current time for IP 8.8.8.8'
   ```

This example showcases how `mcpxy` can make any HTTP API available to an AI assistant with minimal configuration.
