# Example: Exposing a gRPC Service

This example demonstrates how to expose a gRPC service as a set of tools through `mcpany`.

## Overview

This example consists of three main components:

1. **Upstream gRPC Server**: A simple Go-based gRPC server (`greeter_server/`) that provides a `SayHello` RPC.
2. **`mcpany` Configuration**: A YAML file (`config/mcpany.yaml`) that tells `mcpany` how to connect to the gRPC server and discover its services using gRPC reflection.
3. **`mcpany` Server**: The `mcpany` instance that bridges the AI assistant and the gRPC server.

## Running the Example

### 1. Build the `mcpany` Binary

Ensure the `mcpany` binary is built. From the root of the repository, run:

```bash
make build
```

### 2. Run the Upstream gRPC Server

In a separate terminal, start the upstream gRPC server. From this directory (`examples/upstream/grpc`), run:

```bash
go run ./greeter_server/server/main.go
```

The server will start and listen on port `50051`.

### 3. Run the `mcpany` Server

In another terminal, start the `mcpany` server using the provided script.

```bash
./start.sh
```

The `mcpany` server will start and listen for JSON-RPC requests on port `50050`.

## Interacting with the Tool

Once both servers are running, you can interact with the tool using `curl`.

### Using `curl`

1. **Initialize a session:**
   First, send an `initialize` request to the server to establish a session. The server will respond with a session ID in the `Mcp-Session-Id` header.

   ```bash
   SESSION_ID=$(curl -i -X POST -H "Content-Type: application/json" -d '{"jsonrpc": "2.0", "method": "initialize", "params": {"client_name": "curl-client", "client_version": "v0.0.1"}, "id": 1}' http://localhost:50050 2>/dev/null | grep -i "Mcp-Session-Id" | awk '{print $2}' | tr -d '\r')
   ```

2. **Call the tool:**
   Now, you can call the `SayHello` tool by sending a `tools/call` request with the session ID you received in the previous step.

   ```bash
   curl -X POST -H "Content-Type: application/json" -H "Mcp-Session-Id: $SESSION_ID" -d '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "greeter-service.SayHello", "arguments": {"name": "World"}}, "id": 2}' http://localhost:50050
   ```

   You should receive a JSON response with a greeting:

   ```json
   {
     "message": "Hello, World"
   }
   ```

This example highlights how `mcpany` can seamlessly integrate existing gRPC services with AI assistants, enabling them to interact with strongly-typed, high-performance APIs.
