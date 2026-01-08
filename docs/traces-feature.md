# Feature: Request Tracer & Observability

## Overview

The **Request Tracer** is a new observability feature that allows users to visualize the execution path of complex MCP tool chains. It provides a structured "waterfall" view of request traces, including:

- **Trace List**: A searchable, filterable list of recent requests with status indicators and duration.
- **Trace Details**: A detailed view of a specific trace, showing the hierarchy of spans (steps), latencies, and input/output payloads.
- **Waterfall Visualization**: A graphical representation of the timeline of each step in the chain.

## Screenshot

![Trace View](.audit/ui/2025-05-23/trace_view.png)

## Implementation Details

- **Frontend**: Built with Next.js, using `ResizablePanel` for a flexible split-pane layout.
- **Visualization**: Custom CSS-based timeline visualization for performance and simplicity.
- **Mock API**: Currently backed by a mock API (`/api/traces`) for demonstration purposes, ready to be connected to a real backend tracing system (e.g., OpenTelemetry).

## Usage

1.  Navigate to **Traces** in the sidebar.
2.  Select a trace from the list on the left.
3.  Inspect the execution waterfall and details on the right.
