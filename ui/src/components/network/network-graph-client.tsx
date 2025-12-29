/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useState } from 'react';
import {
  ReactFlow,
  MiniMap,
  Controls,
  Background,
  BackgroundVariant,
  Panel,
  Node,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
    SheetDescription
} from "@/components/ui/sheet";
import {
    RefreshCcw,
    Zap,
    Server,
    Database,
    Users,
    Activity,
    Webhook,
    Layers
} from "lucide-react";

import { useNetworkTopology } from "@/hooks/use-network-topology";
import { Node } from '@xyflow/react';
import { NodeType, NodeStatus } from "@/types/topology";

interface NodeData extends Record<string, unknown> {
    label: string;
    type: NodeType;
    status: NodeStatus;
    metrics?: { qps?: number; latencyMs?: number; errorRate?: number };
    metadata?: Record<string, string>;
}

export function NetworkGraphClient() {
  const { nodes, edges, onNodesChange, onEdgesChange, onConnect, refreshTopology, autoLayout } = useNetworkTopology();
  const [selectedNode, setSelectedNode] = useState<Node<NodeData> | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);

  const onNodeClick = (event: React.MouseEvent, node: Node) => {
      setSelectedNode(node as Node<NodeData>);
      setIsSheetOpen(true);
  };

  const getNodeIcon = (type: NodeType) => {
      switch (type) {
          case 'NODE_TYPE_CORE': return <Server className="h-5 w-5" />;
          case 'NODE_TYPE_SERVICE': return <Database className="h-5 w-5" />;
          case 'NODE_TYPE_CLIENT': return <Users className="h-5 w-5" />;
          case 'NODE_TYPE_TOOL': return <Zap className="h-5 w-5" />;
          case 'NODE_TYPE_API_CALL': return <Activity className="h-5 w-5" />;
          case 'NODE_TYPE_MIDDLEWARE': return <Layers className="h-5 w-5" />;
          case 'NODE_TYPE_WEBHOOK': return <Webhook className="h-5 w-5" />;
          default: return <Activity className="h-5 w-5" />;
      }
  };

  const getStatusBadgeVariant = (status: NodeStatus) => {
      switch (status) {
          case 'NODE_STATUS_ACTIVE': return 'default';
          case 'NODE_STATUS_INACTIVE': return 'secondary';
          case 'NODE_STATUS_ERROR': return 'destructive';
          default: return 'outline';
      }
  };

  return (
    <div className="h-full w-full relative bg-muted/10">
      <div className="absolute top-4 left-4 z-10 space-y-4 pointer-events-none">
          <Card className="w-[300px] pointer-events-auto backdrop-blur-sm bg-background/80 shadow-md border-muted">
              <CardHeader className="p-4 pb-2">
                  <CardTitle className="text-lg flex items-center gap-2">
                      <Activity className="h-5 w-5 text-primary" />
                      Network Topology
                  </CardTitle>
                  <CardDescription>
                      Visualizing {nodes.length} nodes and {edges.length} connections.
                  </CardDescription>
              </CardHeader>
              <CardContent className="p-4 pt-2 flex gap-2">
                  <Button variant="outline" size="sm" onClick={refreshTopology} className="flex-1">
                      <RefreshCcw className="mr-2 h-4 w-4" /> Refresh
                  </Button>
                  <Button size="sm" onClick={autoLayout} className="flex-1">
                      <Zap className="mr-2 h-4 w-4" /> Auto-Layout
                  </Button>
              </CardContent>
          </Card>
      </div>

      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onNodeClick={onNodeClick}
        fitView
        className="bg-background"
      >
        <Controls />
        <MiniMap />
        <Background variant={BackgroundVariant.Dots} gap={12} size={1} />

        <Panel position="bottom-center" className="bg-background/80 p-2 rounded-full border shadow-sm backdrop-blur text-xs text-muted-foreground flex gap-4">
            <div className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-white border border-black dark:bg-slate-900 dark:border-white"></div> Core</div>
            <div className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-blue-50 border border-blue-500 dark:bg-blue-950 dark:border-blue-600"></div> Service</div>
            <div className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-green-50 border border-green-500 dark:bg-green-950 dark:border-green-600"></div> Client</div>
            <div className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-fuchsia-50 border border-fuchsia-500 dark:bg-fuchsia-950 dark:border-fuchsia-600"></div> Tool</div>
            <div className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-orange-50 border border-orange-500 dark:bg-orange-950 dark:border-orange-600"></div> Middleware</div>
            <div className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-pink-50 border border-pink-500 dark:bg-pink-950 dark:border-pink-600"></div> Webhook</div>
        </Panel>
      </ReactFlow>

      <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
        <SheetContent>
            <SheetHeader>
                <SheetTitle className="flex items-center gap-2">
                    {selectedNode && getNodeIcon(selectedNode.data.type)}
                    {selectedNode?.data.label}
                </SheetTitle>
                <SheetDescription>
                    ID: {selectedNode?.id}
                </SheetDescription>
            </SheetHeader>
            <div className="grid gap-4 py-4">
                <div className="flex items-center justify-between p-2 border rounded-md">
                    <span className="text-sm font-medium">Status</span>
                    <Badge variant={selectedNode ? getStatusBadgeVariant(selectedNode.data.status) : 'outline'}>
                        {selectedNode?.data.status}
                    </Badge>
                </div>

                {selectedNode?.data.metrics && (
                    <div className="space-y-2">
                        <h4 className="font-medium text-sm">Metrics</h4>
                        <div className="grid grid-cols-2 gap-2">
                            <Card className="p-2 bg-muted/30">
                                <div className="text-xs text-muted-foreground">QPS</div>
                                <div className="text-lg font-bold">{selectedNode.data.metrics.qps?.toFixed(2) || 0}</div>
                            </Card>
                             <Card className="p-2 bg-muted/30">
                                <div className="text-xs text-muted-foreground">Latency</div>
                                <div className="text-lg font-bold">{selectedNode.data.metrics.latencyMs?.toFixed(1) || 0}ms</div>
                            </Card>
                             <Card className="p-2 bg-muted/30">
                                <div className="text-xs text-muted-foreground">Error Rate</div>
                                <div className="text-lg font-bold">{((selectedNode.data.metrics.errorRate || 0) * 100).toFixed(1)}%</div>
                            </Card>
                        </div>
                    </div>
                )}

                 <div className="space-y-2">
                    <h4 className="font-medium text-sm">Metadata</h4>
                    <pre className="text-xs bg-muted p-2 rounded overflow-x-auto whitespace-pre-wrap">
                        {JSON.stringify(selectedNode?.data.metadata || {}, null, 2)}
                    </pre>
                </div>
            </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}
