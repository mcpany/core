# Network Graph Feature

## Overview
The Network Graph is a new visualization tool that allows users to see the topology of their MCP ecosystem. It displays connections between the MCP Host, connected Servers, and available Tools.

## Key Features
- **Interactive Graph:** Users can zoom, pan, and rearrange nodes.
- **Real-time Status:** Nodes show status indicators (Active/Idle/Error).
- **Details Panel:** Clicking a node reveals detailed metrics and metadata.
- **Visual Distinction:** Different node types (Host, Server, Tool) are visually distinct.

## Screenshot
![Network Graph](.audit/ui/2025-12-26/network_graph.png)

## Technical Details
- Built with `@xyflow/react` (React Flow).
- Custom node components using Tailwind CSS for styling.
- responsive layout.
