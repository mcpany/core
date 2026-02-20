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
  Controls,
  Background,
  MiniMap,
  Node,
  Edge,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Card } from '@/components/ui/card';
import { DebuggerControls } from './debugger-controls';
import { VariableInspector } from './variable-inspector';
import { Trace } from '@/types/trace';
import { traceToGraph } from '@/lib/flow-layout';
import { Badge } from '@/components/ui/badge';
import { Loader2 } from 'lucide-react';

const nodeTypes = {
  user: UserNode,
  agent: AgentNode,
  tool: ToolNode,
  resource: ResourceNode,
  service: ServiceNode,
};

/**
 * AgentFlow component renders the interactive flow visualization.
 * It fetches the latest trace from the backend and displays it as a graph.
 * @returns The AgentFlow component.
 */
export function AgentFlow() {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [isLive, setIsLive] = useState(true); // Default to live
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [loading, setLoading] = useState(true);
  const [lastTraceId, setLastTraceId] = useState<string | null>(null);
  const [lastTraceStatus, setLastTraceStatus] = useState<string | null>(null);

  const fetchLatestTrace = async () => {
      try {
          const res = await fetch('/api/traces');
          if (!res.ok) return;
          const data: Trace[] = await res.json();
          if (data && data.length > 0) {
              // Get the most recent trace
              const latest = data[0];
              // Only update if trace ID changed to avoid resetting layout/zoom constantly
              // Or should we update even if ID is same but status changed?
              if (latest.id !== lastTraceId || latest.status !== lastTraceStatus) {
                  setLastTraceId(latest.id);
                  setLastTraceStatus(latest.status);
                  const { nodes: newNodes, edges: newEdges } = traceToGraph(latest);
                  setNodes(newNodes);
                  setEdges(newEdges);
              }
          } else {
              // No traces
              if (nodes.length > 0) {
                 setNodes([]);
                 setEdges([]);
              }
          }
      } catch (e) {
          console.error("Failed to fetch traces", e);
      } finally {
          setLoading(false);
      }
  };

  useEffect(() => {
      fetchLatestTrace(); // Initial load
  }, []);

  useEffect(() => {
      let interval: NodeJS.Timeout;
      if (isLive) {
          interval = setInterval(fetchLatestTrace, 3000);
      }
      return () => clearInterval(interval);
  }, [isLive, lastTraceId, lastTraceStatus]);

  const onNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
    setSelectedNode(node);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  return (
    <div className="h-[calc(100vh-8rem)] w-full relative bg-background border rounded-lg overflow-hidden shadow-sm">
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        <Card className="p-2 flex gap-2 items-center bg-background/80 backdrop-blur-sm">
          <div className="flex items-center gap-2 px-2">
             <Badge variant={isLive ? "default" : "secondary"} className="cursor-pointer" onClick={() => setIsLive(!isLive)}>
                 {isLive ? "LIVE" : "PAUSED"}
             </Badge>
             {loading && <Loader2 className="h-3 w-3 animate-spin text-muted-foreground" />}
          </div>
          <div className="w-px h-6 bg-border mx-1" />
           {/* Reusing DebuggerControls for Play/Pause visual consistency, but wiring to isLive */}
          <DebuggerControls
            isPlaying={isLive}
            onPlayPause={() => setIsLive(!isLive)}
            onStep={() => { fetchLatestTrace() }} // Manual refresh
            onStop={() => { setIsLive(false); }}
          />
        </Card>
      </div>

      <VariableInspector selectedNode={selectedNode} onClose={() => setSelectedNode(null)} />

      {loading && nodes.length === 0 ? (
          <div className="absolute inset-0 flex items-center justify-center z-0">
              <div className="flex flex-col items-center gap-2 text-muted-foreground">
                  <Loader2 className="h-8 w-8 animate-spin" />
                  <p>Waiting for agent activity...</p>
              </div>
          </div>
      ) : (
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
      )}
    </div>
  );
}
