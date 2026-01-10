# Agent Debugger & Inspector

The Agent Debugger is a middleware that monitors and records HTTP traffic for inspection. It captures detailed information about requests and responses, including headers and bodies, allowing developers to debug agent interactions and "replay" tool calls.

## Features

- **Traffic Logging**: Records Method, Path, Status, Duration, Headers.
- **Body Capture**: Captures Request and Response bodies (JSON, XML, Text).
- **Truncation**: Automatically truncates large bodies (default 10KB) to prevent memory issues.
- **Ring Buffer**: Stores a fixed number of recent entries in memory.

## Configuration

To enable the debugger, configure it in your `config.yaml`:

```yaml
debugger:
  enabled: true
  size: 100 # Number of entries to keep in the ring buffer
```

## API

The debugger exposes an endpoint (typically `/debug/entries`) to retrieve the recorded traffic logs.

### Get Entries

`GET /debug/entries`

**Response:**

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2026-01-01T12:00:00Z",
    "method": "POST",
    "path": "/mcp/v1/tools/call",
    "status": 200,
    "duration": 15000000,
    "request_headers": {
      "Content-Type": ["application/json"]
    },
    "response_headers": {
      "Content-Type": ["application/json"]
    },
    "request_body": "{\"jsonrpc\": \"2.0\", \"method\": \"weather_get\", \"params\": {\"city\": \"London\"}}",
    "response_body": "{\"jsonrpc\": \"2.0\", \"result\": {\"temperature\": 15}}"
  }
]
```

## Replay

With the captured `request_body`, you can replay a request using:
1.  **Curl**: Copy the body and headers and run a curl command.
2.  **Playground**: Paste the JSON-RPC payload into the MCP Any Playground.
3.  **HTTP Client**: Use Postman or similar tools.
