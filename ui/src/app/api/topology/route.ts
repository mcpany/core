/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { NextResponse } from 'next/server';

export async function GET() {
  // Mock Topology Data
  // Nodes: Client, Gateway, Services, Tools

  const nodes = [
    {
      id: 'client',
      type: 'client',
      position: { x: 250, y: 0 },
      data: {
        label: 'Client (User)',
        status: 'active',
        metrics: {
          requests: 120, // req/min or total
          avgReqSize: '1.2KB',
          avgRespSize: '5.4KB'
        }
      },
    },
    {
      id: 'gateway',
      type: 'gateway',
      position: { x: 250, y: 150 },
      data: {
        label: 'MCP Gateway',
        status: 'active',
        metrics: {
          allowed: 1180,
          blocked: 12, // Firewall view
          cacheHitRate: '45%',
          cacheSaved: '2.4MB'
        }
      },
    },
    {
      id: 'svc-weather',
      type: 'service',
      position: { x: 100, y: 300 },
      data: {
        label: 'Weather Service',
        status: 'active',
         metrics: {
          requests: 450,
          latency: '24ms'
        }
      },
    },
    {
      id: 'svc-payments',
      type: 'service',
      position: { x: 400, y: 300 },
      data: {
        label: 'Payment Service',
        status: 'warning',
        metrics: {
          requests: 120,
          latency: '150ms',
          errorRate: '2%'
        }
      },
    },
    {
      id: 'tool-forecast',
      type: 'tool',
      position: { x: 50, y: 450 },
      parentId: 'svc-weather',
      extent: 'parent',
      data: { label: 'get_forecast' },
    },
     {
      id: 'tool-alerts',
      type: 'tool',
      position: { x: 150, y: 450 },
      parentId: 'svc-weather',
      extent: 'parent',
      data: { label: 'get_alerts' },
    }
  ];

  const edges = [
    { id: 'e1', source: 'client', target: 'gateway', animated: true, label: '120 req/s' },
    { id: 'e2', source: 'gateway', target: 'svc-weather', animated: true, label: '45 req/s' },
    { id: 'e3', source: 'gateway', target: 'svc-payments', animated: true, stroke: '#ff0000', label: '12 req/s' },
  ];

  return NextResponse.json({ nodes, edges });
}
