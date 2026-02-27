/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, waitFor } from '@testing-library/react';
import { useNetworkTopology } from './use-network-topology';
import { ServiceHealthProvider } from '../contexts/service-health-context';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import dagre from 'dagre';
import { Graph } from '../types/topology';
import React from 'react';

// Mock fetch globally
global.fetch = vi.fn();

vi.mock('@xyflow/react', async () => {
  const React = await import('react');
  return {
    useNodesState: (initial: any) => React.useState(initial),
    useEdgesState: (initial: any) => React.useState(initial),
    addEdge: vi.fn(),
    MarkerType: { ArrowClosed: 'arrowclosed' },
    Position: { Top: 'top', Bottom: 'bottom', Left: 'left', Right: 'right' },
  };
});

// Mock dagre with a class for Graph
vi.mock('dagre', () => {
  const layout = vi.fn();

  class MockGraph {
    setDefaultEdgeLabel = vi.fn();
    setGraph = vi.fn();
    setNode = vi.fn();
    setEdge = vi.fn();
    node = vi.fn(() => ({ x: 100, y: 100 }));
  }

  return {
    default: {
      graphlib: {
        Graph: MockGraph,
      },
      layout,
    },
  };
});

describe('useNetworkTopology', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <ServiceHealthProvider>{children}</ServiceHealthProvider>
  );

  it('should return initial empty state when topology is null', () => {
    (global.fetch as any).mockResolvedValue({
      ok: true,
      text: async () => JSON.stringify({}), // Empty or null topology
      json: async () => ({}),
      headers: { get: () => null }
    });

    const { result } = renderHook(() => useNetworkTopology(), { wrapper });
    expect(result.current.nodes).toEqual([]);
    expect(result.current.edges).toEqual([]);
  });

  it('should process topology graph correctly', async () => {
    const mockGraph: Graph = {
      core: {
        id: 'core-1',
        label: 'Core Server',
        type: 'NODE_TYPE_CORE',
        status: 'NODE_STATUS_ACTIVE',
        children: [
          {
            id: 'service-1',
            label: 'Service A',
            type: 'NODE_TYPE_SERVICE',
            status: 'NODE_STATUS_ACTIVE',
          },
        ],
      },
      clients: [
        {
          id: 'client-1',
          label: 'Client A',
          type: 'NODE_TYPE_CLIENT',
          status: 'NODE_STATUS_ACTIVE',
        },
      ],
    };

    (global.fetch as any).mockResolvedValue({
      ok: true,
      text: async () => JSON.stringify(mockGraph),
      json: async () => mockGraph,
      headers: { get: () => null }
    });

    const { result } = renderHook(() => useNetworkTopology(), { wrapper });

    await waitFor(() => {
        expect(result.current.nodes.length).toBeGreaterThan(0);
    });

    // Check nodes (Core, Service, Client) -> 3 nodes
    expect(result.current.nodes).toHaveLength(3);

    const coreNode = result.current.nodes.find((n) => n.id === 'core-1');
    expect(coreNode).toBeDefined();
    expect(coreNode?.type).toBe('default');
    expect(coreNode?.data.label).toBe('Core Server');
    // Check styling from getNodeClassName
    expect(coreNode?.className).toContain('bg-white border-black text-black');

    const clientNode = result.current.nodes.find((n) => n.id === 'client-1');
    expect(clientNode).toBeDefined();
    expect(clientNode?.className).toContain('bg-green-50');

    // Check edges
    expect(result.current.edges).toHaveLength(2);

    const clientToCore = result.current.edges.find((e) => e.source === 'client-1' && e.target === 'core-1');
    expect(clientToCore).toBeDefined();

    const coreToService = result.current.edges.find((e) => e.source === 'core-1' && e.target === 'service-1');
    expect(coreToService).toBeDefined();
  });

  it('should not re-layout if structure is the same (caching)', async () => {
    const mockGraph1: Graph = {
      core: { id: 'core-1', label: 'Core', type: 'NODE_TYPE_CORE', status: 'NODE_STATUS_ACTIVE' },
    };

    // First fetch
    (global.fetch as any).mockResolvedValueOnce({
      ok: true,
      text: async () => JSON.stringify(mockGraph1),
      json: async () => mockGraph1,
      headers: { get: () => null }
    });

    const { result, rerender } = renderHook(() => useNetworkTopology(), { wrapper });

    await waitFor(() => {
        expect(result.current.nodes.length).toBeGreaterThan(0);
    });

    // Check dagre layout was called
    expect(dagre.layout).toHaveBeenCalledTimes(1);

    // Second render with SAME structure but label change
    const mockGraph2: Graph = {
      core: { id: 'core-1', label: 'Core Updated', type: 'NODE_TYPE_CORE', status: 'NODE_STATUS_ACTIVE' },
    };

    // Simulate polling update
    (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        text: async () => JSON.stringify(mockGraph2),
        json: async () => mockGraph2,
        headers: { get: () => null }
    });

    // To test caching, we simulate a re-render with the same structure
    rerender();

    // dagre.layout should still only be called once because the structure didn't change
    expect(dagre.layout).toHaveBeenCalledTimes(1);
  });
});
