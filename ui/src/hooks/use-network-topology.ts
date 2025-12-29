/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useCallback } from 'react';
import { Node, Edge, useNodesState, useEdgesState, addEdge, Connection } from '@xyflow/react';

export type NodeType = 'mcp-server' | 'upstream' | 'agent';

export interface NodeData {
    label: string;
    type: NodeType;
    status: 'active' | 'inactive' | 'error';
    details?: Record<string, unknown>;
    [key: string]: unknown;
}

export interface NetworkGraphState {
    nodes: Node<NodeData>[];
    edges: Edge[];
    onNodesChange: any;
    onEdgesChange: any;
    onConnect: (params: Connection) => void;
    refreshTopology: () => void;
    autoLayout: () => void;
}

const initialNodes: Node<NodeData>[] = [
  {
      id: 'mcp-core',
      position: { x: 400, y: 300 },
      data: { label: 'MCP Any Core', type: 'mcp-server', status: 'active' },
      type: 'input',
      style: { background: '#fff', border: '2px solid #000', borderRadius: '8px', padding: '10px', width: 150, textAlign: 'center', fontWeight: 'bold' }
  },
  {
      id: 'fs-1',
      position: { x: 100, y: 100 },
      data: { label: 'Filesystem', type: 'upstream', status: 'active' },
      style: { background: '#f0f9ff', border: '1px solid #0ea5e9', borderRadius: '8px', padding: '10px' }
  },
  {
      id: 'linear-1',
      position: { x: 100, y: 500 },
      data: { label: 'Linear', type: 'upstream', status: 'active' },
      style: { background: '#f0f9ff', border: '1px solid #0ea5e9', borderRadius: '8px', padding: '10px' }
  },
  {
      id: 'agent-claude',
      position: { x: 700, y: 200 },
      data: { label: 'Claude Desktop', type: 'agent', status: 'active' },
      type: 'output',
      style: { background: '#f0fdf4', border: '1px solid #22c55e', borderRadius: '8px', padding: '10px' }
  },
  {
      id: 'agent-cursor',
      position: { x: 700, y: 400 },
      data: { label: 'Cursor', type: 'agent', status: 'inactive' },
      type: 'output',
      style: { background: '#fef2f2', border: '1px solid #ef4444', borderRadius: '8px', padding: '10px' }
  },
];

const initialEdges: Edge[] = [
  { id: 'e1-core', source: 'fs-1', target: 'mcp-core', animated: true },
  { id: 'e2-core', source: 'linear-1', target: 'mcp-core', animated: true },
  { id: 'core-a1', source: 'mcp-core', target: 'agent-claude', animated: true },
  { id: 'core-a2', source: 'mcp-core', target: 'agent-cursor', animated: false, style: { strokeDasharray: 5 } },
];

export function useNetworkTopology() {
    const [nodes, setNodes, onNodesChange] = useNodesState<Node<NodeData>>(initialNodes);
    const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

    const onConnect = useCallback(
        (params: Connection) => setEdges((eds) => addEdge(params, eds)),
        [setEdges],
    );

    const refreshTopology = useCallback(() => {
        // Jiggle nodes slightly
        setNodes((nds) => nds.map(n => ({
            ...n,
            position: {
                x: n.position.x + (Math.random() - 0.5) * 20,
                y: n.position.y + (Math.random() - 0.5) * 20,
            }
        })));
    }, [setNodes]);

    const autoLayout = useCallback(() => {
        // Simple mock auto-layout (reset to initial positions but with slight variation)
        setNodes((nds) => nds.map(n => {
            const initial = initialNodes.find(i => i.id === n.id);
            if (initial) {
                return {
                    ...n,
                    position: initial.position
                };
            }
            return n;
        }));
    }, [setNodes]);

    return {
        nodes,
        edges,
        onNodesChange,
        onEdgesChange,
        onConnect,
        refreshTopology,
        autoLayout
    };
}
