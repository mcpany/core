/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect } from 'vitest';
import { traceToGraph } from '@/lib/flow-layout';
import { Trace } from '@/types/trace';

describe('traceToGraph', () => {
  it('should transform a simple trace into nodes and edges', () => {
    const trace: Trace = {
      id: 'trace-1',
      timestamp: new Date().toISOString(),
      totalDuration: 100,
      status: 'success',
      trigger: 'user',
      rootSpan: {
        id: 'span-1',
        name: 'root',
        type: 'core',
        startTime: 0,
        endTime: 100,
        status: 'success',
        children: [
          {
            id: 'span-2',
            name: 'tool-a',
            type: 'tool',
            startTime: 10,
            endTime: 50,
            status: 'success',
          },
          {
            id: 'span-3',
            name: 'service-b',
            type: 'service',
            startTime: 60,
            endTime: 90,
            status: 'success',
          },
        ],
      },
    };

    const result = traceToGraph(trace);

    // Expected Nodes: User, Core, Tool A, Service B
    expect(result.nodes).toHaveLength(4);
    expect(result.nodes.map(n => n.id)).toContain('user');
    expect(result.nodes.map(n => n.id)).toContain('core');
    expect(result.nodes.map(n => n.id)).toContain('tool-tool-a'); // Sanitized name
    expect(result.nodes.map(n => n.id)).toContain('service-service-b'); // Sanitized name

    // Expected Edges: User->Core, Core->Tool A, Core->Service B
    expect(result.edges).toHaveLength(3);
    expect(result.edges.find(e => e.source === 'user' && e.target === 'core')).toBeDefined();
    expect(result.edges.find(e => e.source === 'core' && e.target === 'tool-tool-a')).toBeDefined();
    expect(result.edges.find(e => e.source === 'core' && e.target === 'service-service-b')).toBeDefined();
  });

  it('should handle nested calls', () => {
    const trace: Trace = {
      id: 'trace-2',
      timestamp: new Date().toISOString(),
      totalDuration: 100,
      status: 'success',
      trigger: 'user',
      rootSpan: {
        id: 'root',
        name: 'root',
        type: 'core',
        startTime: 0,
        endTime: 100,
        status: 'success',
        children: [
          {
            id: 'agent',
            name: 'orchestrator',
            type: 'service',
            startTime: 10,
            endTime: 90,
            status: 'success',
            children: [
              {
                id: 'tool',
                name: 'search',
                type: 'tool',
                startTime: 20,
                endTime: 80,
                status: 'success',
              },
            ],
          },
        ],
      },
    };

    const result = traceToGraph(trace);

    // Nodes: User, Core, Orchestrator, Search
    expect(result.nodes).toHaveLength(4);

    // Edges: User->Core, Core->Orchestrator, Orchestrator->Search
    expect(result.edges).toHaveLength(3);
    expect(result.edges.find(e => e.source === 'core' && e.target === 'service-orchestrator')).toBeDefined();
    expect(result.edges.find(e => e.source === 'service-orchestrator' && e.target === 'tool-search')).toBeDefined();
  });
});
