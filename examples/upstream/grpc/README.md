# Example: Exposing a gRPC Service

This example demonstrates how to expose a gRPC service as a set of tools through `mcpany`.

## Overview

This example consists of three main components:

1. **Upstream gRPC Server**: A simple Go-based gRPC server (`greeter_server/`) that provides a `SayHello` RPC.
2. **`mcpany` Configuration**: A YAML file (`config/mcpxy_config.yaml`) that tells `mcpany` how to connect to the gRPC server and discover its services using gRPC reflection.
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

The `mcpany` server will start and listen for JSON-RPC requests on port `8080`.

## Interacting with the Tool

Once both servers are running, you can interact with the tool using the `gemini` CLI.

### Using the `gemini` CLI

Now, you can call the `SayHello` tool by sending a `tools/call` request.

```bash
gemini --allowed-mcp-server-names mcpany-grpc -p "call the tool greeter-service.SayHello with name World"
```

You should receive a JSON response with a greeting:

```json
{
  "message": "Hello, World"
}
```

This example highlights how `mcpany` can seamlessly integrate existing gRPC services with AI assistants, enabling them to interact with strongly-typed, high-performance APIs.
