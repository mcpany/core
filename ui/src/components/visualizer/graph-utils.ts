/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { Node, Edge, Position } from '@xyflow/react';
import dagre from 'dagre';

// Types matching the Proto JSON response
export interface GraphNode {
    id: string;
    label: string;
    type: string; // NODE_TYPE_*
    status: string; // NODE_STATUS_*
    metadata?: Record<string, string>;
    children?: GraphNode[];
    metrics?: {
        qps: number;
        latency_ms: number;
        error_rate: number;
    };
}

export interface GraphResponse {
    clients: GraphNode[];
    core: GraphNode;
}

const nodeWidth = 220;
const nodeHeight = 80;

/**
 * Transforms the API Graph response into React Flow nodes and edges with layout.
 * @param graph The graph data from the API.
 * @returns React Flow nodes and edges.
 */
export function transformGraphToReactFlow(graph: GraphResponse | null): { nodes: Node[], edges: Edge[] } {
    if (!graph || !graph.core) {
        return { nodes: [], edges: [] };
    }

    const nodes: Node[] = [];
    const edges: Edge[] = [];
    const g = new dagre.graphlib.Graph();

    g.setGraph({ rankdir: 'LR', align: 'UL', nodesep: 50, ranksep: 100 });
    g.setDefaultEdgeLabel(() => ({}));

    // --- Core Node ---
    const coreNode = graph.core;
    // Map Core to 'agent' type for visual distinction (it's the orchestrator)
    nodes.push(mapToNode(coreNode, 'agent', { role: 'Gateway', status: 'Active' }));
    g.setNode(coreNode.id, { width: nodeWidth, height: nodeHeight });

    // --- Services & Tools (Children of Core) ---
    if (coreNode.children) {
        coreNode.children.forEach(child => {
            if (child.type === 'NODE_TYPE_SERVICE') {
                nodes.push(mapToNode(child, 'service', { label: child.label }));
                g.setNode(child.id, { width: nodeWidth, height: nodeHeight });

                // Edge Core -> Service
                edges.push({
                    id: `e-${coreNode.id}-${child.id}`,
                    source: coreNode.id,
                    target: child.id,
                    animated: true,
                    type: 'smoothstep',
                });
                g.setEdge(coreNode.id, child.id);

                // Tools inside Service
                if (child.children) {
                    child.children.forEach(grandChild => {
                        if (grandChild.type === 'NODE_TYPE_TOOL') {
                            nodes.push(mapToNode(grandChild, 'tool', { label: grandChild.label }));
                            g.setNode(grandChild.id, { width: nodeWidth, height: nodeHeight });

                            // Edge Service -> Tool
                            edges.push({
                                id: `e-${child.id}-${grandChild.id}`,
                                source: child.id,
                                target: grandChild.id,
                                type: 'smoothstep',
                            });
                            g.setEdge(child.id, grandChild.id);
                        } else if (grandChild.type === 'NODE_TYPE_MIDDLEWARE') {
                             // Treat middleware as service-like or resource
                             nodes.push(mapToNode(grandChild, 'resource', { label: grandChild.label }));
                             g.setNode(grandChild.id, { width: nodeWidth, height: nodeHeight });

                             // Edge Core -> Middleware (Wait, usually Middleware is inside Core? Or attached?)
                             // In manager.go, Middleware is child of Core.
                             // But here we are iterating Core's children.
                             // Wait, if it's direct child of Core, we already iterate it.
                             // This block is inside 'NODE_TYPE_SERVICE'.
                             // Middleware is usually direct child of Core in manager.go.
                        }
                    });
                }
            } else if (child.type === 'NODE_TYPE_MIDDLEWARE') {
                 nodes.push(mapToNode(child, 'resource', { label: child.label }));
                 g.setNode(child.id, { width: nodeWidth, height: nodeHeight });

                 edges.push({
                    id: `e-${coreNode.id}-${child.id}`,
                    source: coreNode.id,
                    target: child.id,
                    animated: true,
                    style: { strokeDasharray: '5 5' },
                    type: 'smoothstep',
                });
                g.setEdge(coreNode.id, child.id);
            } else if (child.type === 'NODE_TYPE_WEBHOOK') {
                 nodes.push(mapToNode(child, 'resource', { label: child.label }));
                 g.setNode(child.id, { width: nodeWidth, height: nodeHeight });
                 edges.push({
                    id: `e-${coreNode.id}-${child.id}`,
                    source: coreNode.id,
                    target: child.id,
                    type: 'smoothstep',
                });
                g.setEdge(coreNode.id, child.id);
            }
        });
    }

    // --- Clients ---
    if (graph.clients) {
        graph.clients.forEach(client => {
            nodes.push(mapToNode(client, 'user', { label: client.label || client.id }));
            g.setNode(client.id, { width: nodeWidth, height: nodeHeight });

            // Edge Client -> Core
            edges.push({
                id: `e-${client.id}-${coreNode.id}`,
                source: client.id,
                target: coreNode.id,
                animated: true,
                type: 'smoothstep',
            });
            g.setEdge(client.id, coreNode.id);
        });
    }

    dagre.layout(g);

    // Apply computed layout
    const layoutNodes = nodes.map(node => {
        const nodeWithPosition = g.node(node.id);
        return {
            ...node,
            position: {
                x: nodeWithPosition.x - nodeWidth / 2,
                y: nodeWithPosition.y - nodeHeight / 2,
            },
            targetPosition: Position.Left,
            sourcePosition: Position.Right,
        };
    });

    return { nodes: layoutNodes, edges };
}

function mapToNode(protoNode: GraphNode, type: string, dataOverride: Record<string, unknown> = {}): Node {
    return {
        id: protoNode.id,
        type: type,
        position: { x: 0, y: 0 }, // Will be set by layout
        data: {
            label: protoNode.label,
            status: mapStatus(protoNode.status),
            metrics: protoNode.metrics,
            ...dataOverride
        }
    };
}

function mapStatus(status: string): string {
    switch (status) {
        case 'NODE_STATUS_ACTIVE': return 'Active';
        case 'NODE_STATUS_INACTIVE': return 'Inactive';
        case 'NODE_STATUS_ERROR': return 'Error';
        default: return 'Unknown';
    }
}
