/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useCallback, useEffect } from 'react';
import { Node, Edge, useNodesState, useEdgesState, addEdge, Connection, MarkerType } from '@xyflow/react';
import { apiClient } from '@/lib/client';

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

// Static nodes (Agents + Core)
const staticNodes: Node<NodeData>[] = [
  // Center: MCP Any
  {
      id: 'mcp-core',
      position: { x: 400, y: 300 },
      data: { label: 'MCP Any', type: 'mcp-server', status: 'active' },
      type: 'default',
      style: { background: '#fff', border: '2px solid #000', borderRadius: '8px', padding: '10px', width: 150, textAlign: 'center', fontWeight: 'bold' }
  },
  // Left: Agents
  {
      id: 'agent-claude',
      position: { x: 100, y: 200 },
      data: { label: 'Claude Desktop', type: 'agent', status: 'active' },
      type: 'input',
      style: { background: '#f0fdf4', border: '1px solid #22c55e', borderRadius: '8px', padding: '10px' }
  },
  {
      id: 'agent-cursor',
      position: { x: 100, y: 400 },
      data: { label: 'Cursor', type: 'agent', status: 'inactive' },
      type: 'input',
      style: { background: '#fef2f2', border: '1px solid #ef4444', borderRadius: '8px', padding: '10px' }
  },
];

const staticEdges: Edge[] = [
  // Agents -> MCP Any
  { id: 'e-claude-core', source: 'agent-claude', target: 'mcp-core', animated: true, markerEnd: { type: MarkerType.ArrowClosed } },
  { id: 'e-cursor-core', source: 'agent-cursor', target: 'mcp-core', animated: false, style: { strokeDasharray: 5, stroke: '#9ca3af' }, markerEnd: { type: MarkerType.ArrowClosed, color: '#9ca3af' } },
];

export function useNetworkTopology() {
    const [nodes, setNodes, onNodesChange] = useNodesState<Node<NodeData>>(staticNodes);
    const [edges, setEdges, onEdgesChange] = useEdgesState(staticEdges);

    const fetchData = useCallback(async () => {
        try {
            const servicesData = await apiClient.listServices();
            const toolsData = await apiClient.listTools();

            // Map services to nodes
            const serviceNodes: Node<NodeData>[] = (servicesData.services || []).map((svc: any, index: number) => ({
                id: `svc-${svc.id || svc.name}`,
                position: { x: 700, y: 100 + (index * 150) }, // Simple vertical stacking for now
                data: {
                    label: svc.name,
                    type: 'upstream',
                    status: svc.disable ? 'inactive' : 'active',
                    details: svc
                },
                type: 'output',
                style: {
                    background: svc.disable ? '#f3f4f6' : '#f0f9ff',
                    border: svc.disable ? '1px solid #9ca3af' : '1px solid #0ea5e9',
                    borderRadius: '8px',
                    padding: '10px'
                }
            }));

            // Map services to edges (MCP Any -> Service)
            const serviceEdges: Edge[] = (servicesData.services || []).map((svc: any) => ({
                id: `e-core-${svc.id || svc.name}`,
                source: 'mcp-core',
                target: `svc-${svc.id || svc.name}`,
                animated: !svc.disable,
                style: svc.disable ? { stroke: '#9ca3af', strokeDasharray: 5 } : undefined,
                markerEnd: { type: MarkerType.ArrowClosed, color: svc.disable ? '#9ca3af' : undefined }
            }));

            setNodes((prevNodes) => {
                // Merge static nodes with fetched service nodes
                // Preserve positions of existing nodes if they exist?
                // For now, simpler to just replace upstream nodes but keep static ones.
                // Or better: Re-create the full list to ensure freshness.
                // To avoid jumping, we could check if id exists and keep position, but "autoLayout" exists for that.

                // Let's just append/replace.
                return [...staticNodes, ...serviceNodes];
            });

            setEdges([...staticEdges, ...serviceEdges]);

        } catch (error) {
            console.error("Failed to fetch topology data:", error);
        }
    }, [setNodes, setEdges]);

    // Fetch on mount
    useEffect(() => {
        fetchData();
        // Poll every 5 seconds
        const interval = setInterval(fetchData, 5000);
        return () => clearInterval(interval);
    }, [fetchData]);

    const onConnect = useCallback(
        (params: Connection) => setEdges((eds) => addEdge(params, eds)),
        [setEdges],
    );

    const refreshTopology = useCallback(() => {
        fetchData();
    }, [fetchData]);

    const autoLayout = useCallback(() => {
         fetchData(); // Reset to default positions defined in fetchData
    }, [fetchData]);

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
