# Interactive Playground

**Status:** Implemented

## Goal

The Playground is the central specialized interface for Developers to interactively discover, test, and debug MCP tools. It replaces manual CLI calls with a rich, chat-based GUI (Console).

## Actors

- **Developer**: Testing tool implementations.
- **User**: Learning capability of a new service.

## Usage Guide

### 1. Overview

Navigate to `/playground`. The interface presents a chat console ("Console") with access to all available tools on the left sidebar.

### 2. Select a Tool

Browse the sidebar to find the tool you wish to test.

1. Click on a tool name (e.g., `filesystem.list_dir`).
2. A configuration dialog opens, showing the **Tool Description** and a dynamically generated **Input Form**.

### 3. Configure & Execute

Fill in the required arguments in the form. The form validates your input based on the JSON Schema provided by the tool.

1. Enter values (e.g., `/var/log` for path).
2. Click **"Submit"** (or similar action) to populate the chat command.
3. The command is entered into the chat input. Press **Enter** to execute.

Alternatively, you can type commands directly in the chat input using the format: `tool_name {"argument": "value"}`.

### 4. View Results

The execution result is displayed in the chat stream as a structured message.

- **Success**: Shows the returned JSON payload.
- **Error**: Displays the error message and code with distinct styling.
- **Diffing**: If you re-run a tool with the same arguments, the playground will highlight differences in the output compared to the previous run.

## Advanced Features

- **Direct JSON Input**: Type raw JSON arguments directly in the chat input for quick execution.
- **Dry Run**: Toggle "Dry Run" mode to simulate tool execution without side effects (if supported by the tool).
- **History**: Previous tool calls in the session remain visible in the chat history.
- **Share**: Generate a shareable URL with the current tool and arguments pre-filled.

### 5. Session History (Import/Export)

You can manage your playground session history using the buttons in the top right corner.

- **Export**: Save your current session to a JSON file (`playground-history-<date>.json`) for sharing or debugging.
- **Import**: Load a previously exported session file to replay tool executions and results.
- **Clear**: Clear the current session history.
