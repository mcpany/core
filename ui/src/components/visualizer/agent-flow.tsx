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
  Connection,
  addEdge,
  useEdgesState, // Still needed for onConnect if we want manual connections?
  // Probably not for topology view, but let's keep it robust.
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Card } from '@/components/ui/card';
import { DebuggerControls } from './debugger-controls';
import { VariableInspector } from './variable-inspector';
import { useTopology } from '@/hooks/use-topology';
import { Button } from '@/components/ui/button';
import { RefreshCw } from 'lucide-react';

const nodeTypes = {
  user: UserNode,
  agent: AgentNode,
  tool: ToolNode,
  resource: ResourceNode,
  service: ServiceNode,
};

/**
 * AgentFlow component renders the interactive flow visualization.
 * @returns The AgentFlow component.
 */
export function AgentFlow() {
  const [isLive, setIsLive] = useState(false);
  const { nodes, edges, onNodesChange, onEdgesChange, refresh, loading } = useTopology(isLive);
  const [selectedNode, setSelectedNode] = useState<any>(null);

  // We can still allow manual connections if needed, but topology is read-only usually.
  // We'll keep the handler but it might not be used.
  // Note: useTopology manages the state, so onConnect needs access to setEdges?
  // useTopology uses useEdgesState internally and exposes setEdges? No, strictly onEdgesChange.
  // If we want manual edges, we'd need to lift state up or expose setEdges.
  // For visualizer, read-only is fine.

  const onNodeClick = useCallback((_: any, node: any) => {
    setSelectedNode(node);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  return (
    <div className="h-[calc(100vh-8rem)] w-full relative bg-background border rounded-lg overflow-hidden shadow-sm">
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        <Card className="p-2 flex gap-2 items-center bg-background/80 backdrop-blur-sm">
          <DebuggerControls
            isPlaying={isLive}
            onPlayPause={() => setIsLive(!isLive)}
            onStep={refresh} // Step acts as manual refresh
            onStop={() => setIsLive(false)}
          />
          <div className="w-px h-6 bg-border mx-1" />
          <Button variant="ghost" size="icon" onClick={refresh} title="Refresh Topology">
             <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          </Button>
        </Card>
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
        attributionPosition="bottom-left"
        className="bg-muted/10"
      >
        <Controls />
        <MiniMap />
        <Background gap={12} size={1} />
      </ReactFlow>
    </div>
  );
}
