/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import dagre from 'dagre';
import { Node, Edge, Position } from '@xyflow/react';
import { TopologyGraph } from '@/hooks/use-topology';

const dagreGraph = new dagre.graphlib.Graph();
dagreGraph.setDefaultEdgeLabel(() => ({}));

const nodeWidth = 180;
const nodeHeight = 60;

/**
 * Calculates the layout for the network topology graph using Dagre.
 * @param graphData - The topology graph data from the API.
 * @returns The nodes and edges with calculated positions.
 */
export const getLayoutedElements = (
  graphData: TopologyGraph | null
): { nodes: Node[]; edges: Edge[] } => {
  if (!graphData) return { nodes: [], edges: [] };

  const nodes: Node[] = [];
  const edges: Edge[] = [];

  dagreGraph.setGraph({ rankdir: 'LR' });

  // 1. Flatten the Graph
  // Clients
  (graphData.clients || []).forEach((client) => {
      const nodeId = client.id || `client-${Math.random()}`; // Ensure ID
      nodes.push({
          id: nodeId,
          type: 'user', // Map to custom-nodes 'user' or 'client'
          data: { label: client.label || client.id },
          position: { x: 0, y: 0 },
      });
      // Edge to Core
      if (graphData.core) {
          edges.push({
              id: `e-${nodeId}-${graphData.core.id}`,
              source: nodeId,
              target: graphData.core.id,
              animated: true,
              style: { stroke: '#22c55e' }
          });
      }
  });

  // Core
  if (graphData.core) {
      nodes.push({
          id: graphData.core.id,
          type: 'agent', // Core is the Agent/Gateway
          data: { label: graphData.core.label, role: 'Gateway', status: 'Active' },
          position: { x: 0, y: 0 },
      });

      // Services (Children of Core)
      (graphData.core.children || []).forEach((service) => {
          // Services can be Middlewares or Upstreams
          // Check Type or Prefix
          // For now, map everything to 'service' node type unless specific
          const serviceId = service.id;
          nodes.push({
              id: serviceId,
              type: 'service',
              data: { label: service.label },
              position: { x: 0, y: 0 },
          });
          edges.push({
              id: `e-${graphData.core.id}-${serviceId}`,
              source: graphData.core.id,
              target: serviceId,
              animated: true,
          });

          // Tools (Children of Service)
          (service.children || []).forEach((tool) => {
              const toolId = tool.id;
              nodes.push({
                  id: toolId,
                  type: 'tool',
                  data: { label: tool.label },
                  position: { x: 0, y: 0 },
              });
              edges.push({
                  id: `e-${serviceId}-${toolId}`,
                  source: serviceId,
                  target: toolId,
                  type: 'smoothstep',
              });
          });
      });
  }

  // 2. Compute Layout
  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      targetPosition: Position.Left,
      sourcePosition: Position.Right,
      // Shift slightly to center
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    };
  });

  return { nodes: layoutedNodes, edges };
};
