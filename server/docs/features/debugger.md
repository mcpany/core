# Agent Debugger & Inspector

The Agent Debugger is a middleware that monitors and records HTTP traffic for inspection. It allows developers to "replay" traffic and inspect requests and responses to debug agent interactions.

## Configuration

To enable the debugger, configure it in your `config.yaml` (example):

```yaml
debugger:
  enabled: true
  size: 100 # Number of entries to keep in the ring buffer
```

## Features

### Traffic Inspection
The debugger captures detailed information about every request, including:
- **Timestamp**: When the request was made.
- **Method & Path**: HTTP method and URL path.
- **Status & Duration**: Response status code and execution time.
- **Headers**: Request and response headers.
- **Body**: Request and response bodies (truncated to 10KB).

### Traffic Replay
The debugger provides an API to replay requests. This is useful for:
- Debugging failed tool calls by re-executing them with modified parameters.
- Testing API endpoints without setting up a full client.

## API Reference

### Get Entries
Retrieve the captured traffic logs.

**Endpoint**: `GET /debug/entries`

**Response**:
```json
[
  {
    "id": "uuid",
    "timestamp": "2024-01-01T12:00:00Z",
    "method": "POST",
    "path": "/mcp/v1/tools/call",
    "status": 200,
    "duration": 1500000,
    "request_headers": { ... },
    "response_headers": { ... },
    "request_body": "{\"jsonrpc\": \"2.0\", ...}",
    "response_body": "{\"jsonrpc\": \"2.0\", ...}"
  }
]
```

### Replay Request
Execute an arbitrary HTTP request from the server side.

**Endpoint**: `POST /debug/replay`

**Request Body**:
```json
{
  "method": "POST",
  "url": "http://localhost:8080/mcp/v1/tools/call",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{\"jsonrpc\": \"2.0\", ...}"
}
```

**Response**:
```json
{
  "status": 200,
  "headers": { ... },
  "body": "...",
  "duration_ms": 15
}
```
