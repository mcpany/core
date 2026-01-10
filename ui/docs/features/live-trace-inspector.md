# Live Trace Inspector & Replay

**Date:** 2026-01-10
**Feature:** Live Trace Inspector & Replay
**Author:** Jules

## Overview
This feature enhances the Traces view by adding "Live Mode" for real-time monitoring and a "Replay" capability for tool calls, allowing developers to quickly debug and re-run tool executions in the Playground.

## Key Features

1.  **Live Mode Toggle:**
    *   A new toggle button in the Traces list header allows users to switch between static and live modes.
    *   When "Live" is enabled, the trace list auto-refreshes every 3 seconds.
    *   Visual indicator (Play/Pause icon and color) shows the current state.

2.  **Trace Replay:**
    *   When viewing a Trace detail that involves a Tool call, a "Replay in Playground" button is available.
    *   Clicking this button navigates to the Playground with the Tool name and Arguments pre-filled.
    *   This significantly speeds up the "Debug -> Fix -> Verify" loop.

3.  **Playground Integration:**
    *   The Playground now accepts `?tool=NAME&args=JSON` URL parameters to pre-fill the input field.
    *   It intelligently handles JSON argument formatting for better readability.

## Screenshot

![Live Trace Inspector](live_trace_inspector.png)

## Implementation Details

-   **Frontend:** Modified `TraceList` to include the toggle and `TracesPage` to handle polling. Updated `TraceDetail` to include the Replay action. Enhanced `PlaygroundClient` to parse URL parameters.
-   **State Management:** React `useState` and `useEffect` used for polling and URL parsing.
-   **Routing:** `useRouter` and `useSearchParams` used for navigation and parameter handling.
