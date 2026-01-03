# Network Graph Feature

## Overview
The Network Graph is a new visualization tool that displays the connected topology of the MCP Any instance. It provides a real-time (or near real-time) view of:

-   **MCP Any Core**: The central hub.
-   **Upstream Services**: Connected tools and resources (e.g., Filesystem, Linear).
-   **Agents**: Connected clients (e.g., Claude Desktop, Cursor).

## Implementation Details
-   **Library**: `@xyflow/react` (React Flow)
-   **Components**: `NetworkGraphClient` handles the rendering and state.
-   **Features**:
    -   Interactive zoom/pan.
    -   Node details on click (Slide-over sheet).
    -   Status indicators (color-coded nodes).
    -   Legend panel.

## Visual Verification
![Network Graph](network_graph.png)
