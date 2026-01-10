# MCP Playground Pro

The **MCP Playground Pro** is a new, premium feature designed to elevate the developer experience for testing and debugging MCP tools. It replaces the basic "Playground" with a robust, "Apple-style" chat interface that supports rich interactions, tool discovery, and execution history.

## Features

-   **Enhanced Chat Interface:** A clean, glassmorphism-inspired chat UI with distinct styles for user messages, assistant responses, tool calls, and results.
-   **Persistent Tool Sidebar:** A collapsible sidebar listing all available tools with search and filtering capabilities, allowing for quick discovery.
-   **Rich Message Rendering:**
    -   Tool arguments and results are displayed in collapsible, syntax-highlighted JSON blocks for better readability.
    -   Error messages are clearly distinct with visual indicators.
-   **Advanced Input:**
    -   Supports direct command entry in `tool_name {json_args}` format.
    -   "Quick Use" buttons in the sidebar to pre-fill the tool configuration form.
-   **Tool Configuration Dialog:** A dedicated dialog for configuring tool arguments before execution, validated against the tool's schema.

## Screenshots

![MCP Playground Pro](.audit/ui/2026-01-10/mcp_playground_pro.png)

## Usage

1.  Navigate to the **Playground** page.
2.  Browse available tools in the left sidebar.
3.  Click "Use" on a tool to open the configuration dialog, or type the command directly in the input field.
4.  Execute the tool and view the results in the chat stream.
