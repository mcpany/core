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
  Connection,
  addEdge,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Play, Pause, RefreshCw } from "lucide-react";
import { VariableInspector } from './variable-inspector';
import { Trace } from '@/types/trace';
import { traceToGraph } from '@/lib/visualizer-utils';
import { Loader2 } from "lucide-react";

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
  const [isLive, setIsLive] = useState(true);
  const [traces, setTraces] = useState<Trace[]>([]);
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Fetch traces
  const fetchTraces = async () => {
      try {
          const res = await fetch('/api/traces');
          if (res.ok) {
              const fullData: Trace[] = await res.json();
              // Limit to 50 latest traces to prevent UI lag and memory issues
              const data = fullData.slice(0, 50);
              setTraces(data);

              // If live, or no trace selected, select the first one
              if ((isLive || !selectedTraceId) && data.length > 0) {
                  // Prefer trace with 'execute' in rootSpan.name
                  const interestingTrace = data.find(t => t.rootSpan.name.includes('execute')) || data[0];

                  // If we are live, we always jump to latest (or interesting one).
                  // If not live, we only set if nothing selected.
                  if (isLive) {
                      setSelectedTraceId(interestingTrace.id);
                  } else if (!selectedTraceId) {
                       setSelectedTraceId(interestingTrace.id);
                  }
              }
          }
      } catch (err) {
          console.error("Failed to fetch traces", err);
      } finally {
          setIsLoading(false);
      }
  };

  useEffect(() => {
      fetchTraces();
  }, []);

  useEffect(() => {
      let interval: NodeJS.Timeout;
      if (isLive) {
          // Poll every 5 seconds instead of 2 seconds to reduce backend load in CI
          interval = setInterval(fetchTraces, 5000);
      }
      return () => clearInterval(interval);
  }, [isLive, selectedTraceId]);

  // Update graph when selected trace changes
  useEffect(() => {
      if (selectedTraceId) {
          const trace = traces.find(t => t.id === selectedTraceId);
          if (trace) {
              const { nodes: newNodes, edges: newEdges } = traceToGraph(trace);
              setNodes(newNodes);
              setEdges(newEdges);
          }
      } else {
          setNodes([]);
          setEdges([]);
      }
  }, [selectedTraceId, traces, setNodes, setEdges]);

  const onNodeClick = useCallback((_: any, node: any) => {
    // Find the span info related to this node
    // The node data has label, role.
    // Ideally we pass the full span or object to data.
    // For now, VariableInspector expects simple data.
    // We can enhance this later.
    setSelectedNode(node);
  }, []);

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  const handleTraceChange = (value: string) => {
      setSelectedTraceId(value);
      if (value !== traces[0]?.id) {
          setIsLive(false); // Disable live if user selects old trace
      }
  };

  return (
    <div className="h-[calc(100vh-8rem)] w-full relative bg-background border rounded-lg overflow-hidden shadow-sm">
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        <Card className="p-2 flex gap-2 items-center bg-background/80 backdrop-blur-sm">
           <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={() => setIsLive(!isLive)}
            title={isLive ? "Pause Live Updates" : "Resume Live Updates"}
          >
            {isLive ? <Pause className="h-4 w-4 text-primary" /> : <Play className="h-4 w-4" />}
          </Button>

          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={fetchTraces}
            title="Refresh"
          >
             <RefreshCw className="h-4 w-4" />
          </Button>

          <div className="w-px h-6 bg-border mx-1" />

          <Select value={selectedTraceId || ""} onValueChange={handleTraceChange}>
            <SelectTrigger className="w-[200px] h-8 text-xs font-mono">
              <SelectValue placeholder={isLoading ? "Loading..." : "Select Trace"} />
            </SelectTrigger>
            <SelectContent>
              {traces.length === 0 ? (
                  <SelectItem value="none" disabled>No traces found</SelectItem>
              ) : (
                  traces.map(t => (
                      <SelectItem key={t.id} value={t.id} className="text-xs font-mono">
                          {new Date(t.timestamp).toLocaleTimeString()} - {t.rootSpan.name}
                      </SelectItem>
                  ))
              )}
            </SelectContent>
          </Select>
        </Card>
      </div>

      <VariableInspector selectedNode={selectedNode} onClose={() => setSelectedNode(null)} />

      {isLoading && nodes.length === 0 ? (
          <div className="absolute inset-0 flex items-center justify-center bg-background/50 z-20">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      ) : nodes.length === 0 ? (
           <div className="absolute inset-0 flex items-center justify-center bg-background/50 z-0 pointer-events-none">
              <div className="text-center">
                  <p className="text-muted-foreground font-medium">No activity recorded</p>
                  <p className="text-sm text-muted-foreground/60">Run a tool in the Playground to see the flow.</p>
              </div>
          </div>
      ) : null}

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
