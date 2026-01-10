# Agent Debugger & Inspector

The Agent Debugger is a middleware that monitors and records HTTP traffic for inspection. It captures request/response metadata and payloads (bodies), enabling developers to "replay" traffic and deeply inspect tool execution flow.

## Configuration

To enable the debugger, configure it in your `config.yaml`:

```yaml
debugger:
  enabled: true
  size: 100 # Number of entries to keep in the ring buffer
```

## API

The debugger exposes an endpoint (typically `/debug/entries`) to retrieve the recorded traffic logs.

### Response Format

Each entry in the debug log contains:

```json
[
  {
    "id": "uuid-string",
    "timestamp": "2024-01-01T12:00:00Z",
    "method": "POST",
    "path": "/mcp/v1/tools/call",
    "status": 200,
    "duration": 123456,
    "request_headers": { ... },
    "response_headers": { ... },
    "request_body": "{\"jsonrpc\": \"2.0\", ...}",
    "response_body": "{\"jsonrpc\": \"2.0\", ...}"
  }
]
```

### Body Capture & Privacy

- **Text Only**: The debugger attempts to capture bodies for text-based content types (JSON, XML, Text). Binary data is replaced with a placeholder.
- **Truncation**: Large bodies (default > 10KB) are truncated to prevent memory exhaustion.
- **Security Warning**: The debugger captures raw payloads, which may include sensitive data (API keys in bodies, PII). **Do not enable this in production** unless you have secured access to the debug endpoint.
