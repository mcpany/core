# Playground

**Status:** Implemented

## Goal

The Playground is the central specialized interface for Developers to interactively discover, test, and debug MCP tools. It replaces manual CLI calls with a rich, chat-based GUI.

## Actors

- **Developer**: Testing tool implementations.
- **User**: Learning capability of a new service.

## Usage Guide

### 1. Overview

Navigate to `/playground`. The interface presents a chat-like console where you can interact with tools.

<!-- ![Playground Overview](screenshots/playground_overview.png) -->
<!-- TODO: Update screenshot for Chat Interface -->

### 2. Select a Tool

1. Click the **"Available Tools"** button (top right) to open the tools sheet.
2. Browse the list of registered tools.
3. Click **"Use Tool"** on a tool to configure it.

<!-- ![Available Tools](screenshots/playground_tools.png) -->
<!-- TODO: Update screenshot for Tools Sheet -->

### 3. Configure & Execute Tool

A configuration dialog will open for the selected tool.

1. **Form View**: Fill in the arguments using the generated form fields.
2. **JSON View**: Switch to the "JSON" tab to edit the arguments as raw JSON.
3. **Schema View**: View the underlying JSON Schema for the tool.
4. Click **"Build Command"**.
5. The command (e.g., `tool_name {args}`) will be populated in the input box.
6. Press **Enter** or click the **Send** button to execute.

<!-- ![Configure Tool](screenshots/playground_configure.png) -->
<!-- TODO: Update screenshot for Config Dialog -->

### 4. View Results

The execution result is displayed in the chat stream.

- **Success**: Shows the returned JSON payload in an expandable block.
- **Error**: Displays the error message in a red alert box.
- **Diff**: If a tool is executed multiple times with the same arguments, a **"Show Changes"** button appears to diff the previous and current output.

### 5. Session History (Import/Export)

You can manage your playground session history using the buttons in the top right corner.

- **Export**: Save your current chat session to a JSON file (`playground-session-<date>.json`) for sharing or debugging.
- **Import**: Load a previously exported session file to replay tool executions and results.

## Advanced Features

- **JSON Mode**: You can directly type `tool_name {"key": "value"}` in the input box without using the form builder.
- **Clear Chat**: Use the trash icon to clear the current session history.
