/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useNetworkTopology } from '@/hooks/use-network-topology';
import { Graph } from '@/types/topology';

// Mock dagre and React Flow hooks
vi.mock('dagre', () => ({
  default: {
    graphlib: {
      Graph: class {
        setGraph() {}
        setDefaultEdgeLabel() {}
        setNode() {}
        setEdge() {}
        node(id: string) { return { x: 100, y: 100 }; }
      },
    },
    layout: vi.fn(),
  },
}));

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('useNetworkTopology', () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it('fetches and transforms topology data correctly', async () => {
    const mockGraph: Graph = {
      clients: [
        { id: 'client-1', label: 'Client 1', type: 'NODE_TYPE_CLIENT', status: 'NODE_STATUS_ACTIVE' }
      ],
      core: {
        id: 'core-1',
        label: 'Core',
        type: 'NODE_TYPE_CORE',
        status: 'NODE_STATUS_ACTIVE',
        children: [
          { id: 'svc-1', label: 'Service 1', type: 'NODE_TYPE_SERVICE', status: 'NODE_STATUS_ACTIVE' }
        ]
      }
    };

    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => mockGraph,
    });

    const { result } = renderHook(() => useNetworkTopology());

    await waitFor(() => {
        expect(result.current.nodes.length).toBeGreaterThan(0);
    });

    // Verify Nodes
    const nodeIds = result.current.nodes.map(n => n.id);
    expect(nodeIds).toContain('client-1');
    expect(nodeIds).toContain('core-1');
    expect(nodeIds).toContain('svc-1');

    // Verify Edges
    const edgeIds = result.current.edges.map(e => e.id);
    // Client -> Core
    expect(edgeIds).toContain('e-client-1-core-1');
    // Core -> Service
    expect(edgeIds).toContain('e-core-1-svc-1');
  });

  it('handles fetch errors gracefully', async () => {
    mockFetch.mockRejectedValue(new Error('Network error'));
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    const { result } = renderHook(() => useNetworkTopology());

    // Should verify it handles error (not crashing), nodes empty initially
    expect(result.current.nodes).toEqual([]);

    consoleSpy.mockRestore();
  });
});
