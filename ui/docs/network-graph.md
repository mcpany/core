# Network Graph Visualization

## Overview
The Network Graph feature provides a real-time, interactive topology view of the entire MCP ecosystem. It visualizes the relationships between Clients, the MCP Any Core Server, Services, Tools, and Resources. This "Portainer for MCP" style view allows administrators to quickly understand the system architecture and health.

## Key Features
-   **Hierarchical Visualization**: Automatically lays out nodes from Clients -> Core -> Services -> Tools.
-   **Real-time Metrics**: Displays live QPS, Latency, and Error Rates for nodes and connections.
-   **Interactive Details**: Clicking any node opens a detailed inspector panel with full metadata and metric history.
-   **Status Indication**: Visual cues (color coding) indicate the health status of services and tools (Active, Inactive, Error).
-   **Auto-Layout**: Intelligent "Dagre" layout engine ensures a clean, readable graph even with complex topologies.

## Usage
1.  Navigate to the **Network** tab in the main sidebar.
2.  Use the **Zoom/Pan** controls to navigate the graph.
3.  Click **Refresh** to force a topology update, or rely on the 5-second auto-poll.
4.  Click **Auto-Layout** to reset the view.
5.  **Select a Node** to view detailed properties in the right-side sheet.

## Screenshot
![Network Graph Visualization](../../.audit/ui/2025-12-29/network_graph.png)
