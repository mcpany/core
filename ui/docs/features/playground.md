# Interactive Playground

**Status:** Implemented

## Goal

The Playground is the central specialized interface for Developers to interactive discover, test, and debug MCP tools. It replaces manual CLI calls with a rich, form-based GUI.

## Actors

- **Developer**: Testing tool implementations.
- **User**: Learning capability of a new service.

## Usage Guide

### 1. Overview

Navigate to `/playground`. The interface presents a clean slate with access to all available tools on the left sidebar.

![Playground Overview](screenshots/playground_blank.png)

> **Note**: The page title is "Console".

### 2. Select a Tool

Browse the sidebar to find the tool you wish to test.

1. Click on a tool name (e.g., `filesystem.list_dir`).
2. The main pane updates to show the **Tool Description** and a dynamically generated **Input Form**.

![Tool Selected](screenshots/playground_tool_selected.png)

### 3. Execute Tool

Fill in the required arguments. The form validates your input based on the JSON Schema provided by the tool.

1. Enter values (e.g., `/var/log` for path).
2. Click **"Run Tool"**.

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

### 6. Key Features

The Playground includes several power-user features to enhance productivity.

#### Native File Upload
For tools that require file inputs (e.g., image analysis), the Playground automatically renders a file picker if the schema specifies `contentEncoding: "base64"` or `format: "binary"`. Selected files are automatically base64-encoded before submission.

#### Tool Output Diffing
When re-running a tool with the same arguments, the Playground detects if the output has changed. A "Show Changes" button allows you to view a side-by-side diff of the results, making it easy to spot regressions or dynamic data changes.

#### Saved Presets
You can save frequently used argument combinations as "Presets". This allows you to quickly switch between different test scenarios without re-typing complex inputs.

#### Execution Duration
Each tool execution displays its latency (time taken), helping you identify performance bottlenecks in your upstream services.

#### Context Usage Estimator
To help manage LLM context windows, the Playground estimates the token usage for each tool call and result, preventing accidental context bloat.
