/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useCallback, useEffect } from 'react';
import { Node, Edge, useNodesState, useEdgesState, addEdge, Connection, MarkerType, Position } from '@xyflow/react';
import dagre from 'dagre';
import { Graph, Node as TopologyNode } from '../types/topology';

/**
 * State and handlers for the network topology graph.
 */
export interface NetworkGraphState {
    /** The list of nodes in the graph. */
    nodes: Node[];
    /** The list of edges in the graph. */
    edges: Edge[];
    /** Handler for node changes (drag, selection, etc.). */
    onNodesChange: any;
    /** Handler for edge changes. */
    onEdgesChange: any;
    /** Handler for creating new connections between nodes. */
    onConnect: (params: Connection) => void;
    /** Function to force a refresh of the topology data from the backend. */
    refreshTopology: () => void;
    /** Function to re-run the auto-layout algorithm. */
    autoLayout: () => void;
}

const nodeWidth = 220;
const nodeHeight = 60;

/**
 * Computes the layout of nodes and edges using Dagre.
 *
 * @param nodes - The nodes to layout.
 * @param edges - The edges connecting the nodes.
 * @param direction - The direction of the layout ('TB' for Top-Bottom, 'LR' for Left-Right).
 * @returns An object containing the layouted nodes and edges.
 */
const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'TB') => {
    const dagreGraph = new dagre.graphlib.Graph();
    dagreGraph.setDefaultEdgeLabel(() => ({}));

    dagreGraph.setGraph({ rankdir: direction });

    nodes.forEach((node) => {
        dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
    });

    edges.forEach((edge) => {
        dagreGraph.setEdge(edge.source, edge.target);
    });

    dagre.layout(dagreGraph);

    const layoutedNodes = nodes.map((node) => {
        const nodeWithPosition = dagreGraph.node(node.id);

        // Dagre returns center coordinates, React Flow needs top-left
        return {
            ...node,
            targetPosition: direction === 'TB' ? Position.Top : Position.Left,
            sourcePosition: direction === 'TB' ? Position.Bottom : Position.Right,
            position: {
                x: nodeWithPosition.x - nodeWidth / 2,
                y: nodeWithPosition.y - nodeHeight / 2,
            },
        };
    });

    return { nodes: layoutedNodes, edges };
};

/**
 * Hook to manage the network topology graph.
 *
 * It fetches the topology data from the backend, processes it into React Flow nodes and edges,
 * and handles layout and updates.
 *
 * @returns The current state and handlers for the network graph.
 */
export function useNetworkTopology() {
    const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);

    const fetchData = useCallback(async () => {
        try {
            const res = await fetch('/api/v1/topology');
            if (!res.ok) throw new Error('Failed to fetch topology');
            const graph: Graph = await res.json();

            const newNodes: Node[] = [];
            const newEdges: Edge[] = [];

            // Helper to add node
            const addNode = (tNode: TopologyNode, parentId?: string) => {
                // isGroup removed as unused
                // We flatten the graph for Dagre but visually we might use Groups?
                // For now, let's just make them all top-level nodes connected by edges
                // to simplify the 5-layer layout request.
                // Groups in React Flow are for containment.
                // "Unifi Topology" usually connects them with lines, not boxes inside boxes.

                const flowNode: Node = {
                    id: tNode.id,
                    data: {
                        label: tNode.label,
                        type: tNode.type,
                        status: tNode.status,
                        metrics: tNode.metrics
                    },
                    position: { x: 0, y: 0 }, // Set by layout
                    // Layout styles remain in style
                    style: getNodeLayout(tNode),
                    // Colors and visuals move to className for dark mode support
                    className: getNodeClassName(tNode),
                    type: 'default', // Custom types in future
                };

                newNodes.push(flowNode);

                if (parentId) {
                    newEdges.push({
                        id: `e-${parentId}-${tNode.id}`,
                        source: parentId,
                        target: tNode.id,
                        animated: tNode.status === 'NODE_STATUS_ACTIVE',
                        style: { stroke: '#b1b1b7' },
                        markerEnd: { type: MarkerType.ArrowClosed, color: '#b1b1b7' },
                        label: tNode.metrics ? `${tNode.metrics.qps?.toFixed(1) || 0} QPS` : undefined
                    });
                }

                // Process Children with Truncation
                if (tNode.children && tNode.children.length > 0) {
                     const limit = (tNode.type === 'NODE_TYPE_CORE' || tNode.type === 'NODE_TYPE_SERVICE') ? 100 : 100; // Disable limit for now to see all
                     tNode.children.slice(0, limit).forEach(child => addNode(child, tNode.id));
                }
            };

            // 1. Clients
            if (graph.clients) {
                graph.clients.forEach(client => {
                    addNode(client);
                    // Connect Client to Core
                    if (graph.core) {
                        newEdges.push({
                            id: `e-${client.id}-${graph.core.id}`,
                            source: client.id,
                            target: graph.core.id,
                            animated: true,
                            style: { stroke: '#22c55e' },
                            markerEnd: { type: MarkerType.ArrowClosed, color: '#22c55e' }
                        });
                    }
                });
            }

            // 2. Core (and its children: Services -> Tools -> API)
            if (graph.core) {
                addNode(graph.core);
            }

            const layouted = getLayoutedElements(newNodes, newEdges);
            setNodes(layouted.nodes);
            setEdges(layouted.edges);

        } catch (error) {
            console.error("Failed to fetch topology data:", error);
        }
    }, [setNodes, setEdges]);

    // Fetch on mount
    useEffect(() => {
        fetchData();
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
         fetchData();
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

function getNodeLayout(node: TopologyNode) {
    return {
        borderRadius: '8px',
        padding: '10px',
        fontSize: '12px',
        width: nodeWidth,
        display: 'flex',
        flexDirection: 'column' as const,
        alignItems: 'center',
        justifyContent: 'center',
        borderWidth: node.type === 'NODE_TYPE_CORE' ? '2px' : '1px',
        borderStyle: node.type === 'NODE_TYPE_API_CALL' ? 'dashed' : 'solid',
    };
}

function getNodeClassName(node: TopologyNode): string {
    const base = "transition-all duration-200 shadow-sm hover:shadow-md";

    switch (node.type) {
        case 'NODE_TYPE_CLIENT':
            return `${base} bg-green-50 border-green-500 text-green-900 dark:bg-green-900 dark:border-green-600 dark:text-green-100`;
        case 'NODE_TYPE_CORE':
            return `${base} bg-white border-black text-black font-bold text-sm dark:bg-slate-800 dark:border-white dark:text-white`;
        case 'NODE_TYPE_SERVICE':
            return `${base} bg-blue-50 border-blue-500 text-blue-900 dark:bg-blue-900 dark:border-blue-600 dark:text-blue-100`;
        case 'NODE_TYPE_TOOL':
            return `${base} bg-fuchsia-50 border-fuchsia-500 text-fuchsia-900 dark:bg-fuchsia-900 dark:border-fuchsia-600 dark:text-fuchsia-100`;
        case 'NODE_TYPE_RESOURCE':
            return `${base} bg-indigo-50 border-indigo-500 text-indigo-900 dark:bg-indigo-900 dark:border-indigo-600 dark:text-indigo-100`;
        case 'NODE_TYPE_PROMPT':
            return `${base} bg-violet-50 border-violet-500 text-violet-900 dark:bg-violet-900 dark:border-violet-600 dark:text-violet-100`;
        case 'NODE_TYPE_API_CALL':
            return `${base} bg-zinc-50 border-zinc-500 text-zinc-900 dark:bg-zinc-800 dark:border-zinc-400 dark:text-zinc-300`;
        case 'NODE_TYPE_MIDDLEWARE':
            return `${base} bg-orange-50 border-orange-500 text-orange-900 dark:bg-orange-900 dark:border-orange-600 dark:text-orange-100`;
        case 'NODE_TYPE_WEBHOOK':
            return `${base} bg-pink-50 border-pink-500 text-pink-900 dark:bg-pink-900 dark:border-pink-600 dark:text-pink-100`;
        default:
            return `${base} bg-white border-gray-300 text-gray-900 dark:bg-gray-900 dark:border-gray-600 dark:text-gray-100`;
    }
}
