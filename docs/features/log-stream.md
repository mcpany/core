# Log Stream Feature

The **Log Stream** is a real-time console for monitoring events, errors, and activities across your MCP services and tools. It provides a "tail -f" experience directly in the browser.

## Features

- **Real-time Updates**: Logs are streamed instantly using Server-Sent Events (SSE).
- **Filtering**: Filter logs by severity level (INFO, WARN, ERROR, DEBUG) or text search.
- **Controls**:
    - **Pause/Resume**: Stop the stream to inspect specific logs.
    - **Clear**: Reset the view.
    - **Export**: Download the current log buffer as a `.txt` file.
- **Visuals**: Color-coded log levels and source highlighting for quick scanning.

## Screenshot

![Log Stream Interface](../../.audit/ui/2026-01-02/log_stream.png)

## Technical Implementation

- **Frontend**: React component using `EventSource` for connection management.
- **Backend**: Next.js API route (`/api/logs`) serving `text/event-stream`.
- **State Management**: Client-side buffer with auto-scroll and limit protection (max 2000 lines).

## Usage

Navigate to the **Logs** tab in the sidebar or via the global search (Cmd+K).
