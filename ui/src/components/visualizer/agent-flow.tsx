/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useCallback, useState } from 'react';
import {
  ReactFlow,
  useNodesState,
  useEdgesState,
  addEdge,
  Controls,
  Background,
  MiniMap,
  Connection,
  MarkerType,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Card } from '@/components/ui/card';
import { DebuggerControls } from './debugger-controls';
import { VariableInspector } from './variable-inspector';

const nodeTypes = {
  user: UserNode,
  agent: AgentNode,
  tool: ToolNode,
  resource: ResourceNode,
  service: ServiceNode,
};

const initialNodes = [
  { id: '1', type: 'user', position: { x: 250, y: 0 }, data: { label: 'Alice' } },
  { id: '2', type: 'agent', position: { x: 250, y: 150 }, data: { label: 'Orchestrator', role: 'Main Agent', status: 'Thinking...' } },
  { id: '3', type: 'service', position: { x: 100, y: 300 }, data: { label: 'Postgres DB' } },
  { id: '4', type: 'tool', position: { x: 400, y: 300 }, data: { label: 'Web Search' } },
];

const initialEdges = [
  { id: 'e1-2', source: '1', target: '2', animated: true, label: 'Task: Analyze Data', markerEnd: { type: MarkerType.ArrowClosed } },
  { id: 'e2-3', source: '2', target: '3', animated: true, label: 'Query' },
  { id: 'e2-4', source: '2', target: '4', animated: false, label: 'Search' },
];

/**
 * AgentFlow component renders the interactive flow visualization.
 * @returns The AgentFlow component.
 */
export function AgentFlow() {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [isPlaying, setIsPlaying] = useState(false);
  const [selectedNode, setSelectedNode] = useState<any>(null);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge({ ...params, animated: true }, eds)),
    [setEdges],
  );

  const togglePlay = () => setIsPlaying(!isPlaying);

  const onNodeClick = useCallback((_: any, node: any) => {
    setSelectedNode(node);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  // Simulation effect (placeholder for real "Live" data)
  React.useEffect(() => {
    if (!isPlaying) return;
    const interval = setInterval(() => {
      setEdges((eds) => eds.map(e => ({
        ...e,
        animated: !e.animated || Math.random() > 0.5,
        style: { stroke: Math.random() > 0.5 ? '#22c55e' : '#64748b' } // Green or slate
      })));
    }, 1000);
    return () => clearInterval(interval);
  }, [isPlaying, setEdges]);

  return (
    <div className="h-[calc(100vh-8rem)] w-full relative bg-background border rounded-lg overflow-hidden shadow-sm">
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        <Card className="p-2 flex gap-2 items-center bg-background/80 backdrop-blur-sm">
          <DebuggerControls
            isPlaying={isPlaying}
            onPlayPause={togglePlay}
            onStep={() => { }} // Placeholder for step
            onStop={() => { setIsPlaying(false); setNodes(initialNodes); setEdges(initialEdges); }}
          />
          <div className="w-px h-6 bg-border mx-1" />
          <Select defaultValue="demo1">
            <SelectTrigger className="w-[140px] h-8">
              <SelectValue placeholder="Scenario" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="demo1">Basic Flow</SelectItem>
              <SelectItem value="demo2">Multi-Agent</SelectItem>
            </SelectContent>
          </Select>
        </Card>
      </div>

      <VariableInspector selectedNode={selectedNode} onClose={() => setSelectedNode(null)} />

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
