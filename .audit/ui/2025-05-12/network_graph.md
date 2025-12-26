# Network Graph Feature

## Overview
The **Network Graph** is a new visualization tool in the MCP Any Dashboard that provides a live, interactive topology of the MCP environment. It allows users to see the relationships between the Host, Servers, and Tools at a glance.

## Key Features
- **Interactive Visualization:** Drag, zoom, and pan to explore the network.
- **Node Details:** Click on any node (Host, Server, or Tool) to view detailed metrics and status.
- **Live Status:** Visual indicators for active/idle states and traffic flow.
- **Metrics:** Real-time stats like CPU load, latency, and connection counts.

## Screenshot
![Network Graph](network_graph.png)

## Technical Details
- **Library:** Built with `@xyflow/react` (React Flow).
- **Path:** `/network`
- **Component:** `NetworkGraphClient` (`ui/src/components/network/network-graph-client.tsx`)
