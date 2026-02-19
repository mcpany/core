/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback } from 'react';
import { Node, Edge, useNodesState, useEdgesState } from '@xyflow/react';
import { apiClient } from '@/lib/client';
import dagre from 'dagre';

// Types from Proto
interface ProtoNode {
    id: string;
    label: string;
    type: string; // Enum as string or int? usually string in JSON if emit defaults
    status: string;
    metadata?: Record<string, string>;
    children?: ProtoNode[];
    metrics?: {
        qps?: number;
        error_rate?: number;
        latency_ms?: number;
    };
}

interface ProtoGraph {
    clients: ProtoNode[];
    core: ProtoNode;
}

const nodeWidth = 200;
const nodeHeight = 80;

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

    const layoutedNodes = nodes.map((node) => {
        const nodeWithPosition = dagreGraph.node(node.id);
        return {
            ...node,
            position: {
                x: nodeWithPosition.x - nodeWidth / 2,
                y: nodeWithPosition.y - nodeHeight / 2,
            },
        };
    });

    return { nodes: layoutedNodes, edges };
};

/**
 * Hook to fetch and manage the network topology graph.
 * @returns The topology state and controls.
 */
export function useTopology() {
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);
    const [loading, setLoading] = useState(false);
    const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
    const [isLive, setIsLive] = useState(false);

    const fetchData = useCallback(async () => {
        try {
            const data: ProtoGraph = await apiClient.getTopology();

            const rawNodes: Node[] = [];
            const rawEdges: Edge[] = [];

            // Helper to map type string to UI node type
            const mapType = (t: string | any) => {
                // Proto enums might be strings like "NODE_TYPE_SERVICE" or integers.
                // Assuming strings based on protojson default behavior for enums usually.
                // But let's handle loose matching.
                const typeStr = String(t).toUpperCase();
                if (typeStr.includes('CLIENT')) return 'user'; // UserNode
                if (typeStr.includes('SERVICE')) return 'service'; // ServiceNode
                if (typeStr.includes('TOOL')) return 'tool'; // ToolNode
                if (typeStr.includes('CORE')) return 'agent'; // AgentNode (Core/Gateway)
                if (typeStr.includes('RESOURCE')) return 'resource';
                return 'default';
            };

            // Process Core
            if (data.core) {
                rawNodes.push({
                    id: data.core.id,
                    type: 'agent', // Core is central agent
                    data: {
                        label: data.core.label,
                        role: 'Gateway',
                        status: 'Active'
                    },
                    position: { x: 0, y: 0 },
                });

                // Process Services (Children of Core)
                if (data.core.children) {
                    data.core.children.forEach(svc => {
                        // Middleware pipeline?
                        if (svc.type && String(svc.type).includes('MIDDLEWARE')) {
                            // Skip or visualize differently? Let's skip for cleaner view for now
                            return;
                        }

                        rawNodes.push({
                            id: svc.id,
                            type: mapType(svc.type),
                            data: {
                                label: svc.label,
                                status: svc.metrics ? `${svc.metrics.qps?.toFixed(1)} QPS` : undefined
                            },
                            position: { x: 0, y: 0 },
                        });

                        rawEdges.push({
                            id: `e-${data.core.id}-${svc.id}`,
                            source: data.core.id,
                            target: svc.id,
                            animated: true,
                            style: { stroke: '#64748b' }
                        });

                        // Tools inside Service
                        if (svc.children) {
                            svc.children.forEach(tool => {
                                rawNodes.push({
                                    id: tool.id,
                                    type: mapType(tool.type),
                                    data: { label: tool.label },
                                    position: { x: 0, y: 0 },
                                });
                                rawEdges.push({
                                    id: `e-${svc.id}-${tool.id}`,
                                    source: svc.id,
                                    target: tool.id,
                                    type: 'smoothstep'
                                });
                            });
                        }
                    });
                }
            }

            // Process Clients
            if (data.clients) {
                data.clients.forEach(client => {
                    rawNodes.push({
                        id: client.id,
                        type: 'user',
                        data: { label: client.label || 'Client' },
                        position: { x: 0, y: 0 },
                    });

                    // Connect Client -> Core
                    if (data.core) {
                        rawEdges.push({
                            id: `e-${client.id}-${data.core.id}`,
                            source: client.id,
                            target: data.core.id,
                            animated: true,
                            style: { stroke: '#22c55e' } // Green for active traffic
                        });
                    }
                });
            }

            // Apply Layout
            const layout = getLayoutedElements(rawNodes, rawEdges);
            setNodes(layout.nodes);
            setEdges(layout.edges);
            setLastUpdated(new Date());

        } catch (e) {
            console.error("Failed to fetch topology", e);
        }
    }, [setNodes, setEdges]);

    useEffect(() => {
        fetchData();

        let interval: NodeJS.Timeout;
        if (isLive) {
            interval = setInterval(fetchData, 2000);
        }
        return () => clearInterval(interval);
    }, [fetchData, isLive]);

    return {
        nodes,
        edges,
        onNodesChange,
        onEdgesChange,
        isLive,
        setIsLive,
        refresh: fetchData,
        lastUpdated
    };
}
