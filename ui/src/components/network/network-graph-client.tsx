/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useCallback, useState } from "react";
import {
  ReactFlow,
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  Edge,
  Node,
  Position,
  Handle,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Server,
  Database,
  Globe,
  Cpu,
  Activity,
  ShieldCheck,
  Terminal,
  Zap
} from "lucide-react";
import { cn } from "@/lib/utils";

// --- Custom Node Types ---

const CustomNode = ({ data, selected }: { data: any, selected: boolean }) => {
  const Icon = data.icon || Server;

  return (
    <div className={cn(
      "px-4 py-3 shadow-md rounded-xl bg-card border-2 transition-all duration-300 min-w-[180px]",
      selected ? "border-primary ring-2 ring-primary/20 shadow-lg" : "border-muted hover:border-primary/50",
      data.status === "error" ? "border-destructive/50" : ""
    )}>
      <Handle type="target" position={Position.Top} className="!bg-muted-foreground" />

      <div className="flex items-center gap-3">
        <div className={cn(
          "p-2 rounded-lg",
          data.type === "host" ? "bg-blue-500/10 text-blue-500" :
          data.type === "server" ? "bg-green-500/10 text-green-500" :
          data.type === "tool" ? "bg-amber-500/10 text-amber-500" :
          "bg-muted text-muted-foreground"
        )}>
          <Icon className="size-5" />
        </div>
        <div>
          <div className="font-semibold text-sm">{data.label}</div>
          <div className="text-xs text-muted-foreground capitalize">{data.type}</div>
        </div>
      </div>

      {data.stats && (
         <div className="mt-3 pt-2 border-t text-[10px] grid grid-cols-2 gap-2 text-muted-foreground">
            {Object.entries(data.stats).map(([k, v]: [string, any]) => (
              <div key={k} className="flex justify-between">
                <span>{k}:</span>
                <span className="font-mono text-foreground">{v}</span>
              </div>
            ))}
         </div>
      )}

      {data.status === "active" && (
        <span className="absolute -top-1 -right-1 flex h-3 w-3">
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
          <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
        </span>
      )}

      <Handle type="source" position={Position.Bottom} className="!bg-muted-foreground" />
    </div>
  );
};

const nodeTypes = {
  custom: CustomNode,
};

// --- Mock Data ---

const initialNodes: Node[] = [
  {
    id: "host",
    type: "custom",
    position: { x: 400, y: 50 },
    data: {
      label: "MCP Host",
      type: "host",
      icon: Cpu,
      status: "active",
      stats: { "Uptime": "24h", "Load": "12%" }
    },
  },
  {
    id: "srv-1",
    type: "custom",
    position: { x: 150, y: 250 },
    data: {
      label: "Postgres DB",
      type: "server",
      icon: Database,
      status: "active",
      stats: { "Conns": "5", "Latency": "2ms" }
    },
  },
  {
    id: "srv-2",
    type: "custom",
    position: { x: 400, y: 250 },
    data: {
      label: "GitHub API",
      type: "server",
      icon: Globe,
      status: "active",
      stats: { "RateLimit": "4500", "Status": "OK" }
    },
  },
  {
    id: "srv-3",
    type: "custom",
    position: { x: 650, y: 250 },
    data: {
      label: "Local Filesystem",
      type: "server",
      icon: Server,
      status: "idle",
      stats: { "Read": "0kb/s", "Write": "0kb/s" }
    },
  },
  {
    id: "tool-1",
    type: "custom",
    position: { x: 50, y: 450 },
    data: { label: "Query Tool", type: "tool", icon: Terminal },
  },
  {
    id: "tool-2",
    type: "custom",
    position: { x: 250, y: 450 },
    data: { label: "Schema Insp.", type: "tool", icon: ShieldCheck },
  },
  {
    id: "tool-3",
    type: "custom",
    position: { x: 400, y: 450 },
    data: { label: "PR Agent", type: "tool", icon: Zap },
  },
  {
    id: "tool-4",
    type: "custom",
    position: { x: 650, y: 450 },
    data: { label: "File Reader", type: "tool", icon: Terminal },
  },
];

const initialEdges: Edge[] = [
  { id: "e1", source: "host", target: "srv-1", animated: true },
  { id: "e2", source: "host", target: "srv-2", animated: true },
  { id: "e3", source: "host", target: "srv-3" },
  { id: "e4", source: "srv-1", target: "tool-1" },
  { id: "e5", source: "srv-1", target: "tool-2" },
  { id: "e6", source: "srv-2", target: "tool-3", animated: true, style: { stroke: '#22c55e' } },
  { id: "e7", source: "srv-3", target: "tool-4" },
];

export function NetworkGraphClient() {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges],
  );

  const onNodeClick = (_: React.MouseEvent, node: Node) => {
    setSelectedNode(node);
  };

  const onPaneClick = () => {
    setSelectedNode(null);
  };

  return (
    <div className="flex flex-col h-full gap-4 p-4">
      <div className="flex items-center justify-between">
          <div className="space-y-1">
            <h2 className="text-2xl font-bold tracking-tight flex items-center gap-2">
                <Activity className="text-primary" /> Network Graph
            </h2>
            <p className="text-sm text-muted-foreground">
                Visualize connections between your MCP host, servers, and tools.
            </p>
          </div>
          <Badge variant="outline" className="text-muted-foreground gap-1">
             <div className="size-2 rounded-full bg-green-500 animate-pulse" /> Live Topology
          </Badge>
      </div>

      <div className="flex-1 relative rounded-xl border bg-card shadow-sm overflow-hidden h-[600px]">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeClick={onNodeClick}
          onPaneClick={onPaneClick}
          nodeTypes={nodeTypes}
          fitView
          className="bg-background/50"
        >
          <Controls className="bg-background border shadow-sm" />
          <MiniMap
            className="bg-background border shadow-sm rounded-lg"
            nodeColor={(n) => {
                if (n.data.type === 'host') return '#3b82f6';
                if (n.data.type === 'server') return '#22c55e';
                if (n.data.type === 'tool') return '#f59e0b';
                return '#eee';
            }}
          />
          <Background color="#888" gap={16} size={1} className="opacity-10" />
        </ReactFlow>

        {/* Detail Panel */}
        {selectedNode && (
            <div className="absolute top-4 right-4 w-80 animate-in slide-in-from-right-10 duration-200 fade-in">
                <Card className="backdrop-blur-xl bg-background/80 shadow-2xl border-primary/20">
                    <CardHeader className="pb-3">
                        <CardTitle className="text-lg flex items-center gap-2">
                            {/* @ts-ignore */}
                            <selectedNode.data.icon className="size-5 text-primary" />
                            {selectedNode.data.label as string}
                        </CardTitle>
                        <CardDescription>
                             ID: {selectedNode.id}
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="grid grid-cols-2 gap-2 text-sm">
                            <div className="text-muted-foreground">Type</div>
                            <div className="font-medium capitalize text-right">{selectedNode.data.type as string}</div>

                            <div className="text-muted-foreground">Status</div>
                            <div className="font-medium text-right flex justify-end">
                                <Badge variant="secondary" className={cn(
                                    "capitalize",
                                    selectedNode.data.status === 'active' ? "bg-green-500/10 text-green-500" : "bg-muted text-muted-foreground"
                                )}>
                                    {selectedNode.data.status as string || 'Unknown'}
                                </Badge>
                            </div>
                        </div>

                        {selectedNode.data.stats && (
                            <div className="space-y-2 pt-2 border-t">
                                <span className="text-xs font-semibold text-muted-foreground uppercase">Metrics</span>
                                <div className="space-y-1">
                                    {Object.entries(selectedNode.data.stats as object).map(([k, v]) => (
                                        <div key={k} className="flex justify-between text-sm">
                                            <span className="text-muted-foreground">{k}</span>
                                            <span className="font-mono">{v}</span>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        )}

                        <div className="pt-2">
                           <button className="w-full text-xs text-primary hover:underline text-center">
                                View Full Logs &rarr;
                           </button>
                        </div>
                    </CardContent>
                </Card>
            </div>
        )}
      </div>
    </div>
  );
}
