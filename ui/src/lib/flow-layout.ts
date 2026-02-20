/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import dagre from 'dagre';
import { Node, Edge, MarkerType } from '@xyflow/react';
import { Trace, Span } from '@/types/trace';

interface GraphResult {
  nodes: Node[];
  edges: Edge[];
}

/**
 * Transforms a Trace into a ReactFlow graph layout.
 *
 * @param trace - The trace to visualize.
 * @returns The nodes and edges for the graph.
 */
export function traceToGraph(trace: Trace): GraphResult {
  const g = new dagre.graphlib.Graph();
  g.setGraph({ rankdir: 'LR', align: 'UL', ranksep: 100, nodesep: 50 });
  g.setDefaultEdgeLabel(() => ({}));

  const nodesMap = new Map<string, Node>();
  const edges: Edge[] = [];
  const processedEdges = new Set<string>();

  // Helper to get node ID
  const getNodeId = (role: string, name?: string) => {
    if (role === 'user') return 'user';
    if (role === 'core') return 'core';
    // Sanitize name for ID
    const sanitized = (name || 'unknown').replace(/[^a-zA-Z0-9-_]/g, '_');
    return `${role}-${sanitized}`;
  };

  // Add User Node
  const userId = getNodeId('user');
  if (!nodesMap.has(userId)) {
      const userNode: Node = {
        id: userId,
        type: 'user',
        data: { label: 'User' },
        position: { x: 0, y: 0 }, // Dagre will set this
      };
      nodesMap.set(userId, userNode);
      g.setNode(userId, { width: 150, height: 60 });
  }

  // Add Core Node
  const coreId = getNodeId('core');
  if (!nodesMap.has(coreId)) {
      const coreNode: Node = {
        id: coreId,
        type: 'agent', // Using 'agent' type for Core as it orchestrates
        data: { label: 'MCP Core', role: 'Gateway' },
        position: { x: 0, y: 0 },
      };
      nodesMap.set(coreId, coreNode);
      g.setNode(coreId, { width: 150, height: 80 });
  }

  // Add Edge User -> Core
  const userEdgeId = `e-${userId}-${coreId}`;
  if (!processedEdges.has(userEdgeId)) {
      edges.push({
          id: userEdgeId,
          source: userId,
          target: coreId,
          animated: true,
          label: 'Request',
          markerEnd: { type: MarkerType.ArrowClosed },
      });
      processedEdges.add(userEdgeId);
      g.setEdge(userId, coreId);
  }

  // Recursive traversal
  const traverse = (span: Span, parentId: string) => {
    let currentId = parentId;

    // Identify if this span represents a distinct node (Service/Tool/Resource)
    // or if it's just an internal span.
    // In MCP Any, the Core calls Tools/Services.
    // So if span.type is tool/service/resource, it's a target node.

    if (span.type === 'tool' || span.type === 'service' || span.type === 'resource') {
        const type = span.type; // 'tool' | 'service' | 'resource'
        const label = span.name;
        // Use serviceName if available for grouping?
        // For now, let's treat each tool as a node.
        // Or if serviceName is present, maybe use that as the node?
        // The previous SequenceDiagram logic used `svc:{serviceName}` or `tool:{name}`.
        // Let's use the span type as node type.

        currentId = getNodeId(type, label);

        if (!nodesMap.has(currentId)) {
            const node: Node = {
                id: currentId,
                type: type,
                data: {
                    label: label,
                    status: span.status === 'error' ? 'Error' : undefined
                },
                position: { x: 0, y: 0 },
            };
            nodesMap.set(currentId, node);
            g.setNode(currentId, { width: 150, height: 60 });
        }

        // Create Edge from Parent -> Current
        // Avoid duplicate edges if multiple calls happen between same nodes
        // But for visualizer, multiple calls might be interesting?
        // Let's keep it simple: one edge per relationship type.
        // Actually, let's allow multiple edges if they are distinct calls?
        // ReactFlow handles multiple edges between nodes if ids are unique.
        // But layout might get messy. Let's use unique edge ID based on span ID.
        // Wait, if I call Tool A 10 times, I don't want 10 lines.
        // I want one line with "10 calls" or just one line.
        // Let's stick to unique edges for now to see the "flow".
        // Or better: unique edge per (Source, Target).

        const edgeKey = `e-${parentId}-${currentId}`;
        if (!processedEdges.has(edgeKey)) {
             edges.push({
                id: edgeKey,
                source: parentId,
                target: currentId,
                animated: true,
                label: span.type === 'tool' ? 'Calls' : 'Accesses',
                markerEnd: { type: MarkerType.ArrowClosed },
            });
            processedEdges.add(edgeKey);
            g.setEdge(parentId, currentId);
        }
    }

    // Continue traversal
    if (span.children) {
        span.children.forEach(child => traverse(child, currentId));
    }
  };

  // Start traversal from Root Span children (since Root is usually the request to Core)
  // The Root Span itself represents the request entering the system.
  // We already added User -> Core.
  // Now we traverse from Core -> Children.
  if (trace.rootSpan.children) {
      trace.rootSpan.children.forEach(child => traverse(child, coreId));
  }

  // Layout
  dagre.layout(g);

  // Apply positions
  const nodes = Array.from(nodesMap.values()).map(node => {
      const nodeWithPos = g.node(node.id);
      return {
          ...node,
          position: {
              x: nodeWithPos.x - nodeWithPos.width / 2,
              y: nodeWithPos.y - nodeWithPos.height / 2,
          },
      };
  });

  return { nodes, edges };
}
