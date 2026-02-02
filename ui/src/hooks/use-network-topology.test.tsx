/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { renderHook, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useNetworkTopology } from './use-network-topology';
import { useTopology } from '../contexts/service-health-context';
import dagre from 'dagre';
import { Graph } from '../types/topology';

// Mock dependencies
vi.mock('../contexts/service-health-context', () => ({
  useTopology: vi.fn(),
}));

// Mock @xyflow/react hooks to use standard React state
vi.mock('@xyflow/react', () => ({
  useNodesState: (initial: any) => React.useState(initial),
  useEdgesState: (initial: any) => React.useState(initial),
  addEdge: vi.fn(),
  MarkerType: { ArrowClosed: 'arrowclosed' },
  Position: { Top: 'top', Bottom: 'bottom', Left: 'left', Right: 'right' },
}));

// Mock dagre
vi.mock('dagre', () => {
  const Graph = vi.fn();
  Graph.prototype.setGraph = vi.fn();
  Graph.prototype.setDefaultEdgeLabel = vi.fn();
  Graph.prototype.setNode = vi.fn();
  Graph.prototype.setEdge = vi.fn();
  Graph.prototype.node = vi.fn().mockReturnValue({ x: 100, y: 100 });

  return {
    default: {
      graphlib: { Graph },
      layout: vi.fn(),
    },
  };
});

describe('useNetworkTopology Optimization', () => {
  const mockRefreshTopology = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should avoid re-layout when topology structure is identical', async () => {
    const initialGraph: Graph = {
      core: {
        id: 'core-1',
        label: 'Core',
        type: 'NODE_TYPE_CORE',
        status: 'NODE_STATUS_ACTIVE',
        children: [
            { id: 'svc-1', label: 'Service 1', type: 'NODE_TYPE_SERVICE', status: 'NODE_STATUS_ACTIVE' }
        ]
      },
      clients: []
    };

    // First render
    let currentGraph = initialGraph;
    (useTopology as any).mockImplementation(() => ({
      latestTopology: currentGraph,
      refreshTopology: mockRefreshTopology,
    }));

    const { result, rerender } = renderHook(() => useNetworkTopology());

    // Should call layout initially
    await waitFor(() => {
        expect(result.current.nodes).toHaveLength(2); // Core + Service
    });
    expect(dagre.layout).toHaveBeenCalledTimes(1);

    // Second render: Same structure, different metrics/status
    currentGraph = {
      ...initialGraph,
      core: {
        ...initialGraph.core!,
        status: 'NODE_STATUS_ERROR' // Status changed
      }
    };

    // Force rerender by updating the mock and calling rerender
    (useTopology as any).mockImplementation(() => ({
      latestTopology: currentGraph,
      refreshTopology: mockRefreshTopology,
    }));

    rerender();

    await waitFor(() => {
         // Verify status updated in nodes
         const coreNode = result.current.nodes.find((n: any) => n.id === 'core-1');
         expect(coreNode?.data.status).toBe('NODE_STATUS_ERROR');
    });

    // CRITICAL CHECK: dagre.layout should NOT have been called again
    expect(dagre.layout).toHaveBeenCalledTimes(1);
  });

  it('should re-layout when topology structure changes', async () => {
      const initialGraph: Graph = {
        core: {
          id: 'core-1',
          label: 'Core',
          type: 'NODE_TYPE_CORE',
          status: 'NODE_STATUS_ACTIVE',
          children: []
        },
        clients: []
      };

      let currentGraph = initialGraph;
      (useTopology as any).mockImplementation(() => ({
        latestTopology: currentGraph,
        refreshTopology: mockRefreshTopology,
      }));

      const { result, rerender } = renderHook(() => useNetworkTopology());

      await waitFor(() => {
          expect(result.current.nodes).toHaveLength(1);
      });
      expect(dagre.layout).toHaveBeenCalledTimes(1);

      // Add a node
      currentGraph = {
        ...initialGraph,
        core: {
            ...initialGraph.core!,
            children: [
                { id: 'svc-new', label: 'New Service', type: 'NODE_TYPE_SERVICE', status: 'NODE_STATUS_ACTIVE' }
            ]
        }
      };

      (useTopology as any).mockImplementation(() => ({
        latestTopology: currentGraph,
        refreshTopology: mockRefreshTopology,
      }));

      rerender();

      await waitFor(() => {
          expect(result.current.nodes).toHaveLength(2);
      });

      // Should have called layout again
      expect(dagre.layout).toHaveBeenCalledTimes(2);
  });
});
