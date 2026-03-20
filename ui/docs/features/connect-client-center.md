# Connect Client Center

The **Connect Client Center** simplifies the process of connecting your favorite AI clients (Claude Desktop, Cursor, VS Code, etc.) to the MCP Any server.

## Features

*   **One-Click Configuration**: Generates copy-pasteable configuration snippets for popular clients.
*   **Dynamic URL Detection**: Automatically detects the correct server URL (including proxying via the UI in development modes).
*   **API Key Integration**: Allows you to input your API key to generate authenticated connection strings.
*   **Platform-Specific Instructions**: Provides paths for configuration files on macOS and Windows.

## Supported Clients

*   **Claude Desktop**: Generates `claude_desktop_config.json` entries (requires `mcp-server-sse-client` bridge or similar).
*   **Cursor**: Provides SSE connection URL for direct integration.
*   **VS Code**: Generic MCP extension configuration.
*   **Gemini CLI**: Command-line instructions.

## Screenshots

![Connect Client Dialog](../screenshots/connect-client.png)
