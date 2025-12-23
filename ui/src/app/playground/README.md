# Playground Feature

The Playground is an interactive chat interface that allows users to test their configured MCP tools, prompts, and resources. It simulates an AI assistant context where users can issue natural language commands, and the system demonstrates how those commands are translated into tool executions.

## Features

-   **Interactive Chat**: Communicate with the MCP agent using natural language.
-   **Tool Visualization**: See exactly which tool is called and with what arguments.
-   **Execution Results**: Inspect the JSON output returned by the tools.
-   **Mocked Backend**: Currently runs with a mocked backend for demonstration purposes, simulating latency and tool logic.

## Screenshot

![Playground UI](../../../../../.audit/ui/2025-12-23/playground_feature.png)

## usage

Navigate to `/playground` from the sidebar. Type a command like "List files" to see the `list_files` tool in action.
