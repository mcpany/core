/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


'use client';

import React, { useEffect, useState, useCallback } from 'react';
import ReactFlow, {
  Background,
  Controls,
  Edge,
  Node,
  useNodesState,
  useEdgesState,
  MarkerType,
  Position
} from 'reactflow';
import 'reactflow/dist/style.css';
import dagre from 'dagre';

import { ClientNode, GatewayNode, ServiceNode } from './custom-nodes';

const nodeTypes = {
  client: ClientNode,
  gateway: GatewayNode,
  service: ServiceNode,
  tool: ServiceNode
};

const dagreGraph = new dagre.graphlib.Graph();
dagreGraph.setDefaultEdgeLabel(() => ({}));

const nodeWidth = 280;
const nodeHeight = 160;

const getLayoutedElements = (nodes: any[], edges: any[], direction = 'TB') => {
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
    return {
      ...node,
      targetPosition: isHorizontal ? Position.Left : Position.Top,
      sourcePosition: isHorizontal ? Position.Right : Position.Bottom,
      position: {
        x: nodeWithPosition.x - nodeWidth / 2,
        y: nodeWithPosition.y - nodeHeight / 2,
      },
    };
  });

  return { nodes: layoutedNodes, edges };
};

interface NetworkGraphProps {
    onNodeClick?: (event: React.MouseEvent, node: Node) => void;
}

export function NetworkGraph({ onNodeClick }: NetworkGraphProps) {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const res = await fetch('/api/topology');
        const data = await res.json();

        const styledEdges = data.edges.map((edge: any) => ({
             ...edge,
             type: 'default', // 'smoothstep' set in defaultEdgeOptions usually acts as fallback or base
             animated: true,
             style: { stroke: edge.stroke || '#64748b', strokeWidth: 2 },
             markerEnd: {
                 type: MarkerType.ArrowClosed,
                 color: edge.stroke || '#64748b',
             },
        }));

        const layouted = getLayoutedElements(data.nodes, styledEdges, 'TB');
        setNodes(layouted.nodes);
        setEdges(layouted.edges);
      } catch (err) {
        console.error("Failed to fetch topology", err);
      }
    };
    fetchData();

    // Poll for updates every 5s
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, [setNodes, setEdges]); // dependencies need to be stable

  return (
    <div className="w-full h-full bg-slate-50 dark:bg-slate-950 rounded-xl border overflow-hidden">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes as any}
        onNodeClick={onNodeClick}
        fitView
        attributionPosition="bottom-right"
        defaultEdgeOptions={{ type: 'smoothstep', animated: true }}
      >
        <Background gap={20} color="#94a3b8" className="opacity-10" />
        <Controls className="bg-white dark:bg-slate-900 border-border" />
      </ReactFlow>
    </div>
  );
}
