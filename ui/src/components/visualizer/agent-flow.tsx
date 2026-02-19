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
  Node,
  Edge
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Card } from '@/components/ui/card';
import { DebuggerControls } from './debugger-controls';
import { VariableInspector } from './variable-inspector';
import { getLayoutedElements } from '@/lib/graph-layout';
import { Trace } from '@/types/trace';
import { Loader2 } from 'lucide-react';

const nodeTypes = {
  user: UserNode,
  agent: AgentNode,
  tool: ToolNode,
  resource: ResourceNode,
  service: ServiceNode,
};

/**
 * AgentFlow component renders the interactive flow visualization using real trace data.
 * @returns The AgentFlow component.
 */
export function AgentFlow() {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [isPlaying, setIsPlaying] = useState(true); // Auto-play by default for "Live" feel
  const [selectedNode, setSelectedNode] = useState<any>(null);
  const [traces, setTraces] = useState<Trace[]>([]);
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Poll for traces
  useEffect(() => {
    const fetchTraces = async () => {
      try {
        const res = await fetch('/api/traces');
        if (res.ok) {
            const data: Trace[] = await res.json();
            setTraces(data);

            // If we have no selected trace, or if "isPlaying" is true, select the latest one
            if (data.length > 0) {
                 if (!selectedTraceId || isPlaying) {
                     // Find the most recent trace
                     // data is already sorted by timestamp desc from API
                     const latest = data[0];
                     if (latest.id !== selectedTraceId) {
                         setSelectedTraceId(latest.id);
                     }
                 }
            }
        }
      } catch (err) {
        console.error("Failed to fetch traces", err);
      } finally {
        setIsLoading(false);
      }
    };

    fetchTraces();
    const interval = setInterval(fetchTraces, 3000);
    return () => clearInterval(interval);
  }, [isPlaying, selectedTraceId]);

  // Update Graph when selected trace changes
  useEffect(() => {
      if (!selectedTraceId) {
          setNodes([]);
          setEdges([]);
          return;
      }

      const trace = traces.find(t => t.id === selectedTraceId);
      if (trace) {
          const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(trace, 'LR');
          setNodes(layoutedNodes);
          setEdges(layoutedEdges);
      }
  }, [selectedTraceId, traces, setNodes, setEdges]);

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

  return (
    <div className="h-[calc(100vh-8rem)] w-full relative bg-background border rounded-lg overflow-hidden shadow-sm">
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        <Card className="p-2 flex gap-2 items-center bg-background/80 backdrop-blur-sm">
          <DebuggerControls
            isPlaying={isPlaying}
            onPlayPause={togglePlay}
            onStep={() => { }}
            onStop={() => setIsPlaying(false)}
          />
          <div className="w-px h-6 bg-border mx-1" />
           <div className="flex items-center gap-2">
               {isLoading && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
               <Select
                value={selectedTraceId || ""}
                onValueChange={(val) => {
                    setSelectedTraceId(val);
                    setIsPlaying(false); // Pause auto-update if user manually selects
                }}
               >
                <SelectTrigger className="w-[200px] h-8 text-xs font-mono">
                  <SelectValue placeholder="Select Trace" />
                </SelectTrigger>
                <SelectContent>
                  {traces.map(t => (
                      <SelectItem key={t.id} value={t.id} className="text-xs font-mono">
                          {new Date(t.timestamp).toLocaleTimeString()} - {t.rootSpan.name}
                      </SelectItem>
                  ))}
                  {traces.length === 0 && <SelectItem value="none" disabled>No traces found</SelectItem>}
                </SelectContent>
              </Select>
           </div>
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

        {nodes.length === 0 && !isLoading && (
            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <div className="text-muted-foreground text-center">
                    <p className="text-lg font-medium">No active flow</p>
                    <p className="text-sm">Run a tool in the Playground to see it visualized here.</p>
                </div>
            </div>
        )}
      </ReactFlow>
    </div>
  );
}
