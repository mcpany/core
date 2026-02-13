# Interactive Playground

**Status:** Implemented

## Goal

The Playground is the central specialized interface for Developers to interactively discover, test, and debug MCP tools. It provides a chat-like interface for executing tools and viewing results.

## Actors

- **Developer**: Testing tool implementations.
- **User**: Learning capability of a new service.

## Usage Guide

### 1. Overview

Navigate to `/playground`. The interface presents a chat history view where you can see previous tool executions and their results.

![Playground Overview](screenshots/playground_blank.png)

### 2. Discover Tools

Click the **"Available Tools"** button in the top right to open the tools catalog.

1. A sheet will slide in listing all registered tools.
2. Browse or search for the tool you wish to test.
3. Click **"Use Tool"** on a specific item.

![Tool Selection](screenshots/playground_tool_selected.png)

### 3. Configure & Execute

A configuration dialog will open for the selected tool.

1. **Form Mode**: Fill in the required arguments using the dynamically generated form.
2. **JSON Mode**: Switch to the "JSON" tab to input raw parameters if preferred.
3. Click **"Build Command"**.

The tool command will be inserted into the chat input and executed immediately.

![Form Filled](screenshots/playground_form_filled.png)

### 4. View Results

The execution result is displayed in the chat stream.

- **Success**: Shows the returned JSON payload (expandable).
- **Error**: Displays the error message and code with distinct styling.
- **Diff View**: If you re-run a tool, you can see a diff of the output compared to the previous run.

## Advanced Features

- **Session Import/Export**: You can save your current session history to a JSON file and restore it later using the **Import** and **Export** buttons in the header.
- **Prompt Integration**: You can import executed prompt results from the **Prompt Workbench** directly into the Playground for further analysis.
