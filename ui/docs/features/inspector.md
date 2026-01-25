# Service Inspector

The **Service Inspector** provides deep visibility into the communication between your MCP client and the upstream service. It acts as a "DevTools" for MCP, allowing you to inspect individual JSON-RPC messages (Requests and Responses) in real-time.

## Features

- **Real-time Traffic:** Watch MCP tool calls, resource reads, and prompt executions as they happen.
- **Filtering:** Automatically filters traces to show only those relevant to the current service.
- **Detailed Inspection:** View the full JSON payload of requests (arguments) and responses (results/errors).
- **Waterfalls:** Visualize the latency breakdown of each call.

## Usage

1. Navigate to **Services** and select a service.
2. Click the **Inspector** tab.
3. Interact with your AI assistant or use the Playground to trigger tools.
4. Observe the traces appearing in the list.
5. Click a trace to view details.

## Troubleshooting

If you see connection errors or unexpected behavior:
1. Check the **Inspector** for error responses from the upstream service.
2. Use the **Connection Diagnostic** tool (troubleshoot icon) for connectivity checks.
