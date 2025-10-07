# Websocket Echo Server Example

This example shows how to create a simple websocket server and expose it as a tool through MCPXY.

## 1. Build the MCPXY binary

From the root of the `core` project, run:

```bash
make build
```

This will create the `mcpxy` binary in the `build/bin` directory.

## 2. Run the Upstream Websocket Server

In a separate terminal, start the upstream websocket echo server. From the `echo_server/server` directory, run:

```bash
go run main.go
```

The server will start on port 8080.

## 3. Run the MCPXY Server

In another terminal, start the MCPXY server which is configured to expose the websocket server. From the root of the `websocket` example directory, run:

```bash
./start.sh
```

The MCPXY server will start on port 8080. Note that the upstream server and the MCPXY server are running on the same port in this example. This is generally not recommended for production, but it is fine for this example.

## 4. Interact with the Tool using Gemini CLI

Now you can use an AI tool like Gemini CLI to interact with the websocket service through MCPXY.

### Configuration

First, configure Gemini CLI to use the local MCPXY server as a tool extension. You can do this by modifying the Gemini CLI configuration file (e.g., `~/.config/gemini/config.yaml`) to add the following extension:

```yaml
extensions:
  mcpxy-websocket-echo:
    http:
      address: http://localhost:8080
```

### List Available Tools

Now, you can ask Gemini CLI to list the available tools:

```
$ gemini list tools
```

You should see the `echo-service./echo` tool in the list.

### Test the Service

Finally, you can test the service by asking Gemini CLI to call the tool:

```
$ gemini call tool echo-service./echo '{"message": "hello"}'
```

You should see a response similar to this:

```json
{
  "message": "hello"
}
```
