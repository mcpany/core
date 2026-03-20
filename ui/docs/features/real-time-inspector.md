# Real-time Inspector

The **Real-time Inspector** allows you to monitor JSON-RPC traffic and tool executions as they happen, providing a seamless debugging experience.

## Features

- **Live Updates**: Traces appear instantly without requiring a page refresh.
- **WebSocket Connection**: Uses a persistent WebSocket connection to stream trace events from the server.
- **Visual Status**: A "Live" indicator shows the connection status.
- **Play/Pause**: Pause the stream to inspect a specific trace without the list scrolling.
- **History**: Upon connection, the inspector retrieves recent history so you don't miss context.

## How to Use

1. Navigate to the **Inspector** page in the sidebar.
2. Observe the "Live" badge in the header turn green.
3. Trigger tool executions or interactions with your MCP server.
4. Watch new traces appear in real-time.
5. Click on a trace to view detailed information about the request, response, and latency.

![Inspector](../screenshots/inspector.png)
