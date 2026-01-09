# Network Graph Feature

## Overview

This feature introduces a visual node-link diagram to the MCP Any dashboard, allowing users to visualize the topology of their connected services, tools, and clients.

## Screenshot

![Network Graph](./screenshots/network.png)

## Implementation Details

- **Library:** `@xyflow/react` (React Flow)
- **Components:**
  - `NetworkGraphClient`: Main container handling state and layout.
  - `Sheet`: Side panel for displaying node details.
  - `Controls`, `MiniMap`, `Background`: Standard React Flow controls.
- **Data Source:** `GET /api/v1/topology` (Mocked for now, matches Server Proto).

## Features

1.  **Visualization:** Displays Core, Services, Tools, Clients, and Middleware as distinct nodes.
2.  **Interactivity:** Click any node to view detailed metrics (QPS, Latency), status, and metadata.
3.  **Filtering:** Toggle visibility of System components (Core/Middleware) or Capability details (Tools).
4.  **Responsiveness:** Fluid layout with auto-layout capabilities.
