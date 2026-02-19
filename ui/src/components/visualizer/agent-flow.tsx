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
import dagre from 'dagre';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Card } from '@/components/ui/card';
import { DebuggerControls } from './debugger-controls';
import { VariableInspector } from './variable-inspector';
import { useTopology } from '@/hooks/use-topology';
import { NodeType, NodeStatus } from '@proto/topology/v1/topology';
import { Loader2, AlertCircle } from 'lucide-react';

const nodeTypes = {
  user: UserNode,
  client: UserNode, // Map client to UserNode style
  agent: AgentNode,
  tool: ToolNode,
  resource: ResourceNode,
  service: ServiceNode,
  core: ServiceNode, // Map core to ServiceNode style but maybe styled differently
  api_call: ResourceNode, // Mock API call node
  middleware: ServiceNode,
  webhook: ServiceNode,
};

const nodeWidth = 172;
const nodeHeight = 36;

const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'TB') => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const isHorizontal = direction === 'LR';
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
    // node.targetPosition = isHorizontal ? Position.Left : Position.Top;
    // node.sourcePosition = isHorizontal ? Position.Right : Position.Bottom;

    // Actually React Flow handles handles automatically if not specified,
    // but dagre gives center coordinates.
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
  const { graph, loading, error, isPaused, setIsPaused, refresh } = useTopology(3000); // Poll every 3s
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [selectedNode, setSelectedNode] = useState<any>(null);

  // Transform Proto Graph to React Flow
  useEffect(() => {
    if (!graph) return;

    const newNodes: Node[] = [];
    const newEdges: Edge[] = [];

    // Helper to map Proto Node Type to React Flow Type
    const mapType = (t: NodeType): string => {
        switch (t) {
            case NodeType.NODE_TYPE_CLIENT: return 'client';
            case NodeType.NODE_TYPE_CORE: return 'core';
            case NodeType.NODE_TYPE_SERVICE: return 'service';
            case NodeType.NODE_TYPE_TOOL: return 'tool';
            case NodeType.NODE_TYPE_RESOURCE: return 'resource';
            case NodeType.NODE_TYPE_API_CALL: return 'api_call';
            case NodeType.NODE_TYPE_MIDDLEWARE: return 'middleware';
            case NodeType.NODE_TYPE_WEBHOOK: return 'webhook';
            default: return 'service';
        }
    };

    // Helper to create node data
    const mapData = (n: any) => ({
        label: n.label,
        status: n.status === NodeStatus.NODE_STATUS_ACTIVE ? 'Active' :
                n.status === NodeStatus.NODE_STATUS_ERROR ? 'Error' : 'Inactive',
        metrics: n.metrics // Pass metrics if available
    });

    if (graph.core) {
        // Add Core
        newNodes.push({
            id: graph.core.id,
            type: 'agent', // Use Agent style for Core
            position: { x: 0, y: 0 },
            data: { ...mapData(graph.core), role: 'Gateway', label: 'MCP Any Core' }
        });

        // Add Children (Services, Middleware, etc)
        const traverse = (parent: any, children: any[]) => {
            if (!children) return;
            children.forEach(child => {
                newNodes.push({
                    id: child.id,
                    type: mapType(child.type),
                    position: { x: 0, y: 0 },
                    data: mapData(child)
                });
                newEdges.push({
                    id: `${parent.id}-${child.id}`,
                    source: parent.id,
                    target: child.id,
                    animated: true,
                    style: { stroke: '#64748b' }
                });
                if (child.children) {
                    traverse(child, child.children);
                }
            });
        };
        traverse(graph.core, graph.core.children);
    }

    if (graph.clients) {
        graph.clients.forEach(client => {
            newNodes.push({
                id: client.id,
                type: 'user',
                position: { x: 0, y: 0 },
                data: mapData(client)
            });
            // Connect Client to Core
            if (graph.core) {
                newEdges.push({
                    id: `${client.id}-${graph.core.id}`,
                    source: client.id,
                    target: graph.core.id,
                    animated: true,
                    markerEnd: { type: MarkerType.ArrowClosed },
                    style: { stroke: '#22c55e' } // Active traffic
                });
            }
        });
    }

    // Apply Layout
    const layouted = getLayoutedElements(newNodes, newEdges, 'TB');
    setNodes(layouted.nodes);
    setEdges(layouted.edges);

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
          <div className="h-[calc(100vh-8rem)] w-full flex items-center justify-center bg-muted/5 border rounded-lg">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              <span className="ml-2 text-muted-foreground">Loading topology...</span>
          </div>
      );
  }

  if (error && !graph) {
      return (
          <div className="h-[calc(100vh-8rem)] w-full flex items-center justify-center bg-muted/5 border rounded-lg">
              <div className="text-center text-destructive">
                  <AlertCircle className="h-8 w-8 mx-auto mb-2" />
                  <p>Failed to load topology.</p>
                  <p className="text-sm opacity-70">{error.message}</p>
              </div>
          </div>
      );
  }

  return (
    <div className="h-[calc(100vh-8rem)] w-full relative bg-background border rounded-lg overflow-hidden shadow-sm">
      <div className="absolute top-4 right-4 z-10 flex gap-2">
        <Card className="p-2 flex gap-2 items-center bg-background/80 backdrop-blur-sm">
          <DebuggerControls
            isPlaying={!isPaused}
            onPlayPause={() => setIsPaused(!isPaused)}
            onStep={refresh}
            onStop={() => setIsPaused(true)}
          />
          <div className="w-px h-6 bg-border mx-1" />
          <span className="text-xs text-muted-foreground px-2">
              {nodes.length} Nodes • {edges.length} Edges
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
