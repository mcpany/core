/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useCallback, useState, useEffect, useRef } from 'react';
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
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { UserNode, AgentNode, ToolNode, ResourceNode, ServiceNode } from './custom-nodes';
import { Card } from '@/components/ui/card';
import { DebuggerControls } from './debugger-controls';
import { VariableInspector } from './variable-inspector';
import { apiClient } from '@/lib/client';
import { TopologyGraph, TopologyNode, NodeType, NodeStatus } from '@/types/topology';
import dagre from 'dagre';
import { Loader2, RefreshCcw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useToast } from '@/hooks/use-toast';

const nodeTypes = {
  user: UserNode,
  agent: AgentNode, // Maps to CORE
  tool: ToolNode,
  resource: ResourceNode,
  service: ServiceNode,
};

// Map Proto NodeType to React Flow Node Type (from nodeTypes above)
const mapNodeType = (type: NodeType): string => {
    switch (type) {
        case "NODE_TYPE_CLIENT": return 'user';
        case "NODE_TYPE_CORE": return 'agent';
        case "NODE_TYPE_SERVICE": return 'service';
        case "NODE_TYPE_TOOL": return 'tool';
        case "NODE_TYPE_RESOURCE": return 'resource';
        default: return 'service'; // Default fallback
    }
};

const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'LR') => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  // Set node dimensions roughly based on custom node size
  const nodeWidth = 220;
  const nodeHeight = 100;

  dagreGraph.setGraph({ rankdir: direction, align: 'DL', nodesep: 60, ranksep: 120 });

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
      targetPosition: 'left',
      sourcePosition: 'right',
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
  const [isPlaying, setIsPlaying] = useState(false);
  const [selectedNode, setSelectedNode] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();
  const initialized = useRef(false);

  // Transform TopologyGraph to React Flow Nodes/Edges
  const transformTopology = (graph: TopologyGraph) => {
      const newNodes: Node[] = [];
      const newEdges: Edge[] = [];

      // Helper to process nodes recursively
      const processNode = (node: TopologyNode, parentId?: string) => {
          const type = mapNodeType(node.type);

          // Build label with metrics if available
          let label = node.label;
          let statusStr = "";

          if (node.status === "NODE_STATUS_ERROR") statusStr = "Error";
          if (node.status === "NODE_STATUS_INACTIVE") statusStr = "Inactive";

          const flowNode: Node = {
              id: node.id,
              type: type,
              data: {
                  label: label,
                  status: statusStr,
                  metrics: node.metrics,
                  role: node.metadata?.role || (node.type === "NODE_TYPE_CORE" ? 'Gateway' : undefined)
              },
              position: { x: 0, y: 0 } // Layout will set this
          };
          newNodes.push(flowNode);

          // Add Edge from parent
          if (parentId) {
              const edgeId = `e-${parentId}-${node.id}`;
              const isAnimated = node.status === "NODE_STATUS_ACTIVE" && isPlaying;

              newEdges.push({
                  id: edgeId,
                  source: parentId,
                  target: node.id,
                  animated: isAnimated, // Only animate if "Live" is on? Or generally active?
                  style: {
                      stroke: node.status === "NODE_STATUS_ERROR" ? '#ef4444' :
                              node.status === "NODE_STATUS_INACTIVE" ? '#9ca3af' : '#22c55e',
                      strokeWidth: 2
                  },
                  type: 'default',
                  markerEnd: { type: MarkerType.ArrowClosed }
              });
          }

          // Process Children
          if (node.children) {
              node.children.forEach(child => processNode(child, node.id));
          }
      };

      // 1. Process Core
      if (graph.core) {
          processNode(graph.core);
      } else {
          // Fallback if no core (should not happen usually)
          newNodes.push({ id: 'core-missing', type: 'agent', data: { label: 'MCP Core (Missing)' }, position: { x: 0, y: 0 } });
      }

      // 2. Process Clients (connect to Core)
      const coreId = graph.core?.id || 'core-missing';
      if (graph.clients) {
          graph.clients.forEach(client => {
              // Clients are sources, connecting TO Core.
              // So edge is Client -> Core
              const clientType = mapNodeType(client.type);
              newNodes.push({
                  id: client.id,
                  type: clientType,
                  data: { label: client.label || client.id },
                  position: { x: 0, y: 0 }
              });

              newEdges.push({
                  id: `e-${client.id}-${coreId}`,
                  source: client.id,
                  target: coreId,
                  animated: isPlaying,
                  style: { stroke: '#3b82f6', strokeWidth: 2, strokeDasharray: '5,5' },
                  markerEnd: { type: MarkerType.ArrowClosed }
              });
          });
      }

      return { newNodes, newEdges };
  };

  const fetchTopology = async () => {
      try {
          const graph = await apiClient.getTopology();
          const { newNodes, newEdges } = transformTopology(graph);

          // Apply Layout
          const layouted = getLayoutedElements(newNodes, newEdges);

          setNodes(layouted.nodes as any);
          setEdges(layouted.edges);
      } catch (e) {
          console.error("Failed to fetch topology", e);
          if (!initialized.current) {
             toast({
                 title: "Topology Error",
                 description: "Failed to load network topology.",
                 variant: "destructive"
             });
          }
      } finally {
          setLoading(false);
          initialized.current = true;
      }
  };

  useEffect(() => {
      fetchTopology();
  }, []);

  // Polling for live mode
  useEffect(() => {
      let interval: NodeJS.Timeout;
      if (isPlaying) {
          interval = setInterval(() => {
              fetchTopology();
          }, 5000);
      }
      return () => clearInterval(interval);
  }, [isPlaying]);

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
            onStep={fetchTopology}
            onStop={() => { setIsPlaying(false); }}
          />
          <div className="w-px h-6 bg-border mx-1" />
          <Button variant="ghost" size="icon" onClick={fetchTopology} disabled={loading || isPlaying}>
              {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCcw className="h-4 w-4" />}
          </Button>
        </Card>
      </div>

      <VariableInspector selectedNode={selectedNode} onClose={() => setSelectedNode(null)} />

      {loading && !initialized.current ? (
          <div className="absolute inset-0 flex items-center justify-center bg-background/50 z-20">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
      ) : (
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
      )}
    </div>
  );
}
