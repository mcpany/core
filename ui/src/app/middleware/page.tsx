/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client"

import { useCallback, useState } from 'react';
import { ReactFlow, Background, Controls, MiniMap, useNodesState, useEdgesState, addEdge, Connection, Edge } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Plus } from "lucide-react"

const initialNodes = [
  { id: '1', position: { x: 0, y: 100 }, data: { label: 'Input Request' }, type: 'input' },
  { id: '2', position: { x: 200, y: 100 }, data: { label: 'Auth Middleware' } },
  { id: '3', position: { x: 400, y: 100 }, data: { label: 'Rate Limiter' } },
  { id: '4', position: { x: 600, y: 100 }, data: { label: 'Router' }, type: 'output' },
];
const initialEdges = [
    { id: 'e1-2', source: '1', target: '2' },
    { id: 'e2-3', source: '2', target: '3' },
    { id: 'e3-4', source: '3', target: '4' },
];

export default function MiddlewarePage() {
    const [nodes, , onNodesChange] = useNodesState(initialNodes);
    const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

    const onConnect = useCallback(
        (params: Connection) => setEdges((eds) => addEdge(params, eds)),
        [setEdges],
    );

  return (
    <div className="space-y-4 h-full flex flex-col">
       <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Middleware Pipeline</h2>
          <p className="text-muted-foreground">Visual configuration of the request processing pipeline.</p>
        </div>
        <Button>
            <Plus className="mr-2 h-4 w-4" /> Add Middleware
        </Button>
      </div>

      <Card className="flex-1 min-h-[500px] border-2 border-dashed">
        <CardContent className="p-0 h-full">
            <ReactFlow
                nodes={nodes}
                edges={edges}
                onNodesChange={onNodesChange}
                onEdgesChange={onEdgesChange}
                onConnect={onConnect}
            >
                <Controls />
                <MiniMap />
                <Background gap={12} size={1} />
            </ReactFlow>
        </CardContent>
      </Card>
    </div>
  );
}
