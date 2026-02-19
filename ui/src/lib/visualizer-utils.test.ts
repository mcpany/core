/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { traceToGraph } from './visualizer-utils';
import { Trace, Span } from '@/types/trace';

describe('traceToGraph', () => {
  it('should convert a simple trace into nodes and edges', () => {
    const trace: Trace = {
      id: 'trace-1',
      rootSpan: {
        id: 'span-1',
        name: 'get_weather',
        type: 'tool',
        startTime: 1000,
        endTime: 2000,
        status: 'success',
        serviceName: 'weather-service',
        children: []
      },
      timestamp: new Date().toISOString(),
      totalDuration: 1000,
      status: 'success',
      trigger: 'user'
    };

    const { nodes, edges } = traceToGraph(trace);

    // Nodes should include User, Core, and Weather Service
    expect(nodes).toHaveLength(3); // User, Core, Weather Service
    expect(nodes.map(n => n.id)).toContain('user');
    expect(nodes.map(n => n.id)).toContain('core');
    expect(nodes.map(n => n.id)).toContain('svc:weather-service');

    // Edges: User->Core, Core->Weather Service
    expect(edges).toHaveLength(2);
    expect(edges[0].source).toBe('user');
    expect(edges[0].target).toBe('core');
    expect(edges[1].source).toBe('core');
    expect(edges[1].target).toBe('svc:weather-service');
  });

  it('should handle nested calls', () => {
    const trace: Trace = {
      id: 'trace-2',
      rootSpan: {
        id: 'span-1',
        name: 'complex_workflow',
        type: 'service',
        serviceName: 'workflow-engine',
        startTime: 1000,
        endTime: 5000,
        status: 'success',
        children: [
            {
                id: 'span-2',
                name: 'fetch_data',
                type: 'tool',
                serviceName: 'db-service',
                startTime: 1100,
                endTime: 2000,
                status: 'success',
                children: []
            },
            {
                id: 'span-3',
                name: 'process_data',
                type: 'tool',
                serviceName: 'compute-service',
                startTime: 2100,
                endTime: 4000,
                status: 'success',
                children: []
            }
        ]
      },
      timestamp: new Date().toISOString(),
      totalDuration: 4000,
      status: 'success',
      trigger: 'user'
    };

    const { nodes, edges } = traceToGraph(trace);

    // User, Core, Workflow Engine, DB Service, Compute Service
    expect(nodes).toHaveLength(5);
    expect(nodes.map(n => n.id)).toContain('svc:workflow-engine');
    expect(nodes.map(n => n.id)).toContain('svc:db-service');
    expect(nodes.map(n => n.id)).toContain('svc:compute-service');

    // Edges:
    // User -> Core
    // Core -> Workflow Engine
    // Workflow Engine -> DB Service
    // Workflow Engine -> Compute Service
    expect(edges).toHaveLength(4);

    const workflowEdge = edges.find(e => e.source === 'core' && e.target === 'svc:workflow-engine');
    expect(workflowEdge).toBeDefined();

    const dbEdge = edges.find(e => e.source === 'svc:workflow-engine' && e.target === 'svc:db-service');
    expect(dbEdge).toBeDefined();

    const computeEdge = edges.find(e => e.source === 'svc:workflow-engine' && e.target === 'svc:compute-service');
    expect(computeEdge).toBeDefined();
  });
});
