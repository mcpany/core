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

### JSON Mode

Switch to the "JSON" tab in the tool configuration dialog to input raw parameters if the form is too constraining.

### Saved Tool Arguments (Presets)

You can save specific argument configurations as presets for frequently used tools.

1.  Configure the tool arguments in the input form.
2.  Click the **Presets** button (top right of the form).
3.  Enter a name for your preset and click **Save**.
4.  Later, select the saved preset from the list to instantly populate the form.
    *   Presets are stored locally in your browser (`localStorage`).

### Native File Upload

The Playground automatically detects input fields that require file content (based on the tool schema).

1.  If a field expects base64 encoded content (e.g., `content` with `encoding: base64`), a **File Upload** input will appear.
2.  Select a file from your local machine.
3.  The file content is automatically read, base64 encoded, and populated into the form field.

### Session History

Previous tool calls in the session remain visible above the input area.

-   **Persistence**: The session history is automatically saved to your browser's local storage (`playground-messages`), so it persists across page reloads.
-   **Clear**: Use the "Clear" button to reset the session.

### Import/Export History

You can manage your playground session history using the buttons in the top right corner.

-   **Export**: Save your current session to a JSON file (`playground-history-<date>.json`) for sharing or debugging.
-   **Import**: Load a previously exported session file to replay tool executions and results.
