# Live Logs Stream

**Status:** Implemented

## Goal
Monitor centralized logs from the Gateway and all connected microservices. The Live Logs view aggregates standard output/error streams into a single, searchable console.

## Usage Guide

### 1. Stream Logs
Navigate to `/logs`. The view connects to the log WebSocket and begins streaming events immediately.
- **Color Coding**:
  - **Blue**: INFO
  - **Yellow**: WARN
  - **Red**: ERROR

![Logs Stream](screenshots/logs_stream.png)

### 2. Search and Filter
Use the search bar at the top to filter logs by keyword (e.g., "error", "payment-service").
- **Service Filter**: Select a specific service from the dropdown to isolate its logs.
- **Level Filter**: Show only Warning/Error logs.

![Filtered Logs](screenshots/logs_filtered.png)

### 3. Pause and Resume
- **Scroll Up**: Scrolling up automatically pauses the live tail, allowing you to read history.
- **Resume**: Click the "Resume" button (or scroll to bottom) to re-enable auto-scrolling.
