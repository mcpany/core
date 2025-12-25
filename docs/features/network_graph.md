# Network Graph Feature

## Overview
The Network Graph feature provides a visual representation of the MCP Any Gateway and its connected upstream services. This visualization helps operators understand the topology of their system at a glance.

## Implementation
- **Library**: React Flow (via `@xyflow/react`)
- **Layout**: Dagre for automatic hierarchical layout.
- **Components**:
    - `NetworkGraph`: Main container handling data fetching and graph state.
    - `ServiceNode`: Custom node component for Upstream Services.
    - `CentralNode`: Custom node component for the Gateway itself.

## Screenshot
![Network Graph](.audit/ui/2025-12-25/network_graph.png)
