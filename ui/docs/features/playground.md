# Interactive Playground

**Status:** Implemented

## Goal

The Playground is the central specialized interface for Developers to interactively discover, test, and debug MCP tools. It replaces manual CLI calls with a rich, form-based GUI.

## Actors

- **Developer**: Testing tool implementations.
- **User**: Learning capability of a new service.

## Usage Guide

### 1. Overview

Navigate to `/playground`. The interface presents a chat-like console where you can execute tools and view results.

> **Note**: The page title is "Console" or "Playground".

### 2. Select a Tool

1. Click the **"Available Tools"** button in the top toolbar to open the tools sheet.
2. Browse or search for the tool you wish to test (e.g., `filesystem.list_dir`).
3. Click **"Use Tool"** next to the desired tool.

### 3. Configure & Execute

A dialog will appear with the **Tool Description** and a dynamically generated **Input Form**.

1. Fill in the required arguments. The form validates your input based on the JSON Schema provided by the tool.
2. Click **"Build Command"**. This will construct the tool call and immediately execute it in the console.

### 4. View Results

The execution result is displayed in the chat stream.

- **Success**: Shows the returned JSON payload (expandable).
- **Error**: Displays the error message and code with distinct styling.

## Advanced Features

- **JSON Mode**: In the tool configuration dialog, switch to the "JSON" tab to input raw parameters if the form is too constraining.
- **History**: Previous tool calls in the session remain visible in the chat history.
- **Presets**: Save frequently used argument sets as presets for quick access.

### 5. Session History (Import/Export)

You can manage your playground session history using the buttons in the top right corner.

- **Export**: Save your current session to a JSON file (`playground-session-<date>.json`) for sharing or debugging.
- **Import**: Load a previously exported session file to replay tool executions and results.
