# GraphQL Upstream Example

This example demonstrates how to configure MCP Any to connect to a GraphQL upstream service.

## Running the Example

1.  **Start the GraphQL server:**

    ```bash
    go run ./examples/graphql/server/server.go
    ```

2.  **Start the MCP Any server with the GraphQL configuration:**

    ```bash
    make run ARGS="--config-paths ./examples/graphql/config.yaml"
    ```

3.  **Call the `hello` tool:**

    ```bash
    curl -X POST -H "Content-Type: application/json" \
      -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "graphql-hello", "arguments": {}}, "id": 1}' \
      http://localhost:50050
    ```
