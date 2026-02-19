/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import dagre from 'dagre';
import { Trace, Span } from '@/types/trace';
import { Node, Edge, MarkerType } from '@xyflow/react';

/**
 * Converts a Trace object into ReactFlow nodes and edges with automatic layout.
 * @param trace - The trace data.
 * @returns An object containing nodes and edges.
 */
export function convertTraceToGraph(trace: Trace): { nodes: Node[], edges: Edge[] } {
  const g = new dagre.graphlib.Graph();
  g.setGraph({ rankdir: 'TB', nodesep: 100, ranksep: 100 });
  g.setDefaultEdgeLabel(() => ({}));

  const nodesMap = new Map<string, Node>();
  const edgesMap = new Map<string, Edge>();

  // Helper to get Node ID from Span
  const getParticipantId = (span: Span): string => {
    // Treat the root span as the Core entry point
    if (span.id === trace.rootSpan.id) return 'core';

    if (span.type === 'core') return 'core';
    if (span.type === 'tool') return `tool-${span.name}`;
    if (span.type === 'service') return `service-${span.serviceName || span.name}`;
    if (span.type === 'resource') return `resource-${span.name}`;
    // Unique ID fallback for unknown types to prevent collisions
    return `unknown-${span.id}`;
  };

  const getParticipantLabel = (span: Span): string => {
      // Treat the root span as the Core entry point
      if (span.id === trace.rootSpan.id) return 'MCP Core';

      if (span.type === 'core') return 'MCP Core';
      if (span.type === 'service') return span.serviceName || span.name;
      return span.name;
  };

  const createNode = (id: string, type: string, label: string, status?: string) => {
      if (!nodesMap.has(id)) {
          // Map internal types to custom node types
          let nodeType = 'agent'; // Default for core

          // Force root/core to be agent type
          if (id === 'core' || label === 'MCP Core') {
              nodeType = 'agent';
          } else if (type === 'tool') {
              nodeType = 'tool';
          } else if (type === 'service') {
              nodeType = 'service';
          } else if (type === 'resource') {
              nodeType = 'resource';
          } else if (type === 'core') {
              nodeType = 'agent';
          }

          nodesMap.set(id, {
              id,
              type: nodeType,
              position: { x: 0, y: 0 }, // Will be set by dagre
              data: {
                  label,
                  status: status === 'pending' ? 'Thinking...' : undefined,
                  role: type === 'core' ? 'Orchestrator' : undefined
              }
          });
          g.setNode(id, { width: 180, height: 80 });
      }
  };

  // Add User Node (Trigger)
  const userId = 'user';
  nodesMap.set(userId, {
      id: userId,
      type: 'user',
      position: { x: 0, y: 0 },
      data: { label: 'User' }
  });
  g.setNode(userId, { width: 180, height: 80 });

  // Recursive traversal
  const traverse = (span: Span, parentId: string) => {
      const currentId = getParticipantId(span);
      createNode(currentId, span.type, getParticipantLabel(span), span.status);

      if (parentId !== currentId) {
           const edgeId = `${parentId}-${currentId}`;
           if (!edgesMap.has(edgeId)) {
                edgesMap.set(edgeId, {
                    id: edgeId,
                    source: parentId,
                    target: currentId,
                    animated: true,
                    markerEnd: { type: MarkerType.ArrowClosed },
                    style: { strokeWidth: 2 }
                });
                g.setEdge(parentId, currentId);
           }
      }

      // Special handling for execute endpoint to show the tool being called
      // This is needed because the backend might not expose the internal tool execution span in the public trace,
      // but we can infer it from the API payload.
      if (span.name.includes('/api/v1/execute') && span.input?.name) {
        const toolName = span.input.name;
        const toolId = `tool-${toolName}`;
        createNode(toolId, 'tool', toolName, span.status);

        // Edge from Core (currentId) to Tool
        const edgeId = `${currentId}-${toolId}`;
        if (!edgesMap.has(edgeId)) {
              edgesMap.set(edgeId, {
                  id: edgeId,
                  source: currentId,
                  target: toolId,
                  animated: true,
                  markerEnd: { type: MarkerType.ArrowClosed },
                  style: { strokeWidth: 2, strokeDasharray: 4 } // Dashed for virtual/inferred calls
              });
              g.setEdge(currentId, toolId);
        }
    }

      if (span.children) {
          span.children.forEach(child => traverse(child, currentId));
      }
  };

  // Start traversal
  traverse(trace.rootSpan, userId);

  // Calculate layout
  dagre.layout(g);

  // Apply layout positions to nodes
  nodesMap.forEach((node) => {
      const nodeWithPos = g.node(node.id);
      node.position = {
          x: nodeWithPos.x - nodeWithPos.width / 2,
          y: nodeWithPos.y - nodeWithPos.height / 2
      };
  });

  return {
      nodes: Array.from(nodesMap.values()),
      edges: Array.from(edgesMap.values())
  };
}
