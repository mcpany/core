# Example: Exposing an HTTP Server

This example demonstrates how to expose a simple, Go-based HTTP server as a tool through `mcpxy`.

## Overview

This example consists of three main components:
1.  **Upstream Server**: A simple Go application (`server/time_server.go`) that serves the current time on an HTTP endpoint.
2.  **`mcpxy` Configuration**: A YAML file (`config/mcpxy.yaml`) that tells `mcpxy` how to connect to the upstream server and what tools to expose.
3.  **`mcpxy` Server**: The `mcpxy` instance that acts as a bridge between the AI assistant and the upstream server.

## Running the Example

### 1. Build the `mcpxy` Binary

First, ensure the `mcpxy` binary is built. From the root of the repository, run:
```bash
make build
```

### 2. Run the Upstream HTTP Server

In a separate terminal, start the upstream HTTP time server. From this directory (`examples/upstream/http`), run:
```bash
go run ./server/time_server.go
```
The server will start and listen on port `8081`.

### 3. Run the `mcpxy` Server

In another terminal, start the `mcpxy` server using the provided shell script. This script points `mcpxy` to the correct configuration file.
```bash
./start.sh
```
The `mcpxy` server will start and listen for JSON-RPC requests on port `50050`.

## Interacting with the Tool

Once both servers are running, you can connect your AI assistant to `mcpxy`.

### Using Gemini CLI

1.  **Add `mcpxy` as an MCP Server:**
    Use the `gemini mcp add` command to register the running `mcpxy` process. Note that the `start.sh` script must be running in another terminal.
    ```bash
    # The 'gemini mcp add' command requires a command to run, but our server is already running.
    # We can use 'sleep infinity' as a placeholder command.
    # The key is to point the Gemini CLI to the correct address where mcpxy is listening.
    gemini mcp add mcpxy-http-time --address http://localhost:50050 --command "sleep" "infinity"
    ```

2.  **List Available Tools:**
    Ask Gemini to list the tools exposed by `mcpxy`.
    ```bash
    gemini list tools
    ```
    You should see the `http-time-server/-/get_time` tool in the list.

3.  **Call the Tool:**
    Now, you can call the tool to get the current time.
    ```bash
    gemini call tool http-time-server/-/get_time
    ```

    You should receive a JSON response with the current time, similar to this:
    ```json
    {
      "time": "2023-10-27T10:00:00Z"
    }
    ```

This example showcases how `mcpxy` can make any HTTP API available to an AI assistant with minimal configuration.