/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useMemo, useEffect } from "react";
import {
    ReactFlow,
    Controls,
    Background,
    BackgroundVariant,
    useNodesState,
    useEdgesState,
    Position,
    Node,
    Edge,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import jsyaml from "js-yaml";
import dagre from "dagre";
import {
    Database,
    Terminal,
    Cpu,
    Box,
    AlertTriangle
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface StackGraphProps {
    yamlContent: string;
}

interface ParsedService {
    name: string;
    image?: string;
    command?: string;
    envCount: number;
    ports: string[];
    type: "image" | "command" | "unknown";
    dependsOn?: string[];
}

const nodeWidth = 250;
const nodeHeight = 100;

const getLayoutedElements = (nodes: Node[], edges: Edge[]) => {
    const dagreGraph = new dagre.graphlib.Graph();
    dagreGraph.setDefaultEdgeLabel(() => ({}));

    dagreGraph.setGraph({ rankdir: 'TB' });

    nodes.forEach((node) => {
        dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
    });

    edges.forEach((edge) => {
        dagreGraph.setEdge(edge.source, edge.target);
    });

    dagre.layout(dagreGraph);

    const layoutedNodes = nodes.map((node) => {
        const nodeWithPosition = dagreGraph.node(node.id);
        node.position = {
            x: nodeWithPosition.x - nodeWidth / 2,
            y: nodeWithPosition.y - nodeHeight / 2,
        };
        return node;
    });

    return { nodes: layoutedNodes, edges };
};

/**
 * ServiceNode component for ReactFlow.
 */
const ServiceNode = React.memo(({ data }: { data: ParsedService }) => {
    let Icon = Cpu;
    if (data.type === "image") Icon = Database;
    if (data.type === "command") Icon = Terminal;

    return (
        <Card className="w-[240px] shadow-md border-l-4 border-l-primary/50 text-xs">
            <CardHeader className="p-2 pb-1 bg-muted/20">
                <CardTitle className="text-sm font-medium flex items-center gap-2">
                    <Icon className="h-4 w-4 text-primary" />
                    <span className="truncate" title={data.name}>{data.name}</span>
                </CardTitle>
            </CardHeader>
            <CardContent className="p-2 space-y-1">
                {data.image && (
                    <div className="flex items-center gap-1 text-[10px] text-muted-foreground truncate">
                        <Box className="h-3 w-3" />
                        <span className="truncate" title={data.image}>{data.image}</span>
                    </div>
                )}
                <div className="flex gap-1 mt-1">
                    {data.envCount > 0 && <Badge variant="secondary" className="text-[9px] px-1 h-4">{data.envCount} Env</Badge>}
                    {data.ports.length > 0 && <Badge variant="secondary" className="text-[9px] px-1 h-4">{data.ports.length} Port</Badge>}
                </div>
            </CardContent>
        </Card>
    );
});
ServiceNode.displayName = "ServiceNode";

const nodeTypes = {
    service: ServiceNode,
};

/**
 * StackGraph component.
 * Visualizes the stack configuration as a graph.
 *
 * @param { yamlContent } - StackGraphProps. Description.
 */
export function StackGraph({ yamlContent }: StackGraphProps) {
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);

    const { services, error } = useMemo(() => {
        try {
            const parsed = jsyaml.load(yamlContent) as any;
            if (!parsed || typeof parsed !== 'object') {
                return { services: [], error: null };
            }

            const rawServices = parsed.services || {};
            let serviceList: ParsedService[] = [];

            const processService = (key: string, val: any): ParsedService => {
                 const env = val.environment || val.env || {};
                 const envCount = Array.isArray(env) ? env.length : Object.keys(env).length;
                 const ports = val.ports || [];
                 const dependsOn = val.depends_on || [];

                 let type: ParsedService['type'] = "unknown";
                 if (val.mcp_service) {
                     type = "image";
                     if (val.mcp_service.stdio_connection) {
                         val.image = val.mcp_service.stdio_connection.container_image;
                         val.command = val.mcp_service.stdio_connection.command;
                     } else if (val.mcp_service.bundle_connection) {
                         val.image = val.mcp_service.bundle_connection.container_image;
                     }
                 } else if (val.image) type = "image";
                 else if (val.command) type = "command";

                 return {
                     name: val.name || key,
                     image: val.image,
                     command: val.command,
                     envCount,
                     ports,
                     type,
                     dependsOn: Array.isArray(dependsOn) ? dependsOn : Object.keys(dependsOn)
                 };
            };

            if (Array.isArray(rawServices)) {
                serviceList = rawServices.map((val: any) => processService(val.name || "unknown", val));
            } else {
                serviceList = Object.entries(rawServices).map(([key, val]: [string, any]) => processService(key, val));
            }

            return { services: serviceList, error: null };
        } catch (e: any) {
            return { services: [], error: e.message };
        }
    }, [yamlContent]);

    useEffect(() => {
        if (error || services.length === 0) {
            setNodes([]);
            setEdges([]);
            return;
        }

        const newNodes: Node[] = services.map((svc) => ({
            id: svc.name,
            type: 'service',
            data: { ...svc },
            position: { x: 0, y: 0 }, // Initial position, will be layouted
        }));

        const newEdges: Edge[] = [];
        services.forEach((svc) => {
            if (svc.dependsOn) {
                svc.dependsOn.forEach((dep) => {
                    // Check if dependency exists
                    if (services.some(s => s.name === dep)) {
                        newEdges.push({
                            id: `${svc.name}-${dep}`,
                            source: dep, // Dependency is source (must start first)
                            target: svc.name,
                            animated: true,
                        });
                    }
                });
            }
        });

        const layouted = getLayoutedElements(newNodes, newEdges);
        setNodes(layouted.nodes);
        setEdges(layouted.edges);

    }, [services, error, setNodes, setEdges]);

    if (error) {
         return (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-4 gap-2 bg-red-50/10">
                <AlertTriangle className="h-8 w-8 text-destructive opacity-50" />
                <p className="text-xs text-destructive font-medium text-center">YAML Syntax Error</p>
                <p className="text-[10px] font-mono opacity-75 max-w-[200px] break-all text-center">{error}</p>
            </div>
        );
    }

    if (services.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground p-4 gap-2">
                <Box className="h-8 w-8 opacity-20" />
                <p className="text-xs">No services defined</p>
            </div>
        );
    }

    return (
        <div className="h-full w-full bg-muted/5">
            <ReactFlow
                nodes={nodes}
                edges={edges}
                onNodesChange={onNodesChange}
                onEdgesChange={onEdgesChange}
                nodeTypes={nodeTypes}
                fitView
                attributionPosition="bottom-right"
            >
                <Background variant={BackgroundVariant.Dots} gap={12} size={1} />
                <Controls showInteractive={false} className="bg-background/80 backdrop-blur border-muted shadow-sm" />
            </ReactFlow>
        </div>
    );
}
