# Interactive Playground

**Status:** Implemented

## Goal

The Playground is the central specialized interface for Developers to interactive discover, test, and debug MCP tools. It replaces manual CLI calls with a rich, chat-based GUI.

## Actors

- **Developer**: Testing tool implementations.
- **User**: Learning capability of a new service.

## Usage Guide

### 1. Overview

Navigate to `/playground`. The interface presents a chat-like console where you can interact with tools.

![Playground Overview](screenshots/playground_blank.png)

### 2. Select a Tool

Click the **"Available Tools"** button in the top right or use the `⌘+K` shortcut (if implemented) to open the tools drawer.

1. Browse the list of available tools.
2. Click **"Use Tool"** on a specific tool (e.g., `filesystem.list_dir`).
3. A configuration dialog will appear with a form based on the tool's JSON Schema.

![Tool Selected](screenshots/playground_tool_selected.png)

### 3. Execute Tool

Fill in the required arguments in the dialog form.

1. Enter values (e.g., `/var/log` for path).
2. Click **"Run Tool"**.
3. The command will be sent to the chat input automatically.

Alternatively, you can type commands directly in the chat input using the format:
`tool_name {"arg": "value"}`

![Form Filled](screenshots/playground_form_filled.png)

### 4. View Results

The execution result is displayed in the chat stream.

- **Success**: Shows the returned JSON payload in a collapsible block.
- **Error**: Displays the error message and code with distinct styling.
- **Diffing**: If you run the same command again and the output changes, a "Show Changes" button appears.

## Advanced Features

- **Import/Export Session**: Save and load your session history using the buttons in the top right corner.
- **Clear Chat**: Reset the session with the trash icon.
