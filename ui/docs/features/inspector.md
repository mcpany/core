# Traffic Inspector

The **Inspector** provides real-time visibility into the JSON-RPC traffic between MCP clients and the MCP Any server. It is an essential tool for debugging connection issues, verifying tool execution payloads, and monitoring protocol behavior.

![Inspector](../screenshots/inspector.png)

## Features

- **Real-time Traffic Stream**: View JSON-RPC requests and responses as they happen.
- **Detailed Payload Inspection**: Inspect full JSON payloads for requests (arguments) and responses (results/errors).
- **Filtering**: Filter traffic by method name or payload content.
- **Status Indicators**: Quickly identify successful and failed operations.
- **Redaction**: Sensitive information (like API keys) is automatically redacted from the logs.

## Usage

1.  Navigate to **Inspector** in the sidebar.
2.  Use your MCP client (e.g., Claude Desktop, Gemini CLI) to interact with the server.
3.  Observe the traffic events appearing in the left pane.
4.  Click on an event to view its details in the right pane.
5.  Use the **Pause** button to freeze the stream for analysis.
6.  Use the **Search** bar to find specific methods or data.
