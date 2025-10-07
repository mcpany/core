# HTTP Time Server Example

This example shows how to create a simple HTTP server and expose it as a tool through MCPXY.

## 1. Build the MCPXY binary

From the root of the `core` project, run:

```bash
make build
```

This will create the `mcpxy` binary in the `build/bin` directory.

## 2. Run the Upstream HTTP Server

In a separate terminal, start the upstream HTTP time server. From the `server` directory, run:

```bash
go run time_server.go
```

The server will start on port 8080.

## 3. Run the MCPXY Server

In another terminal, start the MCPXY server which is configured to expose the HTTP server. From the root of the `http` example directory, run:

```bash
./start.sh
```

The MCPXY server will start on port 8080. Note that the upstream server and the MCPXY server are running on the same port in this example. This is generally not recommended for production, but it is fine for this example.

## 4. Interact with the Tool using Gemini CLI

Now you can use an AI tool like Gemini CLI to interact with the HTTP service through MCPXY.

### Configuration

First, configure Gemini CLI to use the local MCPXY server as a tool extension. You can do this by modifying the Gemini CLI configuration file (e.g., `~/.config/gemini/config.yaml`) to add the following extension:

```yaml
extensions:
  mcpxy-http-time:
    http:
      address: http://localhost:8080
```

### List Available Tools

Now, you can ask Gemini CLI to list the available tools:

```
$ gemini list tools
```

You should see the `time-service.GET/time` tool in the list.

### Test the Service

Finally, you can test the service by asking Gemini CLI to call the tool:

```
$ gemini call tool time-service.GET/time
```

You should see a response similar to this:

```json
{
  "current_time": "2025-10-06 10:00:00",
  "timezone": "UTC"
}
```
