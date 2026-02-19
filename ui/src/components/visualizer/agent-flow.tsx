/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  ReactFlow,
  useNodesState,
  useEdgesState,
  Controls,
  Background,
  MiniMap,
  Panel,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { RefreshCw, Database } from 'lucide-react';
import { VariableInspector } from './variable-inspector';
import { useTopology } from '@/hooks/use-topology';
import { getLayoutedElements } from './layout';
import { apiClient } from '@/lib/client';
import { useToast } from '@/hooks/use-toast';

const nodeTypes = {
  user: UserNode,
  agent: AgentNode,
  tool: ToolNode,
  resource: ResourceNode,
  service: ServiceNode,
};

/**
 * NetworkTopology component renders the real-time service topology.
 * @returns The NetworkTopology component.
 */
export function AgentFlow() {
  const [isLive, setIsLive] = useState(false);
  const { graph, loading, refresh } = useTopology(isLive);
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [selectedNode, setSelectedNode] = useState<any>(null);
  const { toast } = useToast();

  const onNodeClick = useCallback((_: any, node: any) => {
    setSelectedNode(node);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  // Update graph when data changes
  useEffect(() => {
      if (graph) {
          const layouted = getLayoutedElements(graph);
          setNodes(layouted.nodes);
          setEdges(layouted.edges);
      }
  }, [graph, setNodes, setEdges]);

  const handleSeed = async () => {
      try {
          const now = new Date();
          // Seed some dummy traffic
          await apiClient.seedTrafficData([
              { time: now.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'}), total: 50, errors: 2, latency: 150 },
              { time: "12:00", total: 100, errors: 0, latency: 50 } // Some past data
          ]);
          toast({ title: "Traffic Seeded", description: "Injected sample traffic data." });
          refresh();
      } catch (e) {
          toast({ title: "Seed Failed", description: String(e), variant: "destructive" });
      }
  };

  return (
    <div className="h-[calc(100vh-8rem)] w-full relative bg-background border rounded-lg overflow-hidden shadow-sm">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onPaneClick={onPaneClick}
        nodeTypes={nodeTypes}
        fitView
        attributionPosition="bottom-left"
        className="bg-muted/10"
      >
        <Controls />
        <MiniMap />
        <Background gap={12} size={1} />

        <Panel position="top-right" className="flex gap-2">
            <Card className="p-2 flex gap-2 items-center bg-background/80 backdrop-blur-sm">
                <Button
                    variant={isLive ? "default" : "outline"}
                    size="sm"
                    onClick={() => setIsLive(!isLive)}
                    className="gap-2"
                >
                    <RefreshCw className={`h-4 w-4 ${isLive ? "animate-spin" : ""}`} />
                    {isLive ? "Live" : "Paused"}
                </Button>
                <div className="w-px h-6 bg-border mx-1" />
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={refresh}
                    title="Refresh Once"
                >
                    Refresh
                </Button>
                 <div className="w-px h-6 bg-border mx-1" />
                 {/* Developer Tool: Seed Data */}
                 <Button
                    variant="outline"
                    size="sm"
                    onClick={handleSeed}
                    title="Seed Debug Traffic"
                    className="text-xs"
                >
                    <Database className="h-3 w-3 mr-1" /> Seed
                </Button>
            </Card>
        </Panel>
      </ReactFlow>

      <VariableInspector selectedNode={selectedNode} onClose={() => setSelectedNode(null)} />
    </div>
  );
}
