/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useCallback, useEffect, useState, useMemo } from 'react';
import {
  ReactFlow,
  useNodesState,
  useEdgesState,
  Controls,
  Background,
  MiniMap,
  Node,
  Edge,
  MarkerType,
  Connection,
  addEdge,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import dagre from 'dagre';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { RefreshCw, Play, Pause } from 'lucide-react';
import { useTopology, TopologyGraph, TopologyNode } from '@/hooks/use-topology';
import { VariableInspector } from './variable-inspector';

const nodeTypes = {
  user: UserNode,
  agent: AgentNode,
  tool: ToolNode,
  resource: ResourceNode,
  service: ServiceNode,
  // Mapping API call to ToolNode for now, or could use ResourceNode
  api_call: ResourceNode,
};

// Layout helper
const getLayoutedElements = (nodes: Node[], edges: Edge[]) => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const isHorizontal = true;
  dagreGraph.setGraph({ rankdir: isHorizontal ? 'LR' : 'TB' });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: 180, height: 80 });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    return {
      ...node,
      targetPosition: isHorizontal ? 'left' : 'top',
      sourcePosition: isHorizontal ? 'right' : 'bottom',
      // We are shifting the dagre node position (anchor=center center) to the top left
      // so it matches the React Flow node anchor point (top left).
      position: {
        x: nodeWithPosition.x - 180 / 2,
        y: nodeWithPosition.y - 80 / 2,
      },
    };
  });

  return { nodes: layoutedNodes, edges };
};

const mapNodeType = (type: string): string => {
    switch (type) {
        case 'NODE_TYPE_CLIENT': return 'user';
        case 'NODE_TYPE_CORE': return 'agent'; // Gateway acts as Agent
        case 'NODE_TYPE_SERVICE': return 'service';
        case 'NODE_TYPE_TOOL': return 'tool';
        case 'NODE_TYPE_API_CALL': return 'api_call';
        default: return 'service';
    }
};

const mapNodeLabel = (node: TopologyNode): string => {
    if (node.type === 'NODE_TYPE_CORE') return 'MCP Gateway';
    return node.label;
};

/**
 * AgentFlow component renders the live topology visualization.
 * @returns The AgentFlow component.
 */
export function AgentFlow() {
  const [isLive, setIsLive] = useState(false);
  const { graph, refresh, loading } = useTopology(isLive ? 5000 : null);

  // React Flow state
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [selectedNode, setSelectedNode] = useState<any>(null);

  // Transform graph to React Flow elements
  useEffect(() => {
      if (!graph) return;

      const newNodes: Node[] = [];
      const newEdges: Edge[] = [];

      // Helper to process node and its children recursively
      const processNode = (node: TopologyNode, parentId?: string) => {
          const flowType = mapNodeType(node.type);
          const flowLabel = mapNodeLabel(node);

          // Create Node
          newNodes.push({
              id: node.id,
              type: flowType,
              data: {
                  label: flowLabel,
                  status: node.status,
                  metrics: node.metrics,
                  metadata: node.metadata
              },
              position: { x: 0, y: 0 }, // Will be set by layout
          });

          // Create Edge from Parent
          if (parentId) {
              newEdges.push({
                  id: `e-${parentId}-${node.id}`,
                  source: parentId,
                  target: node.id,
                  animated: true,
                  style: { stroke: node.status === 'NODE_TYPE_ERROR' ? '#ef4444' : '#64748b' },
              });
          }

          // Process Children
          if (node.children) {
              node.children.forEach(child => processNode(child, node.id));
          }
      };

      // 1. Process Core and its subtree (Services -> Tools)
      if (graph.core) {
          processNode(graph.core);
      }

      // 2. Process Clients -> Core
      if (graph.clients) {
          graph.clients.forEach(client => {
              processNode(client);
              // Link Client to Core explicitly
              if (graph.core) {
                  newEdges.push({
                      id: `e-${client.id}-${graph.core.id}`,
                      source: client.id,
                      target: graph.core.id,
                      animated: true,
                      markerEnd: { type: MarkerType.ArrowClosed },
                  });
              }
          });
      }

      // Apply Layout
      const layouted = getLayoutedElements(newNodes, newEdges);
      setNodes(layouted.nodes as any);
      setEdges(layouted.edges);

  }, [graph, setNodes, setEdges]);

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
          <div className="flex items-center gap-2 mr-2">
              <Switch id="live-mode" checked={isLive} onCheckedChange={setIsLive} />
              <Label htmlFor="live-mode" className="text-xs font-medium flex items-center gap-1 cursor-pointer">
                  {isLive ? <Pause className="h-3 w-3 text-green-500" /> : <Play className="h-3 w-3" />}
                  Live Mode
              </Label>
          </div>
          <div className="w-px h-6 bg-border mx-1" />
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => refresh()} disabled={loading}>
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
        className="bg-muted/5"
      >
        <Controls />
        <MiniMap />
        <Background gap={12} size={1} />
      </ReactFlow>
    </div>
  );
}
