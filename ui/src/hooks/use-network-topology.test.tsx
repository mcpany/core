/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, waitFor } from '@testing-library/react';
import { useNetworkTopology } from './use-network-topology';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import dagre from 'dagre';
import { Graph } from '../types/topology';
import React from 'react';
import { apiClient } from '../lib/client';

// Mock dependencies
vi.mock('../lib/client', () => ({
  apiClient: {
    getTopology: vi.fn(),
  },
}));

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

  it('should return initial loading state', () => {
    (apiClient.getTopology as any).mockReturnValue(new Promise(() => {})); // Pending promise
    const { result } = renderHook(() => useNetworkTopology());
    expect(result.current.loading).toBe(true);
    expect(result.current.error).toBe(null);
  });

  it('should process topology graph correctly on success', async () => {
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

    (apiClient.getTopology as any).mockResolvedValue(mockGraph);

    const { result } = renderHook(() => useNetworkTopology());

    await waitFor(() => {
        expect(result.current.loading).toBe(false);
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

  it('should handle API errors', async () => {
    const error = new Error('Network error');
    (apiClient.getTopology as any).mockRejectedValue(error);

    const { result } = renderHook(() => useNetworkTopology());

    await waitFor(() => {
        expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe(error);
    expect(result.current.nodes).toEqual([]);
  });

  it('should not re-layout if structure is the same (caching)', async () => {
    const mockGraph1: Graph = {
      core: { id: 'core-1', label: 'Core', type: 'NODE_TYPE_CORE', status: 'NODE_STATUS_ACTIVE' },
    };

    (apiClient.getTopology as any).mockResolvedValue(mockGraph1);

    const { result, rerender } = renderHook(() => useNetworkTopology());

    await waitFor(() => {
        expect(result.current.loading).toBe(false);
    });

    expect(dagre.layout).toHaveBeenCalledTimes(1);

    // Second render with SAME structure
    const mockGraph2: Graph = {
      core: { id: 'core-1', label: 'Core Updated', type: 'NODE_TYPE_CORE', status: 'NODE_STATUS_ACTIVE' }, // Label changed, structure same
    };

    (apiClient.getTopology as any).mockResolvedValue(mockGraph2);
    result.current.refreshTopology();

    await waitFor(() => {
         // We need to wait for the effect of refreshTopology
         // Since it's async, we can check if nodes updated
         // But here we rely on mockResolvedValue being picked up
    });

    // We need to manually trigger the update logic or simulate the poll/refresh
    // Re-rendering hooks doesn't re-run effects unless deps change.
    // Calling refreshTopology triggers fetchTopology.

    // Wait for the state update
    await waitFor(() => {
        const coreNode = result.current.nodes.find((n) => n.id === 'core-1');
        expect(coreNode?.data.label).toBe('Core Updated');
    });

    // Layout should NOT be called again (still 1) because structure hash matches
    expect(dagre.layout).toHaveBeenCalledTimes(1);
  });
});
