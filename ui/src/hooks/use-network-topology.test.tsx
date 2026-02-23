import { renderHook } from '@testing-library/react';
import { useNetworkTopology } from './use-network-topology';
import { useTopology } from '../contexts/service-health-context';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import dagre from 'dagre';
import { Graph } from '../types/topology';
import React from 'react';

// Mock dependencies
vi.mock('../contexts/service-health-context', () => ({
  useTopology: vi.fn(),
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
  const mockRefreshTopology = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    (useTopology as any).mockReturnValue({
      latestTopology: null,
      refreshTopology: mockRefreshTopology,
    });
  });

  it('should return initial empty state when topology is null', () => {
    const { result } = renderHook(() => useNetworkTopology());
    expect(result.current.nodes).toEqual([]);
    expect(result.current.edges).toEqual([]);
  });

  it('should process topology graph correctly', () => {
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

    (useTopology as any).mockReturnValue({
      latestTopology: mockGraph,
      refreshTopology: mockRefreshTopology,
    });

    const { result } = renderHook(() => useNetworkTopology());

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

  it('should not re-layout if structure is the same (caching)', () => {
    const mockGraph1: Graph = {
      core: { id: 'core-1', label: 'Core', type: 'NODE_TYPE_CORE', status: 'NODE_STATUS_ACTIVE' },
    };

    // First render
    (useTopology as any).mockReturnValue({
      latestTopology: mockGraph1,
      refreshTopology: mockRefreshTopology,
    });

    const { result, rerender } = renderHook(() => useNetworkTopology());

    // Check dagre layout was called
    expect(dagre.layout).toHaveBeenCalledTimes(1);

    // Second render with SAME structure
    const mockGraph2: Graph = {
      core: { id: 'core-1', label: 'Core Updated', type: 'NODE_TYPE_CORE', status: 'NODE_STATUS_ACTIVE' }, // Label changed, structure same
    };

    (useTopology as any).mockReturnValue({
      latestTopology: mockGraph2,
      refreshTopology: mockRefreshTopology,
    });

    rerender();

    // Layout should NOT be called again
    expect(dagre.layout).toHaveBeenCalledTimes(1);

    // But data should update
    const coreNode = result.current.nodes.find((n) => n.id === 'core-1');
    expect(coreNode?.data.label).toBe('Core Updated');
  });

  it('should re-layout if structure changes', () => {
     const mockGraph1: Graph = {
      core: { id: 'core-1', label: 'Core', type: 'NODE_TYPE_CORE', status: 'NODE_STATUS_ACTIVE' },
    };

    // First render
    (useTopology as any).mockReturnValue({
      latestTopology: mockGraph1,
      refreshTopology: mockRefreshTopology,
    });

    const { result, rerender } = renderHook(() => useNetworkTopology());
    expect(dagre.layout).toHaveBeenCalledTimes(1);

    // Second render with DIFFERENT structure
    const mockGraph2: Graph = {
      core: {
          id: 'core-1',
          label: 'Core',
          type: 'NODE_TYPE_CORE',
          status: 'NODE_STATUS_ACTIVE',
          children: [{ id: 'child-1', label: 'Child', type: 'NODE_TYPE_SERVICE', status: 'NODE_STATUS_ACTIVE' }]
      },
    };

    (useTopology as any).mockReturnValue({
      latestTopology: mockGraph2,
      refreshTopology: mockRefreshTopology,
    });

    rerender();

    // Layout SHOULD be called again
    expect(dagre.layout).toHaveBeenCalledTimes(2);
    expect(result.current.nodes).toHaveLength(2);
  });
});
