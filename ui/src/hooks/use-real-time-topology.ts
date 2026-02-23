/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import { Node, Edge, useNodesState, useEdgesState } from '@xyflow/react';
import { apiClient } from '@/lib/client';
import * as dagre from 'dagre';

// Types from Proto
interface ProtoNode {
    id: string;
    label: string;
    type: string;
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
export function useRealTimeTopology() {
    const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
    const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
    const [isLive, setIsLive] = useState(false);

    // ⚡ BOLT: Refs to store previous state for memoization
    const prevStructureHash = useRef<string>("");
    const nodesRef = useRef<Node[]>([]);

    // Update nodesRef whenever nodes change so fetchData can access the latest positions
    useEffect(() => {
        nodesRef.current = nodes;
    }, [nodes]);

    const fetchData = useCallback(async () => {
        try {
            const data: ProtoGraph = await apiClient.getTopology();

            const rawNodes: Node[] = [];
            const rawEdges: Edge[] = [];

            // Helper to map type string to UI node type
            const mapType = (t: string | any) => {
                const typeStr = String(t).toUpperCase();
                if (typeStr.includes('CLIENT')) return 'user';
                if (typeStr.includes('SERVICE')) return 'service';
                if (typeStr.includes('TOOL')) return 'tool';
                if (typeStr.includes('CORE')) return 'agent';
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
                        status: 'Active',
                        metrics: { qps: 0, errorRate: 0, latencyMs: 0 }
                    },
                    position: { x: 0, y: 0 },
                });

                // Process Services (Children of Core)
                if (data.core.children) {
                    data.core.children.forEach(svc => {
                        if (svc.type && String(svc.type).includes('MIDDLEWARE')) {
                            return;
                        }

                        const metrics = svc.metrics || { qps: 0, error_rate: 0, latency_ms: 0 };
                        const qps = metrics.qps || 0;
                        const errorRate = metrics.error_rate || 0;
                        const latencyMs = metrics.latency_ms || 0;

                        rawNodes.push({
                            id: svc.id,
                            type: mapType(svc.type),
                            data: {
                                label: svc.label,
                                status: qps > 0 ? `${qps.toFixed(1)} QPS` : undefined,
                                metrics: { qps, errorRate, latencyMs }
                            },
                            position: { x: 0, y: 0 },
                        });

                        // Edge Core -> Service inherits service metrics (traffic flowing TO service)
                        rawEdges.push({
                            id: `e-${data.core.id}-${svc.id}`,
                            source: data.core.id,
                            target: svc.id,
                            type: 'traffic', // Use custom traffic edge
                            data: { qps, errorRate, latencyMs },
                            animated: false,
                        });

                        // Tools inside Service
                        if (svc.children) {
                            svc.children.forEach(tool => {
                                // Tools don't have individual metrics in current proto usually,
                                // but if they did, we would use them.
                                // For now, assume a fraction of service traffic or 0.
                                // Ideally backend provides per-tool metrics.
                                // But `GetGraph` in `topology.go` doesn't set metrics on Tool nodes yet.
                                // We'll leave them as default edges or inherit if needed.
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
                                    type: 'smoothstep', // Standard edge for internal hierarchy
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
                    // Client traffic is aggregated at Core? Or do we know per-client?
                    // Currently `GetGraph` doesn't provide per-client metrics.
                    // We can estimate or just show "Active" (green) if any traffic.
                    if (data.core) {
                        rawEdges.push({
                            id: `e-${client.id}-${data.core.id}`,
                            source: client.id,
                            target: data.core.id,
                            type: 'traffic',
                            data: { qps: 1, errorRate: 0 }, // Fake minimal traffic to show connection
                            animated: false,
                        });
                    }
                });
            }

            // ⚡ BOLT: Calculate structure hash to avoid unnecessary re-layouts
            const structureHash = rawNodes.map(n => n.id).sort().join(',') + '|' + rawEdges.map(e => e.id).sort().join(',');

            if (structureHash === prevStructureHash.current && nodesRef.current.length > 0) {
                // Structure matches, only update data (metrics) and preserve positions
                const currentNodesMap = new Map(nodesRef.current.map(n => [n.id, n]));

                const updatedNodes = rawNodes.map(node => {
                    const existing = currentNodesMap.get(node.id);
                    if (existing) {
                        return {
                            ...node,
                            position: existing.position
                        };
                    }
                    return node;
                });

                setNodes(updatedNodes);
                setEdges(rawEdges); // Update edges to propagate new metric data
            } else {
                // Structure changed or first load, run full layout
                const layout = getLayoutedElements(rawNodes, rawEdges);
                setNodes(layout.nodes);
                setEdges(layout.edges);
                prevStructureHash.current = structureHash;
            }

            setLastUpdated(new Date());

        } catch (e) {
            console.error("Failed to fetch topology", e);
        }
    }, [setNodes, setEdges]);

    // Initial fetch on mount
    useEffect(() => {
        fetchData();
    }, [fetchData]);

    // Polling logic
    useEffect(() => {
        if (!isLive) return;

        let timeoutId: NodeJS.Timeout;
        let isMounted = true;

        const loop = async () => {
            if (!isMounted) return;

            // ⚡ BOLT: Paused polling while tab is hidden to save resources.
            // Randomized Selection from Top 5 High-Impact Targets (Network/IO)
            if (document.hidden) {
                timeoutId = setTimeout(loop, 1000);
                return;
            }

            await fetchData();

            if (isMounted) {
                // ⚡ BOLT: Switched to setTimeout to prevent request waterfalls.
                // Randomized Selection from Top 5 High-Impact Targets (Network/IO)
                timeoutId = setTimeout(loop, 1000);
            }
        };

        // Start the loop after delay to avoid immediate double-fetch
        timeoutId = setTimeout(loop, 1000);

        return () => {
            isMounted = false;
            clearTimeout(timeoutId);
        };
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
