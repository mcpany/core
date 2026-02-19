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
  MarkerType,
  Node,
  Edge,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { DebuggerControls } from './debugger-controls';
import { VariableInspector } from './variable-inspector';
import { getLayoutedElements } from '@/lib/graph-layout';
import { Trace, Span } from '@/types/trace';
import { RefreshCcw } from 'lucide-react';

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
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [isPlaying, setIsPlaying] = useState(false);
  const [selectedNode, setSelectedNode] = useState<any>(null);
  const [traces, setTraces] = useState<Trace[]>([]);
  const [selectedTraceId, setSelectedTraceId] = useState<string>("");

  // Fetch traces (poll if playing)
  const fetchTraces = useCallback(async () => {
      try {
          const res = await fetch('/api/traces');
          if (res.ok) {
              const data = await res.json();
              setTraces(data);
              // Select the first one if none selected or if we want to "follow" the live stream (implied by isPlaying)
              // If playing, we always switch to latest.
              if ((data.length > 0 && !selectedTraceId) || (isPlaying && data.length > 0 && data[0].id !== selectedTraceId)) {
                  setSelectedTraceId(data[0].id);
              }
          }
      } catch (e) {
          console.error("Failed to fetch traces", e);
      }
  }, [selectedTraceId, isPlaying]);

  useEffect(() => {
      fetchTraces();
  }, []);

  useEffect(() => {
      let interval: NodeJS.Timeout;
      if (isPlaying) {
          interval = setInterval(fetchTraces, 2000);
      }
      return () => clearInterval(interval);
  }, [isPlaying, fetchTraces]);


  // Convert Trace to Graph
  useEffect(() => {
      const trace = traces.find(t => t.id === selectedTraceId) || traces[0];
      if (!trace) {
          setNodes([]);
          setEdges([]);
          return;
      }

      const newNodes: Node[] = [];
      const newEdges: Edge[] = [];
      const nodeMap = new Map<string, Node>();

      // Helper to get or create node
      const getOrCreateNode = (span: Span, role: string, parentNodeId?: string): string => {

          let type = 'agent';
          let label = span.name;

          if (span.type === 'tool') type = 'tool';
          else if (span.type === 'service') type = 'service';
          else if (span.type === 'resource') type = 'resource';
          else if (span.name === 'User') type = 'user';

          // Special case for Core
          if (role === 'core') {
              type = 'agent';
              label = 'MCP Core';
          }

          const id = span.id;

          if (!nodeMap.has(id)) {
               const node: Node = {
                  id,
                  type,
                  position: { x: 0, y: 0 }, // Layout will fix this
                  data: {
                      label,
                      role: type === 'agent' ? 'Orchestrator' : undefined,
                      status: span.status === 'pending' ? 'Running' : span.status,
                      input: span.input,
                      output: span.output,
                      errorMessage: span.errorMessage
                  },
              };
              newNodes.push(node);
              nodeMap.set(id, node);
          }
          return id;
      };

      // Traverse Trace
      const traverse = (span: Span, parentId?: string) => {
          const nodeId = getOrCreateNode(span, parentId ? 'child' : 'core');

          if (parentId) {
               newEdges.push({
                  id: `e-${parentId}-${nodeId}`,
                  source: parentId,
                  target: nodeId,
                  animated: true,
                  label: span.name, // Edge label is the action
                  markerEnd: { type: MarkerType.ArrowClosed },
                  style: span.status === 'error' ? { stroke: '#ef4444' } : { stroke: '#64748b' }
               });
          }

          if (span.children) {
              span.children.forEach(child => traverse(child, nodeId));
          }
      };

      // Add a virtual "User" node as root trigger
      const userNode: Node = {
          id: 'user-trigger',
          type: 'user',
          position: { x: 0, y: 0 },
          data: { label: 'User' }
      };
      newNodes.push(userNode);
      nodeMap.set('user-trigger', userNode);

      // Start traversal from trace root, connecting to User
      const rootId = getOrCreateNode(trace.rootSpan, 'core');
      newEdges.push({
          id: `e-user-${rootId}`,
          source: 'user-trigger',
          target: rootId,
          animated: true,
          label: 'Request',
          markerEnd: { type: MarkerType.ArrowClosed }
      });

      if (trace.rootSpan.children) {
           trace.rootSpan.children.forEach(child => traverse(child, rootId));
      }

      // Apply Layout
      const layouted = getLayoutedElements(newNodes, newEdges);
      setNodes(layouted.nodes);
      setEdges(layouted.edges);

  }, [selectedTraceId, traces, setNodes, setEdges]);

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
            onPlayPause={() => setIsPlaying(!isPlaying)}
            onStep={() => { }}
            onStop={() => setIsPlaying(false)}
          />
          <div className="w-px h-6 bg-border mx-1" />
          <Select value={selectedTraceId} onValueChange={setSelectedTraceId}>
            <SelectTrigger className="w-[200px] h-8 text-xs font-mono">
              <SelectValue placeholder="Select Trace" />
            </SelectTrigger>
            <SelectContent>
              {traces.map(t => (
                  <SelectItem key={t.id} value={t.id} className="text-xs font-mono">
                      {new Date(t.timestamp).toLocaleTimeString()} - {t.rootSpan.name}
                  </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button variant="ghost" size="icon" onClick={() => fetchTraces()} className="h-8 w-8">
             <RefreshCcw className="h-4 w-4" />
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
