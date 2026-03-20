# Real-time Log Streaming UI

The Real-time Log Streaming UI provides a live view of system activities, audit logs, and tool executions.

## Overview

This feature allows administrators to monitor the server's activity in real-time without needing to access raw log files on the server or query an external SIEM immediately.

## Features

- **Live Feed**: Logs appear instantly as they occur.
- **Filtering**: Filter logs by severity, source, or keyword.
- **Pause/Resume**: Pause the stream to inspect specific log entries.

## Usage

Navigate to the **Logs** section in the dashboard sidebar to view the stream.

## Implementation

The UI is implemented in `ui/src/app/logs/page.tsx` and uses the `LogStream` component (`ui/src/components/logs/log-stream.tsx`) which connects to the server's log stream (likely via WebSocket or polling).
