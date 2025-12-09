# Stdio Service Example

This example demonstrates how to expose a command-line tool as an MCP tool using the `stdio` service type.

## 1. The Command-Line Tool

The `my-tool` is a simple Go program that reads a JSON object from stdin, extracts the `name` field, and writes a JSON object with a greeting to stdout.

## 2. The `mcpany` Configuration

The `config.yaml` file tells `mcpany` how to register the `my-tool` program as a tool.

```yaml
# examples/demo/stdio/config.yaml
upstreamServices:
  - name: "my-stdio-service"
    stdioService:
      command: "./examples/demo/stdio/my-tool-bin"
      calls:
        - operationId: "greet"
          description: "A simple tool that greets the user."
```

## 3. Running the Example

1.  **Run the `mcpany` Server**

    In a terminal, start the `mcpany` server using the provided shell script from the `examples/demo/stdio` directory:

    ```bash
    ./start.sh
    ```

    The `mcpany` server will start and listen for JSON-RPC requests on port `50050`.

2.  **Interact with the Tool**

    You can now interact with the tool using a JSON-RPC client.

    **List Tools Request:**

    ```bash
    curl -X POST -H "Content-Type: application/json" \
      -d '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' \
      http://localhost:50050
    ```

    **Call Tool Request:**

    ```bash
    curl -X POST -H "Content-Type: application/json" \
      -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "my-stdio-service/-/greet", "arguments": {"name": "world"}}, "id": 2}' \
      http://localhost:50050
    ```
