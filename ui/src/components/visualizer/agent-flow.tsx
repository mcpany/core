/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useCallback, useState, useEffect } from 'react';
import {
  ReactFlow,
  Controls,
  Background,
  MiniMap,
  Node,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Card } from '@/components/ui/card';
import { DebuggerControls } from './debugger-controls';
import { VariableInspector } from './variable-inspector';
import { useNetworkTopology } from '@/hooks/use-network-topology';
import { Button } from '@/components/ui/button';
import { LayoutDashboard, RefreshCw } from 'lucide-react';

const nodeTypes = {
  // Mapping topology types to visual components
  // We use 'default' usually, but custom types allow better styling
  'NODE_TYPE_CLIENT': UserNode,
  'NODE_TYPE_CORE': AgentNode, // Represent Core as the "Agent/Orchestrator"
  'NODE_TYPE_SERVICE': ServiceNode,
  'NODE_TYPE_TOOL': ToolNode,
  'NODE_TYPE_RESOURCE': ResourceNode,
  // Fallbacks
  'default': ServiceNode,
  'user': UserNode,
  'agent': AgentNode,
  'tool': ToolNode,
  'service': ServiceNode,
};

/**
 * AgentFlow component renders the interactive flow visualization.
 * @returns The AgentFlow component.
 */
export function AgentFlow() {
  const {
      nodes,
      edges,
      onNodesChange,
      onEdgesChange,
      onConnect,
      refreshTopology,
      autoLayout
  } = useNetworkTopology();

  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [isPlaying, setIsPlaying] = useState(true); // Auto-refresh is on by default in context

  const onNodeClick = useCallback((_: any, node: Node) => {
    setSelectedNode(node);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  // Map nodes to have correct 'type' property for React Flow based on data.type
  const mappedNodes = nodes.map(n => ({
      ...n,
      type: n.data.type as string || 'default'
  }));

  return (
    <div className="h-[calc(100vh-8rem)] w-full relative bg-background border rounded-lg overflow-hidden shadow-sm">
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        <Card className="p-2 flex gap-2 items-center bg-background/80 backdrop-blur-sm">
          <Button variant="ghost" size="icon" onClick={refreshTopology} title="Refresh Topology">
              <RefreshCw className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" onClick={autoLayout} title="Auto Layout">
              <LayoutDashboard className="h-4 w-4" />
          </Button>
          <div className="w-px h-6 bg-border mx-1" />
          <DebuggerControls
            isPlaying={isPlaying}
            onPlayPause={() => setIsPlaying(!isPlaying)}
            onStep={refreshTopology}
            onStop={() => setIsPlaying(false)}
          />
        </Card>
      </div>

      <VariableInspector selectedNode={selectedNode} onClose={() => setSelectedNode(null)} />

      <ReactFlow
        nodes={mappedNodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
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
