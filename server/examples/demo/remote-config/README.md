# Remote Configuration Example

This example demonstrates how to load an MCP Any server configuration from a remote URL. This is useful for sharing and reusing configurations without having to manually copy and paste files.

## Running the Example

1.  **Start the mock server:**

    In a separate terminal, run the mock server to serve the remote configuration file:

    ```bash
    go run mock_server.go
    ```

2.  **Run the MCP Any server:**

    In another terminal, run the MCP Any server with the remote configuration URL:

    ```bash
    make run ARGS="--config-paths http://localhost:8080/config.yaml"
    ```

3.  **List the available tools:**

    You should see the `hello-remote` tool from the remote configuration file in the list of available tools:

    ```bash
    curl -X POST -H "Content-Type: application/json" \
      -d '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' \
      http://localhost:50050
    ```

    You should see a response like this:

    ```json
    {
      "jsonrpc": "2.0",
      "result": {
        "tools": [
          {
            "name": "hello-remote/-/get_user",
            "description": "Get user by ID"
          }
        ]
      },
      "id": 1
    }
    ```
