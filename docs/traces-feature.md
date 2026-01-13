# Feature: Traffic Inspector & Replay

## Overview

The **Traffic Inspector & Replay** (formerly Request Tracer) is an advanced observability feature that allows users to visualize and interact with the execution path of complex MCP tool chains. It provides a structured "waterfall" view of request traces and enables debugging via replay.

- **Trace List**: A searchable, filterable list of recent requests with status indicators and duration.
- **Trace Details**: A detailed view of a specific trace, showing the hierarchy of spans (steps), latencies, and input/output payloads.
- **Waterfall Visualization**: A graphical representation of the timeline of each step in the chain.
- **Replay**: Ability to replay a specific trace to verify fixes or debug issues.

## Screenshot

![Trace View](images/trace_view.png)

## Implementation Details

- **Frontend**: Built with Next.js, using `ResizablePanel` for a flexible split-pane layout.
- **Visualization**: Custom CSS-based timeline visualization for performance and simplicity.
- **Mock API**: Currently backed by a mock API (`/api/traces`) for demonstration purposes, ready to be connected to a real backend tracing system (e.g., OpenTelemetry). The backend already supports OpenTelemetry tracing, but the frontend currently uses the mock API for visualization.

## Usage

1.  Navigate to **Traces** in the sidebar.
2.  Select a trace from the list on the left.
3.  Inspect the execution waterfall and details on the right.
4.  Click **Replay Trace** to re-execute the request (simulation in mock mode).
