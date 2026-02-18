# Interactive Playground

**Status:** Implemented

## Goal

The Playground is the central specialized interface for Developers to interactive discover, test, and debug MCP tools. It replaces manual CLI calls with a rich, form-based GUI.

## Actors

- **Developer**: Testing tool implementations.
- **User**: Learning capability of a new service.

## Usage Guide

### 1. Overview

Navigate to `/playground`. The interface presents a chat-like console with quick access to common actions.

![Playground Overview](screenshots/playground_blank.png)

### 2. Select a Tool

1. Click the **"Available Tools"** button in the toolbar.
2. A sheet will open listing all registered tools.
3. Click **"Use Tool"** on the desired tool (e.g., `filesystem.list_dir`).
4. A **Configuration Dialog** opens, displaying the tool description and a dynamically generated **Input Form**.

![Tool Selected](screenshots/playground_tool_selected.png)

### 3. Execute Tool

Fill in the required arguments in the dialog. The form validates your input based on the JSON Schema provided by the tool.

1. Enter values (e.g., `/var/log` for path).
2. Click **"Build Command"**.
3. The command is populated in the input bar. Press **Enter** or click **Send** to execute.

![Form Filled](screenshots/playground_form_filled.png)

### 4. View Results

The execution result is displayed in the chat stream.

- **Success**: Shows the returned JSON payload.
- **Error**: Displays the error message and code with distinct styling.

## Advanced Features

- **JSON Mode**: Within the Configuration Dialog, switch to the **"JSON"** tab to input raw parameters if the form is too constraining.
- **History**: Previous tool calls in the session remain visible above.

### 5. Session History (Import/Export)

You can manage your playground session history using the buttons in the top right corner.

- **Export**: Save your current session to a JSON file (`playground-session-<date>.json`) for sharing or debugging.
- **Import**: Load a previously exported session file to replay tool executions and results.
