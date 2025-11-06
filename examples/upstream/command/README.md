# Example: Wrapping a Command-Line Tool

This example demonstrates how to wrap the `date` command-line tool and expose its functionality as tools through `mcpany`. This powerful feature allows you to integrate any command-line tool into your AI assistant's workflow.

> [!NOTE]
> The examples in this directory are currently not functional. They are being updated to reflect the latest changes in the `mcpany` server.

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

Once the server is running, you can connect your AI assistant to `mcpany`.

### Using Gemini CLI

1. **Add `mcpany` as an MCP Server:**
   Register the running `mcpany` process with the Gemini CLI.

   ```bash
   gemini mcp add mcpany-command-date --address http://localhost:8080 --command "sleep" "infinity"
   ```

2. **List Available Tools:**
   Ask Gemini to list the tools.

   ```bash
   gemini list tools
   ```

   You should see the `datetime-service.get_current_date` and `datetime-service.get_current_date_iso` tools in the list.

3. **Call a Tool:**
   Call the `get_current_date` tool to get the current date.

   ```bash
   gemini call tool datetime-service.get_current_date
   ```

   You should receive a JSON response containing the current date.

This example shows how easily you can extend your AI assistant with any command-line tool, opening up endless possibilities for automation and integration.
