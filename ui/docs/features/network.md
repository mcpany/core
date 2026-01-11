# Network Topology Graph

**Status:** Implemented

## Goal
Visualize the relationships between services, tools, and clients. The Network Graph provides a topological view of the entire MCP ecosystem, making it easy to understand dependencies and routing.

## Usage Guide

### 1. View Graph
Navigate to `/network`. The graph renders automatically using a force-directed layout.
- **Nodes**: Represent Services, Clients, and Shared Resources.
- **Edges**: Represent active connections or dependencies.

![Network Graph](screenshots/network_graph.png)

### 2. Inspect Node
Click on any node (e.g., a Service node) to open the **Details Panel**.
- **Metrics**: Real-time uptime and active connection count.
- **Tools**: List of tools exposed by this service.

![Node Details](screenshots/node_detail_panel.png)

### 3. Filter and Zoom
- **Zoom**: Use your mouse wheel or trackpad to zoom in/out.
- **Drag**: Pan across the canvas to view large topologies.
- **Filter**: Use the controls to show/hide specific node types (e.g., hide offline services).
