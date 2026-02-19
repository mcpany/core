/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import dagre from 'dagre';
import { Node, Edge, Position } from '@xyflow/react';
import { Trace, Span } from '@/types/trace';

const nodeWidth = 180;
const nodeHeight = 60;

/**
 * Transforms a trace into layouted nodes and edges for React Flow.
 * @param trace The trace data to visualize.
 * @param direction The direction of the layout ('TB' or 'LR').
 * @returns An object containing nodes and edges.
 */
export const getLayoutedElements = (trace: Trace, direction = 'TB'): { nodes: Node[], edges: Edge[] } => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  dagreGraph.setGraph({ rankdir: direction });

  const nodesMap = new Map<string, Node>();
  const edgesMap = new Map<string, Edge>();

  // Helper to ensure unique node IDs
  const getParticipantId = (span: Span): string => {
      let id = span.serviceName ? `svc:${span.serviceName}` : `tool:${span.name}`;
      if (span.type === 'core') return 'core';
      if (span.type === 'resource') return `res:${span.name}`;

      // Sanitize ID
      id = id.replace(/[^a-zA-Z0-9-_:]/g, "_");
      return id;
  };

  const getParticipantLabel = (span: Span): string => {
      if (span.type === 'core') return 'MCP Core';
      if (span.type === 'service') return span.serviceName || span.name;
      return span.name;
  };

  const getParticipantType = (span: Span): string => {
      if (span.type === 'core') return 'agent'; // Use AgentNode for Core
      if (span.type === 'service') return 'service';
      if (span.type === 'resource') return 'resource';
      return 'tool';
  };

  // Add User Node
  const userNodeId = 'user';
  if (!nodesMap.has(userNodeId)) {
      const userNode: Node = {
          id: userNodeId,
          type: 'user',
          data: { label: 'Client' },
          position: { x: 0, y: 0 },
      };
      nodesMap.set(userNodeId, userNode);
      dagreGraph.setNode(userNodeId, { width: nodeWidth, height: nodeHeight });
  }

  // Add Core Node (implied recipient of initial request)
  const coreNodeId = 'core';
  if (!nodesMap.has(coreNodeId)) {
      const coreNode: Node = {
          id: coreNodeId,
          type: 'agent', // Visualization uses 'agent' for Core
          data: { label: 'MCP Core', role: 'Gateway', status: trace.status === 'error' ? 'Error' : 'Active' },
          position: { x: 0, y: 0 },
      };
      nodesMap.set(coreNodeId, coreNode);
      dagreGraph.setNode(coreNodeId, { width: nodeWidth, height: nodeHeight });
  }

  // Edge: User -> Core
  const initialEdgeId = `e-user-core`;
  if (!edgesMap.has(initialEdgeId)) {
      edgesMap.set(initialEdgeId, {
          id: initialEdgeId,
          source: userNodeId,
          target: coreNodeId,
          animated: true,
          label: 'Request',
      });
      dagreGraph.setEdge(userNodeId, coreNodeId);
  }

  // Traverse Spans
  const traverse = (span: Span, parentId: string) => {
      const nodeId = getParticipantId(span);

      // If the node is 'core' (self-call), we skip creating a new node but still track the edge if needed?
      // Usually spans represent calls to *other* things.
      // If span.type is 'tool' or 'service', it's a distinct node.

      if (!nodesMap.has(nodeId)) {
          const node: Node = {
              id: nodeId,
              type: getParticipantType(span),
              data: { label: getParticipantLabel(span) },
              position: { x: 0, y: 0 },
          };
          nodesMap.set(nodeId, node);
          dagreGraph.setNode(nodeId, { width: nodeWidth, height: nodeHeight });
      }

      // Edge: Parent -> Current
      // If parent is same as current, it's a self-call (loop), dagre handles it but might look messy.
      if (parentId !== nodeId) {
          const edgeId = `e-${parentId}-${nodeId}`;
          if (!edgesMap.has(edgeId)) {
               edgesMap.set(edgeId, {
                  id: edgeId,
                  source: parentId,
                  target: nodeId,
                  animated: true,
                  label: span.name.length > 20 ? span.name.substring(0, 17) + '...' : span.name,
              });
              dagreGraph.setEdge(parentId, nodeId);
          }
      }

      // Children
      if (span.children) {
          span.children.forEach(child => traverse(child, nodeId));
      }
  };

  // Start traversal from root span.
  // Root span is usually executed by Core.
  // So we traverse root's children as calls FROM Core.
  // Wait, if root span IS the request to core, then its children are what core does.

  if (trace.rootSpan.children) {
      trace.rootSpan.children.forEach(child => traverse(child, coreNodeId));
  }

  dagre.layout(dagreGraph);

  const nodes = Array.from(nodesMap.values()).map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      targetPosition: direction === 'LR' ? Position.Left : Position.Top,
      sourcePosition: direction === 'LR' ? Position.Right : Position.Bottom,
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    };
  });

  return { nodes, edges: Array.from(edgesMap.values()) };
};
