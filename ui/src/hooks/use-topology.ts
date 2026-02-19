/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback } from 'react';
import { Node, Edge, useNodesState, useEdgesState } from '@xyflow/react';
import dagre from 'dagre';
import { apiClient } from '@/lib/client';

// Types matching the Proto definition (mapped to JSON)
export interface TopologyNode {
    id: string;
    label: string;
    type: NodeType;
    status: NodeStatus;
    metadata?: Record<string, string>;
    children?: TopologyNode[];
    metrics?: NodeMetrics;
}

export enum NodeType {
    NODE_TYPE_UNSPECIFIED = 0,
    NODE_TYPE_CLIENT = 1,
    NODE_TYPE_CORE = 2,
    NODE_TYPE_SERVICE = 3,
    NODE_TYPE_TOOL = 4,
    NODE_TYPE_RESOURCE = 5,
    NODE_TYPE_PROMPT = 6,
    NODE_TYPE_API_CALL = 7,
    NODE_TYPE_MIDDLEWARE = 8,
    NODE_TYPE_WEBHOOK = 9,
}

export enum NodeStatus {
    NODE_STATUS_UNSPECIFIED = 0,
    NODE_STATUS_ACTIVE = 1,
    NODE_STATUS_INACTIVE = 2,
    NODE_STATUS_ERROR = 3,
}

export interface NodeMetrics {
    qps: number;
    latencyMs: number;
    errorRate: number;
}

export interface TopologyGraph {
    clients: TopologyNode[];
    core: TopologyNode;
}

const nodeWidth = 180;
const nodeHeight = 50;

const getLayoutedElements = (nodes: Node[], edges: Edge[]) => {
    const dagreGraph = new dagre.graphlib.Graph();
    dagreGraph.setDefaultEdgeLabel(() => ({}));

    dagreGraph.setGraph({ rankdir: 'LR' });

    nodes.forEach((node) => {
        dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
    });

    edges.forEach((edge) => {
        dagreGraph.setEdge(edge.source, edge.target);
    });

    dagre.layout(dagreGraph);

    nodes.forEach((node) => {
        const nodeWithPosition = dagreGraph.node(node.id);
        node.position = {
            x: nodeWithPosition.x - nodeWidth / 2,
            y: nodeWithPosition.y - nodeHeight / 2,
        };
    });

    return { nodes, edges };
};

export function useTopology(isLive: boolean = false) {
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchTopology = useCallback(async () => {
        try {
            const data: TopologyGraph = await apiClient.getTopology();

            const newNodes: Node[] = [];
            const newEdges: Edge[] = [];

            // 1. Core Node
            const coreNode = data.core;
            if (coreNode) {
                newNodes.push({
                    id: coreNode.id,
                    type: 'service', // Using generic type for now, or map to specific
                    data: { label: coreNode.label, status: 'active', role: 'Gateway' },
                    position: { x: 0, y: 0 },
                });

                // Process Core Children (Services, Middleware)
                if (coreNode.children) {
                    coreNode.children.forEach(child => {
                        newNodes.push({
                            id: child.id,
                            type: 'service',
                            data: {
                                label: child.label,
                                status: child.status === NodeStatus.NODE_STATUS_ACTIVE ? 'active' : 'inactive',
                                metrics: child.metrics
                            },
                            position: { x: 0, y: 0 },
                        });
                        newEdges.push({
                            id: `e-${coreNode.id}-${child.id}`,
                            source: coreNode.id,
                            target: child.id,
                            animated: true,
                            style: { stroke: '#22c55e' }
                        });

                        // Process Service Children (Tools)
                        if (child.children) {
                            child.children.forEach(tool => {
                                newNodes.push({
                                    id: tool.id,
                                    type: 'tool',
                                    data: { label: tool.label },
                                    position: { x: 0, y: 0 },
                                });
                                newEdges.push({
                                    id: `e-${child.id}-${tool.id}`,
                                    source: child.id,
                                    target: tool.id,
                                    type: 'smoothstep',
                                });
                            });
                        }
                    });
                }
            }

            // 2. Clients
            if (data.clients) {
                data.clients.forEach(client => {
                    newNodes.push({
                        id: client.id,
                        type: 'user',
                        data: { label: client.label || 'Client' },
                        position: { x: 0, y: 0 },
                    });
                    // Connect Client to Core
                    if (coreNode) {
                        newEdges.push({
                            id: `e-${client.id}-${coreNode.id}`,
                            source: client.id,
                            target: coreNode.id,
                            animated: true,
                            label: 'Session',
                        });
                    }
                });
            }

            // Apply Layout
            const layouted = getLayoutedElements(newNodes, newEdges);
            setNodes(layouted.nodes);
            setEdges(layouted.edges);
            setError(null);
        } catch (err: any) {
            console.error("Failed to fetch topology", err);
            setError(err.message || "Failed to load topology");
        } finally {
            setLoading(false);
        }
    }, [setNodes, setEdges]);

    useEffect(() => {
        fetchTopology();

        let interval: NodeJS.Timeout;
        if (isLive) {
            interval = setInterval(fetchTopology, 5000);
        }
        return () => clearInterval(interval);
    }, [isLive, fetchTopology]);

    return { nodes, edges, onNodesChange, onEdgesChange, loading, error, refresh: fetchTopology };
}
