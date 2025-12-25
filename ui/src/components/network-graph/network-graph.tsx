/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useCallback, useEffect } from 'react';
import {
  ReactFlow,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  Edge,
  Background,
  Controls,
  MiniMap,
  ReactFlowProvider,
  Node,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Service } from '@/types/service';
import ServiceNode from './service-node';
import CentralNode from './central-node';
import { getLayoutedElements, transformServicesToGraph } from '@/lib/graph-utils';

const nodeTypes = {
  service: ServiceNode,
  central: CentralNode,
};

interface NetworkGraphProps {
    services: Service[];
}

const NetworkGraphInternal = ({ services }: NetworkGraphProps) => {
  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);

  useEffect(() => {
      const { nodes: initialNodes, edges: initialEdges } = transformServicesToGraph(services);
      const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
        initialNodes,
        initialEdges
      );

      setNodes(layoutedNodes);
      setEdges(layoutedEdges);
  }, [services, setNodes, setEdges]);

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges],
  );

  return (
    <div className="w-full h-full min-h-[500px] border rounded-xl overflow-hidden shadow-inner bg-slate-50 dark:bg-slate-900/50">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        nodeTypes={nodeTypes}
        fitView
        attributionPosition="bottom-right"
      >
        <Background color="#94a3b8" gap={16} size={1} />
        <Controls />
        <MiniMap zoomable pannable />
      </ReactFlow>
    </div>
  );
};

export const NetworkGraph = (props: NetworkGraphProps) => (
    <ReactFlowProvider>
        <NetworkGraphInternal {...props} />
    </ReactFlowProvider>
);
