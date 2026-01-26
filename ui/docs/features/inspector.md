# Inspector

The Inspector provides a real-time console for debugging MCP JSON-RPC traffic and tool executions. Unlike the Traces view, which is designed for deep-dive analysis of individual requests, the Inspector offers a streamlined, log-like view perfect for monitoring traffic as it happens.

![Inspector](../screenshots/inspector.png)

## Key Features

### Live Mode
Toggle "Live Mode" to automatically refresh the trace list every 2 seconds. This is ideal for watching tool calls in real-time while interacting with an AI assistant.

### Filtering
Easily find relevant traces with client-side filtering:
- **Search**: Filter by method name or Trace ID.
- **Status**: Filter by "Success" or "Error" to quickly identify failures.

### Detailed Inspection
Click on any row to open the detailed trace view, which includes:
- Request and Response payloads (JSON).
- Execution waterfall timeline.
- Diagnostics and suggestions for errors.
