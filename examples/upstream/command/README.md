# Example: Wrapping a Command-Line Tool

This example demonstrates how to wrap a simple shell script and expose it as a tool through `mcpany`. This powerful feature allows you to integrate any command-line tool into your AI assistant's workflow.

## Overview

This example consists of three main components:

1. **Upstream Script**: A simple shell script (`server/hello.sh`) that prints a greeting.
2. **`mcpany` Configuration**: A YAML file (`config/mcpany.yaml`) that tells `mcpany` how to execute the script.
3. **`mcpany` Server**: The `mcpany` instance that runs the script and returns its output.

## Running the Example

### 1. Build the `mcpany` Binary

First, ensure the `mcpany` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Make the Script Executable

The shell script must be executable. From this directory (`examples/upstream/command`), run:

```bash
chmod +x ./server/hello.sh
```

### 3. Run the `mcpany` Server

Start the `mcpany` server using the provided shell script.

```bash
./start.sh
```

The `mcpany` server will start and listen for JSON-RPC requests on port `50050`.

## Interacting with the Tool

Once the server is running, you can connect your AI assistant to `mcpany`.

### Using Gemini CLI

1. **Add `mcpany` as an MCP Server:**
   Register the running `mcpany` process with the Gemini CLI.

   ```bash
   gemini mcp add mcpany-command-hello --address http://localhost:50050 --command "sleep" "infinity"
   ```

2. **List Available Tools:**
   Ask Gemini to list the tools.

   ```bash
   gemini list tools
   ```

   You should see the `command-hello-world/-/hello` tool in the list.

3. **Call the Tool:**
   Call the tool to execute the script.

   ```bash
   gemini call tool command-hello-world/-/hello
   ```

   You should see the output of the script:

   ```
   Hello, World!
   ```

This example shows how easily you can extend your AI assistant with any command-line tool, opening up endless possibilities for automation and integration.
