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
  Position,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Card } from '@/components/ui/card';
import { VariableInspector } from './variable-inspector';
import { useTopology, TopologyNode } from '@/hooks/use-topology';
import dagre from 'dagre';
import { Loader2, RefreshCw, Power } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';

const nodeTypes = {
  user: UserNode,
  agent: AgentNode,
  tool: ToolNode,
  resource: ResourceNode,
  service: ServiceNode,
  // Mapping
  CLIENT: UserNode,
  CORE: AgentNode,
  SERVICE: ServiceNode,
  TOOL: ToolNode,
  API_CALL: ResourceNode, // Use resource for generic API call or create new?
  MIDDLEWARE: ServiceNode, // Reuse service style for middleware
  WEBHOOK: ServiceNode,
};

const dagreGraph = new dagre.graphlib.Graph();
dagreGraph.setDefaultEdgeLabel(() => ({}));

const nodeWidth = 172;
const nodeHeight = 36;

const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'LR') => {
  dagreGraph.setGraph({ rankdir: direction });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    // We are shifting the dagre node position (anchor=center center) to the top left
    // so it matches the React Flow node anchor point (top left).
    return {
      ...node,
      targetPosition: direction === 'LR' ? Position.Left : Position.Top,
      sourcePosition: direction === 'LR' ? Position.Right : Position.Bottom,
      // We explicitly cast Position to any because React Flow types are strict
      // but Position enum matches strings 'left', 'right', etc.
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
  const { graph, loading, error, isPolling, setIsPolling, refresh } = useTopology();
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [selectedNode, setSelectedNode] = useState<any>(null);

  // Transform TopologyGraph to React Flow Nodes/Edges
  useEffect(() => {
    if (!graph) return;

    const newNodes: Node[] = [];
    const newEdges: Edge[] = [];

    const processNode = (node: TopologyNode, parentId?: string) => {
        // Map Type
        let type = 'service'; // default
        if (node.type === 'CLIENT') type = 'user';
        else if (node.type === 'CORE') type = 'agent';
        else if (node.type === 'TOOL') type = 'tool';
        else if (node.type === 'API_CALL') type = 'resource';
        else if (node.type === 'MIDDLEWARE') type = 'service';

        // Add Node
        newNodes.push({
            id: node.id,
            type: type,
            position: { x: 0, y: 0 }, // Position will be set by dagre
            data: {
                label: node.label,
                status: node.status,
                metrics: node.metrics,
                metadata: node.metadata
            }
        });

        // Add Edge from parent
        if (parentId) {
            const isTraffic = (node.metrics?.qps || 0) > 0;
            newEdges.push({
                id: `e-${parentId}-${node.id}`,
                source: parentId,
                target: node.id,
                animated: isTraffic,
                style: { stroke: isTraffic ? '#22c55e' : '#64748b', strokeWidth: isTraffic ? 2 : 1 },
                label: isTraffic ? `${node.metrics?.qps.toFixed(1)} QPS` : undefined
            });
        }

        // Process children
        if (node.children) {
            node.children.forEach(child => processNode(child, node.id));
        }
    };

    // 1. Add Core
    if (graph.core) {
        processNode(graph.core);
    } else {
        // Fallback for empty graph?
        return;
    }

    // 2. Add Clients and connect to Core
    if (graph.clients) {
        graph.clients.forEach(client => {
            processNode(client);
            // Explicitly connect Client -> Core
            const isTraffic = true; // Clients are usually active sessions
            newEdges.push({
                id: `e-${client.id}-${graph.core.id}`,
                source: client.id,
                target: graph.core.id,
                animated: true,
                style: { stroke: '#3b82f6' }
            });
        });
    }

    // Apply Layout
    const layout = getLayoutedElements(newNodes, newEdges);
    setNodes(layout.nodes);
    setEdges(layout.edges);

  }, [graph, setNodes, setEdges]);

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

  if (loading && !graph) {
      return (
          <div className="flex h-[500px] items-center justify-center text-muted-foreground">
              <Loader2 className="mr-2 h-8 w-8 animate-spin" />
              Loading topology...
          </div>
      );
  }

  if (error) {
      return (
          <div className="flex h-[500px] items-center justify-center text-destructive">
              Error: {error}
          </div>
      );
  }

  return (
    <div className="h-[calc(100vh-8rem)] w-full relative bg-background border rounded-lg overflow-hidden shadow-sm">
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        <Card className="p-2 flex gap-4 items-center bg-background/80 backdrop-blur-sm shadow-sm border-muted">
          <div className="flex items-center gap-2">
              <Switch id="live-mode" checked={isPolling} onCheckedChange={setIsPolling} />
              <Label htmlFor="live-mode" className="flex items-center gap-1 cursor-pointer">
                  {isPolling ? <div className="h-2 w-2 rounded-full bg-green-500 animate-pulse" /> : <Power className="h-3 w-3 text-muted-foreground" />}
                  Live
              </Label>
          </div>
          <div className="w-px h-6 bg-border" />
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => refresh()}>
              <RefreshCw className="h-4 w-4" />
          </Button>
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
        nodeTypes={nodeTypes as any}
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
