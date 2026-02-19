/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useCallback, useState } from 'react';
import {
  ReactFlow,
  Controls,
  Background,
  MiniMap,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { VariableInspector } from './variable-inspector';
import { useRealTimeTopology } from '@/hooks/use-real-time-topology';
import { Play, Pause, RefreshCw, Database } from 'lucide-react';
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
 * AgentFlow component renders the interactive flow visualization.
 * Renamed conceptually to NetworkTopology but keeping component name for compatibility.
 * @returns The AgentFlow component.
 */
export function AgentFlow() {
  const { nodes, edges, onNodesChange, onEdgesChange, isLive, setIsLive, refresh, lastUpdated } = useRealTimeTopology();
  const [selectedNode, setSelectedNode] = useState<any>(null);
  const { toast } = useToast();
  const [seeding, setSeeding] = useState(false);

  const onNodeClick = useCallback((_: any, node: any) => {
    setSelectedNode(node);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  const handleSeedData = async () => {
      setSeeding(true);
      try {
          // Seed some dummy traffic
          const points = [
              { time: "10:00", total: 100, errors: 2, latency: 50 },
              { time: "10:01", total: 120, errors: 0, latency: 45 },
          ];
          await apiClient.seedTrafficData(points);
          toast({ title: "Traffic Seeded", description: "Injected sample traffic data." });
          refresh();
      } catch (e) {
          toast({ title: "Seeding Failed", variant: "destructive", description: String(e) });
      } finally {
          setSeeding(false);
      }
  };

  return (
    <div className="h-[calc(100vh-8rem)] w-full relative bg-background border rounded-lg overflow-hidden shadow-sm">
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        <Card className="p-2 flex gap-4 items-center bg-background/80 backdrop-blur-sm shadow-sm border">
          <div className="flex items-center gap-2">
              <Switch id="live-mode" checked={isLive} onCheckedChange={setIsLive} />
              <Label htmlFor="live-mode" className="text-xs font-medium cursor-pointer flex items-center gap-1">
                  {isLive ? <Play className="h-3 w-3 text-green-500" /> : <Pause className="h-3 w-3 text-muted-foreground" />}
                  Live
              </Label>
          </div>

          <div className="h-4 w-px bg-border" />

          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={refresh} title="Refresh">
              <RefreshCw className="h-4 w-4" />
          </Button>

          {/* Dev / Demo Only */}
          <Button
            variant="outline"
            size="sm"
            className="h-8 text-xs gap-1"
            onClick={handleSeedData}
            disabled={seeding}
            title="Inject fake traffic for demo"
          >
              <Database className="h-3 w-3" />
              Seed Data
          </Button>
        </Card>
      </div>

      <div className="absolute bottom-4 left-4 z-10">
          <div className="text-[10px] text-muted-foreground bg-background/50 backdrop-blur px-2 py-1 rounded border">
              Last Updated: {lastUpdated ? lastUpdated.toLocaleTimeString() : 'Never'}
          </div>
      </div>

      <VariableInspector selectedNode={selectedNode} onClose={() => setSelectedNode(null)} />

      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onPaneClick={onPaneClick}
        nodeTypes={nodeTypes}
        fitView
        attributionPosition="bottom-right"
        className="bg-muted/5"
      >
        <Controls />
        <MiniMap />
        <Background gap={12} size={1} />
      </ReactFlow>
    </div>
  );
}
