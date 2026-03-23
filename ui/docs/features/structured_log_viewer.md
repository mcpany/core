# Structured Log Viewer

The **Structured Log Viewer** enhances the logging experience in MCP Any by automatically detecting and formatting JSON log messages. This feature makes it significantly easier to debug complex services that emit structured data.

## Features

- **Auto-Detection**: Automatically identifies log messages that contain valid JSON.
- **Interactive Expansion**: Collapses JSON logs by default to keep the log stream readable, with a one-click expand button.
- **Syntax Highlighting**: Displays expanded JSON with syntax highlighting for better readability.
- **Dark Mode Support**: Fully integrated with the application's dark theme.

## Usage

1.  Navigate to the **Logs** page in the MCP Any dashboard.
2.  When a log entry contains a JSON object (e.g., from a tool execution result or API response), a **Chevron (>)** icon will appear next to the message.
3.  Click the chevron to expand and view the formatted JSON structure.

## Screenshot

![Structured Log Viewer](../screenshots/structured_log_viewer.png)
