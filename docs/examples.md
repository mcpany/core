# 🧪 Examples

This document provides examples of how to use the MCP Any. It includes instructions on how to run the example services and how to interact with the server.

## Running the Examples

The examples are located in the `proto/examples` directory. Each example includes a server that can be run to demonstrate how to use MCP Any with a different type of service.

### Weather Example

The weather example demonstrates how to use MCP Any with a gRPC service.

1. **Start the main server:**

   ```bash
   make server
   ```

2. **Start the example weather server:**
   In a new terminal window, run the following command:

   ```bash
   go run proto/examples/weather/v1/server/main.go
   ```

### User Service Example

The user service example demonstrates how to use MCP Any with a gRPC service that uses gRPC reflection.

1. **Start the main server:**

   ```bash
   make server
   ```

2. **Start the example user service server:**
   In a new terminal window, run the following command:

   ```bash
   go run proto/examples/userservice/v1/server/main.go
   ```

## Interacting with the Server

You can interact with the MCP Any server using any gRPC client. The following examples use `grpcurl`.

### List Tools

To list the available tools, run the following command:

```bash
grpcurl -plaintext localhost:8080 mcpany.mcp_router.v1.McpRouter/ListTools
```

### Execute a Tool

To execute a tool, you need to know the tool's name and the required inputs. For example, to use the `GetWeather` tool from the weather example, you would run the following command:

```bash
grpcurl -plaintext -d '{"tool_id": "weather/GetWeather", "inputs": {"location": "London"}}' localhost:8080 mcpany.mcp_router.v1.McpRouter/ExecuteTool
```
