# 🧪 Examples

This document provides examples of how to use the MCP Any. It includes instructions on how to run the example services and how to interact with the server.

## Running the Examples

The examples are located in the `server/examples/upstream_service_demo` directory. Each example includes a server that can be run to demonstrate how to use MCP Any with a different type of service.

### Weather Example

The weather example demonstrates how to use MCP Any with a HTTP service.

1. **Start the main server:**

   ```bash
   make run
   ```

2. **Start the example weather server:**
   In a new terminal window, run the following command:

   ```bash
   go run server/examples/upstream_service_demo/http/server/weather_server/weather_server.go
   ```

### Greeter Service Example

The greeter service example demonstrates how to use MCP Any with a gRPC service that uses gRPC reflection.

1. **Start the main server:**

   ```bash
   make run
   ```

2. **Start the example greeter service server:**
   In a new terminal window, run the following command:

   ```bash
   go run server/examples/upstream_service_demo/grpc/greeter_server/server/main.go
   ```

## Interacting with the Server

You can interact with the MCP Any server using its JSON-RPC API. The following examples use `curl`.

### List Tools

To list the available tools, run the following command:

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' \
  http://localhost:50050
```

### Call a Tool

To call a tool, you need to know the tool's name and the required inputs. For example, to use the `GetWeather` tool from the weather example (assuming it is registered as `weather/-/GetWeather`), you would run:

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "weather/-/GetWeather", "arguments": {"location": "London"}}, "id": 2}' \
  http://localhost:50050
```
