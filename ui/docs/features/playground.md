# Interactive Playground

**Status:** Implemented

## Goal

The Playground is the central specialized interface for Developers to interactive discover, test, and debug MCP tools. It replaces manual CLI calls with a rich, form-based GUI.

## Actors

- **Developer**: Testing tool implementations.
- **User**: Learning capability of a new service.

## Usage Guide

### 1. Overview

Navigate to `/playground`. The interface presents a clean slate with a button to access available tools.

![Playground Overview](screenshots/playground_blank.png)

> **Note**: The page title is "Console".

### 2. Select a Tool

Click **"Available Tools"** to open the tool selection sheet.

1. Browse or search for the tool you wish to test.
2. Click **"Use Tool"** on a tool (e.g., `filesystem.list_dir`).
3. A dialog opens showing the **Tool Description** and a dynamically generated **Input Form**.

![Tool Selected](screenshots/playground_tool_selected.png)

### 3. Execute Tool

Fill in the required arguments. The form validates your input based on the JSON Schema provided by the tool.

1. Enter values (e.g., `/var/log` for path).
2. Click **"Build Command"** to populate the console input.
3. Press **Enter** or click the **Send** button to execute.

![Form Filled](screenshots/playground_form_filled.png)

### 4. View Results

The execution result is displayed in the chat stream.

- **Success**: Shows the returned JSON payload.
- **Error**: Displays the error message and code with distinct styling.

## Advanced Features

- **JSON Mode**: Toggle to "JSON" tab to input raw parameters if the form is too constraining.
- **History**: Previous tool calls in the session remain visible above.

### 5. Session History (Import/Export)

You can manage your playground session history using the buttons in the top right corner.

- **Export**: Save your current session to a JSON file (`playground-history-<date>.json`) for sharing or debugging.
- **Import**: Load a previously exported session file to replay tool executions and results.
