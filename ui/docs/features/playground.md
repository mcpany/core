# Interactive Playground

**Status:** Implemented

## Goal

The Playground is the central specialized interface for Developers to discover, test, and debug MCP tools interactively. It replaces manual CLI calls with a rich, form-based GUI.

## Actors

- **Developer**: Testing tool implementations.
- **User**: Learning capability of a new service.

## Usage Guide

### 1. Overview

Navigate to `/playground`. The interface presents a clean slate with a chat-like history.

![Playground Overview](screenshots/playground_blank.png)

> **Note**: The page title is "Playground".

### 2. Select a Tool

Click the **Available Tools** button (or press `Cmd+K` if configured) to open the tool selection sheet.

1. Browse the list of registered tools, grouped by service.
2. Click **"Use Tool"** on a tool card (e.g., `filesystem.list_dir`).
3. A configuration dialog opens showing the **Tool Description** and a dynamically generated **Input Form**.

![Tool Selected](screenshots/playground_tool_selected.png)

### 3. Configure & Execute Tool

Fill in the required arguments in the dialog. The form validates your input based on the JSON Schema provided by the tool.

- **Form Mode**: Standard input fields for arguments.
- **JSON Mode**: Switch to the "JSON" tab to input raw parameters if the form is too constraining.
- **Schema Mode**: View the raw JSON schema for the tool input.
- **Presets**: Select from saved presets (if available) to quickly populate common arguments.
- **Native File Upload**: For tools accepting base64 encoded files, use the native file picker.

Click **"Run Tool"** to execute.

![Form Filled](screenshots/playground_form_filled.png)

### 4. View Results

The execution result is displayed in the chat stream.

- **Success**: Shows the returned JSON payload in a collapsible viewer.
- **Error**: Displays the error message and code with distinct styling.
- **Duration**: The execution time (latency) is displayed next to the result (e.g., `45ms`).

## Advanced Features

### Tool Output Diffing
If you run the same tool multiple times, you can compare the output of the current execution with the previous one. Click **"Show Changes"** on a result card to view a side-by-side diff.

### Session History (Import/Export)
You can manage your playground session history using the buttons in the top right corner.

- **Export**: Save your current session (including inputs, outputs, and errors) to a JSON file (`playground-session-<date>.json`) for sharing or debugging.
- **Import**: Load a previously exported session file to replay tool executions and view past results.

### Code Generation
From the tool configuration dialog, you can generate client code for the current tool call:
- Click the **Code** dropdown (</>)
- Select **Copy as Curl** or **Copy as Python** to get a ready-to-run snippet.
