/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useCallback, useState, useEffect } from 'react';
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
import { Card } from '@/components/ui/card';
import { VariableInspector } from './variable-inspector';
import { convertTraceToGraph } from '@/lib/trace-to-graph';
import { Trace } from '@/types/trace';
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";

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
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [selectedNode, setSelectedNode] = useState<any>(null);
  const [isLive, setIsLive] = useState(true);
  const [lastTraceId, setLastTraceId] = useState<string | null>(null);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge({ ...params, animated: true }, eds)),
    [setEdges],
  );

  const onNodeClick = useCallback((_: any, node: any) => {
    setSelectedNode(node);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  const fetchLatestTrace = async () => {
      try {
          const res = await fetch('/api/traces');
          if (!res.ok) return;

          const data: Trace[] = await res.json();
          if (data.length > 0) {
              // Filter out noisy traces (polls) to show the most relevant recent activity
              // We want to see tool executions, not just health checks
              const ignoredPatterns = ['/debug/entries', '/api/v1/doctor', '/api/v1/topology', '/api/v1/tools', '/api/v1/metrics'];
              const interesting = data.find(t => {
                  const name = t.rootSpan.name;
                  return !ignoredPatterns.some(pattern => name.includes(pattern));
              });

              // If we found an interesting trace, use it. Otherwise fallback to latest (e.g. if nothing happened yet)
              // But if we have an existing trace and the new latest is noise, keep the existing one!
              const target = interesting || (lastTraceId ? null : data[0]);

              if (target && target.id !== lastTraceId) {
                  setLastTraceId(target.id);
                  const { nodes: newNodes, edges: newEdges } = convertTraceToGraph(target);
                  setNodes(newNodes);
                  setEdges(newEdges);
              }
          }
      } catch (e) {
          console.error("Failed to fetch traces", e);
      }
  };

  useEffect(() => {
      fetchLatestTrace(); // Initial load
      let interval: NodeJS.Timeout;
      if (isLive) {
          interval = setInterval(fetchLatestTrace, 2000);
      }
      return () => clearInterval(interval);
  }, [isLive, lastTraceId]);

  return (
    <div className="h-[calc(100vh-8rem)] w-full relative bg-background border rounded-lg overflow-hidden shadow-sm">
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        <Card className="p-2 flex gap-4 items-center bg-background/80 backdrop-blur-sm shadow-sm border">
           <div className="flex items-center gap-2">
                <Switch id="live-mode" checked={isLive} onCheckedChange={setIsLive} />
                <Label htmlFor="live-mode" className="text-xs font-medium cursor-pointer flex items-center gap-1">
                    {isLive ? <span className="flex h-2 w-2 rounded-full bg-green-500 animate-pulse" /> : <span className="flex h-2 w-2 rounded-full bg-gray-300" />}
                    Live Mode
                </Label>
           </div>
           <div className="h-4 w-px bg-border" />
           <span className="text-xs text-muted-foreground font-mono">
                {lastTraceId ? `Trace: ${lastTraceId.substring(0, 8)}` : "Waiting for activity..."}
           </span>
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
