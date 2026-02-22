# Interactive Playground

**Status:** Implemented

## Goal

The Playground is the central specialized interface for Developers to interactive discover, test, and debug MCP tools. It replaces manual CLI calls with a rich, form-based GUI.

## Actors

- **Developer**: Testing tool implementations.
- **User**: Learning capability of a new service.

## Usage Guide

### 1. Overview

Navigate to `/playground`. The interface presents a clean slate.

![Playground Overview](screenshots/playground_blank.png)

> **Note**: The page title is "Console".

### 2. Select a Tool

Click the **Available Tools** button (or press `Cmd+K`) to open the tool selection sheet.

1. Click "Use Tool" on a tool card (e.g., `filesystem.list_dir`).
2. A configuration dialog opens showing the **Tool Description** and a dynamically generated **Input Form**.

![Tool Selected](screenshots/playground_tool_selected.png)

### 3. Execute Tool

Fill in the required arguments in the dialog. The form validates your input based on the JSON Schema provided by the tool.

1. Enter values (e.g., `/var/log` for path).
2. Click **"Run Tool"**.

![Form Filled](screenshots/playground_form_filled.png)

### 4. View Results

The execution result is displayed in the chat stream.

- **Success**: Shows the returned JSON payload.
- **Error**: Displays the error message and code with distinct styling.

## Advanced Features

- **JSON Mode**: Switch to the "JSON" tab in the tool configuration dialog to input raw parameters if the form is too constraining.
- **History**: Previous tool calls in the session remain visible above.
- **Saved Tool Arguments (Presets)**: Save frequently used argument combinations as presets for quick reuse.
    - Click the **Bookmark** icon in the tool dialog.
    - Click the **+** button to create a new preset.
    - Enter a name and click **Save**.
    - Click a saved preset to instantly populate the form.
- **Native File Upload**: For tools that accept file inputs (base64 encoded strings or binary format), the form automatically renders a file picker. Selected files are automatically converted to base64 strings.

### 5. Session History (Import/Export)

You can manage your playground session history using the buttons in the top right corner.

- **Export**: Save your current session to a JSON file (`playground-history-<date>.json`) for sharing or debugging.
- **Import**: Load a previously exported session file to replay tool executions and results.
