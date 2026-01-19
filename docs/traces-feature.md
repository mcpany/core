# Feature: Traffic Inspector & Replay

**Status:** Beta (Connected to Backend)
**Roadmap Alignment:** Interactive Tool Usage History

## Overview

The **Traffic Inspector & Replay** (formerly Request Tracer) is an advanced observability feature that allows users to visualize and interact with the execution path of complex MCP tool chains. It provides a structured "waterfall" view of request traces and enables debugging via replay.

- **Trace List**: A searchable, filterable list of recent requests with status indicators and duration.
- **Trace Details**: A detailed view of a specific trace, showing the hierarchy of spans (steps), latencies, and input/output payloads.
- **Waterfall Visualization**: A graphical representation of the timeline of each step in the chain.
- **Replay**: Ability to replay a specific trace to verify fixes or debug issues.

## Screenshot

![Trace View](images/trace_view.png)

## Implementation Details

- **Code Location:**
    -   Backend Source: `server/pkg/middleware/debugger.go` (implied source of `/debug/entries`).
    -   Frontend BFF: `ui/src/app/api/traces/route.ts` (Fetches and transforms data).
    -   UI Components: `ui/src/components/traces/trace-detail.tsx`, `ui/src/components/traces/trace-list.tsx`.
-   **Frontend**: Built with Next.js, using `ResizablePanel` for a flexible split-pane layout.
-   **Visualization**: Custom CSS-based timeline visualization for performance and simplicity.
-   **Backend Integration**: Connected to the real backend `Agent Debugger` API (`/debug/entries`).
-   **Data Transformation**: The frontend API route (`ui/src/app/api/traces/route.ts`) transforms flat debug logs from the backend into a hierarchical trace view for visualization.

## Usage

1.  Navigate to **Traces** in the sidebar.
2.  Select a trace from the list on the left.
3.  Inspect the execution waterfall and details on the right.
4.  Click **Replay in Playground** to open the trace arguments in the Interactive Playground for re-execution.
