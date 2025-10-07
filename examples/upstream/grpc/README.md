# gRPC Greeter Server Example

This example shows how to create a simple gRPC server and expose it as a tool through MCPXY.

## 1. Build the MCPXY binary

From the root of the `core` project, run:

```bash
make build
```

This will create the `mcpxy` binary in the `build/bin` directory.

## 2. Generate Protobuf Files

Before running the server, you need to generate the Go code from the `.proto` file. From the `greeter_server` directory, run:

```bash
./generate.sh
```

## 3. Run the Upstream gRPC Server

In a separate terminal, start the upstream gRPC greeter server. From the `greeter_server/server` directory, run:

```bash
go run main.go
```

The server will start on port 50051.

## 4. Run the MCPXY Server

In another terminal, start the MCPXY server which is configured to expose the gRPC server. From the root of the `grpc` example directory, run:

```bash
./start.sh
```

The MCPXY server will start on port 8080.

## 5. Interact with the Tool using Gemini CLI

Now you can use an AI tool like Gemini CLI to interact with the gRPC service through MCPXY.

### Configuration

First, configure Gemini CLI to use the local MCPXY server as a tool extension. You can do this by modifying the Gemini CLI configuration file (e.g., `~/.config/gemini/config.yaml`) to add the following extension:

```yaml
extensions:
  mcpxy-grpc-greeter:
    http:
      address: http://localhost:8080
```

### List Available Tools

Now, you can ask Gemini CLI to list the available tools:

```
$ gemini list tools
```

You should see the `greeter-service.SayHello` tool in the list.

### Test the Service

Finally, you can test the service by asking Gemini CLI to call the tool:

```
$ gemini call tool greeter-service.SayHello '{"name": "World"}'
```

You should see a response similar to this:

```json
{
  "message": "Hello World"
}
```
