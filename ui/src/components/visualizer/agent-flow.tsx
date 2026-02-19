/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useCallback, useState, useEffect, useMemo } from 'react';
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
  Node,
  Edge,
  useReactFlow,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import dagre from 'dagre';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Card } from '@/components/ui/card';
import { DebuggerControls } from './debugger-controls';
import { VariableInspector } from './variable-inspector';
import { useTopology, NodeType, Node as TopologyNode } from "@/hooks/use-topology";
import { Loader2 } from "lucide-react";

const nodeTypes = {
  user: UserNode,
  agent: AgentNode, // Map Client to AgentNode for now
  tool: ToolNode,
  resource: ResourceNode,
  service: ServiceNode,
  core: ServiceNode, // Use ServiceNode for Core for now, or create custom
};

const getLayoutedElements = (nodes: Node[], edges: Edge[]) => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const nodeWidth = 180;
  const nodeHeight = 80;

  dagreGraph.setGraph({ rankdir: 'LR' });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    };
  });

  return { nodes: layoutedNodes, edges };
};

/**
 * AgentFlow component renders the interactive flow visualization.
 * @returns The AgentFlow component.
 */
export function AgentFlow() {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [isLive, setIsLive] = useState(false);
  const [selectedNode, setSelectedNode] = useState<any>(null);

  // Use topology hook with polling if live
  const { graph, loading, refresh } = useTopology(isLive ? 3000 : null);

  const transformGraph = useCallback((graphData: any) => {
      if (!graphData) return { nodes: [], edges: [] };

      const flowNodes: Node[] = [];
      const flowEdges: Edge[] = [];

      // Add Core
      if (graphData.core) {
          flowNodes.push({
              id: graphData.core.id,
              type: 'service', // Reuse service style for core
              data: { label: 'MCP Gateway', role: 'Core', status: 'Active' },
              position: { x: 0, y: 0 }
          });

          // Core Children (Services)
          if (graphData.core.children) {
              graphData.core.children.forEach((svc: TopologyNode) => {
                  flowNodes.push({
                      id: svc.id,
                      type: 'service',
                      data: { label: svc.label, role: 'Service', status: 'Active' },
                      position: { x: 0, y: 0 }
                  });
                  flowEdges.push({
                      id: `e-${graphData.core.id}-${svc.id}`,
                      source: graphData.core.id,
                      target: svc.id,
                      animated: true,
                      type: 'smoothstep',
                  });

                  // Service Children (Tools)
                  if (svc.children) {
                      svc.children.forEach((tool: TopologyNode) => {
                          flowNodes.push({
                              id: tool.id,
                              type: 'tool',
                              data: { label: tool.label },
                              position: { x: 0, y: 0 }
                          });
                          flowEdges.push({
                              id: `e-${svc.id}-${tool.id}`,
                              source: svc.id,
                              target: tool.id,
                              type: 'smoothstep',
                          });
                      });
                  }
              });
          }
      }

      // Add Clients
      if (graphData.clients) {
          graphData.clients.forEach((client: TopologyNode) => {
              flowNodes.push({
                  id: client.id,
                  type: 'user', // Or agent
                  data: { label: client.label || 'Client' },
                  position: { x: 0, y: 0 }
              });
              // Connect to Core
              if (graphData.core) {
                  flowEdges.push({
                      id: `e-${client.id}-${graphData.core.id}`,
                      source: client.id,
                      target: graphData.core.id,
                      animated: true,
                      markerEnd: { type: MarkerType.ArrowClosed },
                      type: 'smoothstep',
                  });
              }
          });
      }

      return getLayoutedElements(flowNodes, flowEdges);
  }, []);

  useEffect(() => {
      if (graph) {
          const { nodes: layoutedNodes, edges: layoutedEdges } = transformGraph(graph);
          setNodes(layoutedNodes);
          setEdges(layoutedEdges);
      }
  }, [graph, transformGraph, setNodes, setEdges]);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge({ ...params, animated: true }, eds)),
    [setEdges],
  );

  const togglePlay = () => setIsLive(!isLive);

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
            onPlayPause={togglePlay}
            onStep={() => refresh()}
            onStop={() => setIsLive(false)}
          />
          <div className="w-px h-6 bg-border mx-1" />
          <div className="flex items-center gap-2 px-2 text-xs text-muted-foreground">
              {loading && <Loader2 className="h-3 w-3 animate-spin" />}
              {graph?.clients?.length || 0} Clients
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
      </ReactFlow>
    </div>
  );
}
