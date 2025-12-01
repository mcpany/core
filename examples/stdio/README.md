# Stdio Service Example

This example demonstrates how to use the `stdio` service to wrap a command-line tool.

## Running the Example

1.  **Run the `mcpany` server with the `stdio` service configuration:**

    ```bash
    make run ARGS="--config-paths examples/stdio/config.yaml"
    ```

2.  **Call the `echo` tool:**

    ```bash
    curl -X POST -H "Content-Type: application/json" \
      -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "stdio-echo/-/echo", "arguments": {"message": "Hello from stdio!"}}, "id": 1}' \
      http://localhost:50050
    ```
