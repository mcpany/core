# Example: Wrapping `kubectl`

This example demonstrates how to wrap the `kubectl` command-line tool and expose its functionality as tools through `mcpxy`. This allows an AI assistant to interact with a Kubernetes cluster.

## Overview

This example consists of two main components:

1. **`mcpxy` Configuration**: A YAML file (`config/mcpxy.yaml`) that defines how to translate `mcpxy` tool calls into `kubectl` commands.
2. **`mcpxy` Server**: The `mcpxy` instance that executes the `kubectl` commands.

## Prerequisites

- A running Kubernetes cluster.
- `kubectl` installed and configured to connect to your cluster.

## Running the Example

### 1. Build the `mcpxy` Binary

Ensure the `mcpxy` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Run the `mcpxy` Server

Start the `mcpxy` server using the provided shell script.

```bash
./start.sh
```

The `mcpxy` server will start and listen for JSON-RPC requests on port `50050`.

## Interacting with the Tool

Once the server is running, you can connect your AI assistant to `mcpxy`.

### Using Gemini CLI

1. **Add `mcpxy` as an MCP Server:**
   Register the running `mcpxy` process with the Gemini CLI.

   ```bash
   gemini mcp add mcpxy-kubectl --address http://localhost:50050 --command "sleep" "infinity"
   ```

2. **List Available Tools:**
   Ask Gemini to list the tools.

   ```bash
   gemini list tools
   ```

   You should see tools like `kubectl/-/get-pods` and `kubectl/-/get-deployments`.

3. **Call a Tool:**
   Call the `get-pods` tool to list the pods in the `default` namespace.

   ```bash
   gemini call tool kubectl/-/get-pods '{"namespace": "default"}'
   ```

   You should receive a JSON response containing the list of pods.

This example showcases how `mcpxy` can be used to create powerful integrations with existing command-line tools, enabling AI assistants to perform complex tasks like managing a Kubernetes cluster.
