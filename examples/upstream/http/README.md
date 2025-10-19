# Example: Exposing an HTTP Server

This example demonstrates how to expose a simple, Go-based HTTP server as a tool through `mcpxy`.

## Overview

This example consists of two main components:

1.  **`mcpxy` Configuration**: A YAML file (`config/mcpxy.yaml`) that tells `mcpxy` how to connect to the upstream server and what tools to expose.
2.  **`mcpxy` Server**: The `mcpxy` instance that acts as a bridge between the AI assistant and the upstream server.

## Running the Example

### 1. Build the `mcpxy` Binary

First, ensure the `mcpxy` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Run the `mcpxy` Server

In another terminal, start the `mcpxy` server using the provided shell script. This script points `mcpxy` to the correct configuration file.

```bash
./start.sh
```

The `mcpxy` server will start and listen for JSON-RPC requests on port `50050`.

## Interacting with the Tool

Once the server is running, you can connect your AI assistant to `mcpxy`.

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

    You should see the `time-service/-/get_time_by_ip` and `ip-location-service/-/getLocation` tools in the list.

3.  **Call the Tools:**
    Now, you can ask Gemini for the current time for a specific IP address.

    ```bash
    gemini 'what is the current time for 8.8.8.8'
    ```

    You should receive a JSON response with the current time, similar to this:

    ```json
    {
      "time": "2023-10-27T10:00:00Z"
    }
    ```

    You can also ask Gemini for the location of a specific IP address.

    ```bash
    gemini 'what is the location of 8.8.8.8 in json format'
    ```

    You should receive a JSON response with the location details.

This example showcases how `mcpxy` can make any HTTP API available to an AI assistant with minimal configuration.
