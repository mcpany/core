# Network Inspector

The **Network Inspector** provides real-time visibility into the JSON-RPC traffic between the MCP Any server and connected clients/tools. It allows you to debug communication issues by inspecting raw request and response payloads.

![Inspector Screenshot](../screenshots/inspector.png)

## Features

- **Real-time Traffic Stream**: View JSON-RPC messages as they happen.
- **Request/Response Pairing**: Automatically correlates requests with their responses (by ID, though currently displayed as linear events).
- **Deep Inspection**: Click on any event to view the full JSON payload (params, results, errors).
- **Filtering**: Search by method name or payload content.
- **Status Indicators**: Quickly identify successful operations vs errors.

## Usage

1.  Navigate to **Inspector** in the sidebar.
2.  Perform actions in the Playground or connect an external client (e.g., Claude Desktop, Gemini CLI).
3.  Watch the traffic appear in the list.
4.  Click on a row to see details.
5.  Use the **Pause** button to stop the stream and inspect a specific moment.
6.  Use **Clear** to reset the view.

## Debugging Tips

- Filter for `tools/call` to see tool executions.
- Look for `isError: true` in responses to diagnose tool failures.
- Check latency timings to identify slow operations.
