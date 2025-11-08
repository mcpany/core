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

Once the server is running, you can interact with the tools using the `gemini` CLI.

### Using the `gemini` CLI

Now, you can call the `get_current_date` tool by sending a `tools/call` request.

```bash
gemini --allowed-mcp-server-names mcpany-command -p "call the tool datetime-service.get_current_date"
```

You can also call the `get_current_date_iso` tool to get the date in ISO 8601 format:

```bash
gemini --allowed-mcp-server-names mcpany-command -p "call the tool datetime-service.get_current_date_iso"
```

This example shows how easily you can extend your AI assistant with any command-line tool, opening up endless possibilities for automation and integration.
