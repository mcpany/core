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
    Activity
} from "lucide-react";

import { useNetworkTopology, NodeData } from "@/hooks/use-network-topology";

export function NetworkGraphClient() {
  const { nodes, edges, onNodesChange, onEdgesChange, onConnect, refreshTopology, autoLayout } = useNetworkTopology();
  const [selectedNode, setSelectedNode] = useState<Node<NodeData> | null>(null);
  const [isSheetOpen, setIsSheetOpen] = useState(false);

  const onNodeClick = (event: React.MouseEvent, node: Node) => {
      // Cast to correct type if needed, or rely on React Flow types
      setSelectedNode(node as Node<NodeData>);
      setIsSheetOpen(true);
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
            <div className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-white border border-black"></div> Core</div>
            <div className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-blue-50 border border-blue-500"></div> Upstream</div>
            <div className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-green-50 border border-green-500"></div> Agent (Active)</div>
            <div className="flex items-center gap-1"><div className="w-2 h-2 rounded-full bg-red-50 border border-red-500"></div> Agent (Inactive)</div>
        </Panel>
      </ReactFlow>

      <Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
        <SheetContent>
            <SheetHeader>
                <SheetTitle className="flex items-center gap-2">
                    {selectedNode?.data.type === 'mcp-server' ? <Server className="h-5 w-5" /> :
                     selectedNode?.data.type === 'upstream' ? <Database className="h-5 w-5" /> :
                     <Users className="h-5 w-5" />}
                    {selectedNode?.data.label}
                </SheetTitle>
                <SheetDescription>
                    ID: {selectedNode?.id}
                </SheetDescription>
            </SheetHeader>
            <div className="grid gap-4 py-4">
                <div className="flex items-center justify-between p-2 border rounded-md">
                    <span className="text-sm font-medium">Status</span>
                    <Badge variant={selectedNode?.data.status === 'active' ? 'default' : 'destructive'}>
                        {selectedNode?.data.status}
                    </Badge>
                </div>
                <div className="space-y-2">
                    <h4 className="font-medium text-sm">Metrics</h4>
                    <div className="grid grid-cols-2 gap-2">
                        <Card className="p-2 bg-muted/30">
                            <div className="text-xs text-muted-foreground">Requests/min</div>
                            <div className="text-lg font-bold">1,240</div>
                        </Card>
                         <Card className="p-2 bg-muted/30">
                            <div className="text-xs text-muted-foreground">Latency</div>
                            <div className="text-lg font-bold">45ms</div>
                        </Card>
                    </div>
                </div>
                 <div className="space-y-2">
                    <h4 className="font-medium text-sm">Metadata</h4>
                    <pre className="text-xs bg-muted p-2 rounded overflow-x-auto">
                        {JSON.stringify(selectedNode?.data, null, 2)}
                    </pre>
                </div>
            </div>
        </SheetContent>
      </Sheet>
    </div>
  );
}
