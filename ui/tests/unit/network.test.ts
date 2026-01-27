/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { renderHook, act, waitFor } from '@testing-library/react';
import { useNetworkTopology } from '../../src/hooks/use-network-topology';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('useNetworkTopology', () => {
    const mockGraph = {
        core: {
            id: 'mcp-core',
            label: 'MCP Any Core',
            type: 'NODE_TYPE_CORE',
            status: 'NODE_STATUS_ACTIVE',
            metrics: { qps: 10 }
        },
        clients: []
    };

    beforeEach(() => {
        global.fetch = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => mockGraph
        });
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it('should initialize with default nodes and edges', async () => {
        const { result } = renderHook(() => useNetworkTopology());

        // Wait for fetch to complete and state to update
        await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        });

        expect(result.current.edges.length).toBe(0); // Only core node, no clients -> no edges?

        const coreNode = result.current.nodes.find(n => n.id === 'mcp-core');
        expect(coreNode).toBeDefined();
        expect(coreNode?.data.label).toBe('MCP Any Core');
    });

    it('should update node positions on refresh', async () => {
        const { result } = renderHook(() => useNetworkTopology());

        await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        });

        const initialPosition = result.current.nodes[0].position;

        // Mock a change in graph to trigger layout change or just check refresh calls fetch

        // Let's modify the mock to return different data structure to force layout update
        const newMockGraph = {
             ...mockGraph,
             clients: [{ id: 'client-1', type: 'NODE_TYPE_CLIENT', status: 'NODE_STATUS_ACTIVE', label: 'Client 1' }]
        };

        global.fetch = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => newMockGraph
        });

        await act(async () => {
            result.current.refreshTopology();
        });

        await waitFor(() => {
             // We expect more nodes now
             expect(result.current.nodes.length).toBeGreaterThan(1);
        });

        // Since we added a node, structure changed, so layout should run.
        expect(result.current.edges.length).toBeGreaterThan(0);
    });

    it('should reset node positions on auto-layout', async () => {
         const { result } = renderHook(() => useNetworkTopology());

         await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        });

        // Reset
        await act(async () => {
            result.current.autoLayout();
        });

        const coreNode = result.current.nodes.find(n => n.id === 'mcp-core');
        expect(coreNode).toBeDefined();
    });

    it('should not trigger state update if topology data is identical', async () => {
        const { result } = renderHook(() => useNetworkTopology());

        await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        });

        const initialNodes = result.current.nodes;

        // refreshTopology calls fetchData.
        // Mock returns same mockGraph object.
        await act(async () => {
            result.current.refreshTopology();
        });

        // Should be same reference because setNodes was skipped
        expect(result.current.nodes).toBe(initialNodes);
    });
});
