/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Trace, Span } from '@/types/trace';
import { Node, Edge, MarkerType } from '@xyflow/react';
import dagre from 'dagre';

/**
 * Node data structure for custom nodes.
 */
interface NodeData {
  label: string;
  role?: string;
  status?: string;
}

/**
 * traceToGraph converts a Trace object into ReactFlow nodes and edges with automatic layout.
 * @param trace - The trace object to visualize.
 * @returns An object containing nodes and edges.
 */
export function traceToGraph(trace: Trace): { nodes: Node[], edges: Edge[] } {
  if (!trace) {
    return { nodes: [], edges: [] };
  }

  const nodesMap = new Map<string, Node>();
  const edges: Edge[] = [];
  const g = new dagre.graphlib.Graph();

  g.setGraph({ rankdir: 'LR', nodesep: 100, ranksep: 200 });
  g.setDefaultEdgeLabel(() => ({}));

  // Add User Node (Trigger)
  const userId = 'user';
  nodesMap.set(userId, {
      id: userId,
      type: 'user',
      position: { x: 0, y: 0 },
      data: { label: 'User', role: 'Trigger' }
  });
  g.setNode(userId, { width: 180, height: 60 });

  // Add MCP Core Node (Always present as entry point)
  const coreId = 'core';
  nodesMap.set(coreId, {
      id: coreId,
      type: 'agent',
      position: { x: 0, y: 0 },
      data: { label: 'MCP Core', role: 'Orchestrator', status: 'Active' }
  });
  g.setNode(coreId, { width: 180, height: 60 });

  // Edge from User to Core
  edges.push({
      id: 'e-user-core',
      source: userId,
      target: coreId,
      animated: true,
      label: 'Request',
      markerEnd: { type: MarkerType.ArrowClosed },
      style: { stroke: '#2563eb' }
  });
  g.setEdge(userId, coreId);

  // Helper to create or get a node ID
  const getOrCreateNode = (span: Span): string => {
    // Map 'backend' service to Core
    if (span.serviceName === 'backend') {
        return coreId;
    }

    // Identify participants based on span type/serviceName
    let nodeId = span.serviceName ? `svc:${span.serviceName}` : `tool:${span.name}`;
    // Sanitize ID
    nodeId = nodeId.replace(/[^a-zA-Z0-9-_:]/g, "_");

    if (!nodesMap.has(nodeId)) {
      let type = 'service'; // Default
      let label = span.serviceName || span.name;
      let role = 'Service';

      if (span.type === 'tool') {
        type = 'tool';
        role = 'Tool';
      } else if (span.type === 'resource') {
        type = 'resource';
        role = 'Resource';
      } else if (span.type === 'core') {
        type = 'agent';
        role = 'Core Component';
      }

      const node: Node = {
        id: nodeId,
        type: type,
        position: { x: 0, y: 0 }, // Will be set by dagre
        data: { label, role }
      };

      nodesMap.set(nodeId, node);
      g.setNode(nodeId, { width: 180, height: 60 });
    }

    return nodeId;
  };

  // Traverse Spans
  const traverse = (span: Span, parentNodeId: string) => {
      let nodeId = getOrCreateNode(span);

      // Synthetic Tool Node logic
      // If we are at Core, and the span has input.name (tool execution), create a target tool node
      if (nodeId === coreId && span.input && typeof span.input.name === 'string') {
          const toolName = span.input.name;
          const toolNodeId = `tool:${toolName.replace(/[^a-zA-Z0-9-_:]/g, "_")}`;

          if (!nodesMap.has(toolNodeId)) {
              nodesMap.set(toolNodeId, {
                  id: toolNodeId,
                  type: 'tool',
                  position: { x: 0, y: 0 },
                  data: { label: toolName, role: 'Target Tool' }
              });
              g.setNode(toolNodeId, { width: 180, height: 60 });
          }

          // Edge from Core to Tool
          const edgeId = `e-${parentNodeId}-${toolNodeId}-${span.id}`;
          // Check uniqueness strictly?
          const exists = edges.some(e => e.id === edgeId);
          if (!exists) {
              edges.push({
                  id: edgeId,
                  source: parentNodeId, // Core
                  target: toolNodeId,
                  animated: true,
                  label: 'Execute',
                  markerEnd: { type: MarkerType.ArrowClosed },
              });
              g.setEdge(parentNodeId, toolNodeId);
          }

          nodeId = toolNodeId; // Continue traversal from Tool
      } else {
          // Standard edge logic
          if (nodeId !== parentNodeId) {
              const edgeId = `e-${parentNodeId}-${nodeId}-${span.id}`;
              const exists = edges.some(e => e.id === edgeId);
              if (!exists) {
                  edges.push({
                      id: edgeId,
                      source: parentNodeId,
                      target: nodeId,
                      animated: true, // Active flow
                      label: span.name, // Operation name
                      markerEnd: { type: MarkerType.ArrowClosed },
                  });
                  g.setEdge(parentNodeId, nodeId);
              }
          }
      }

      if (span.children) {
          span.children.forEach(child => traverse(child, nodeId));
      }
  };

  traverse(trace.rootSpan, coreId);

  // Calculate Layout
  dagre.layout(g);

  // Apply positions
  const nodes = Array.from(nodesMap.values()).map(node => {
      const nodeWithPos = g.node(node.id);
      return {
          ...node,
          position: {
              x: nodeWithPos.x - nodeWithPos.width / 2,
              y: nodeWithPos.y - nodeWithPos.height / 2
          }
      };
  });

  return { nodes, edges };
}
