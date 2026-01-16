# Interactive Playground

**Status:** Implemented

## Goal

The Playground is the central specialized interface for Developers to interactively discover, test, and debug MCP tools. It replaces manual CLI calls with a rich, chat-based GUI with form support.

## Actors

- **Developer**: Testing tool implementations.
- **User**: Learning capability of a new service.

## Usage Guide

### 1. Overview

Navigate to `/playground`. The interface presents a chat interface where you can interact with tools.

### 2. Discover Tools

Click the **"Available Tools"** button (Cmd+K) to open the Tools Drawer. This lists all registered tools.

1. Browse the list or search for a tool.
2. Click on a tool (e.g., `filesystem.list_dir`) to configure it.

### 3. Configure & Execute

A configuration modal appears with a generated form based on the tool's schema.

1. Fill in the required arguments (e.g., `/var/log` for path).
   - You can toggle between **Form** view and **JSON** view for complex inputs.
2. Click **"Build Command"**.
3. The command is populated in the chat input and executed immediately.

### 4. View Results

The execution result is displayed in the chat stream.

- **Tool Call**: Shows the arguments sent to the tool.
- **Result**: Displays the JSON payload returned by the tool (expandable).
- **Error**: Displays error messages if the execution failed.

## Advanced Features

- **JSON Input**: You can manually type tool calls in the chat input using the format `tool_name {"arg": "value"}`.
- **History**: Previous tool calls and results remain visible in the chat history.
