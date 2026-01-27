/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { renderHook, act, waitFor } from '@testing-library/react';
import { useNetworkTopology } from '../../src/hooks/use-network-topology';
import { ServiceHealthProvider } from '../../src/contexts/service-health-context';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import React from 'react';

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
            json: async () => mockGraph,
            text: async () => JSON.stringify(mockGraph)
        });
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    // Helper wrapper
    const wrapper = ({ children }: { children: React.ReactNode }) => (
        <ServiceHealthProvider>{children}</ServiceHealthProvider>
    );

    it('should initialize with default nodes and edges', async () => {
        const { result } = renderHook(() => useNetworkTopology(), { wrapper });

        // Wait for fetch to complete and state to update
        await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        }, { timeout: 2000 });

        expect(result.current.edges.length).toBe(0); // Only core node, no clients -> no edges?

        const coreNode = result.current.nodes.find(n => n.id === 'mcp-core');
        expect(coreNode).toBeDefined();
        expect(coreNode?.data.label).toBe('MCP Any Core');
    });

    it('should update node positions on refresh', async () => {
        const { result } = renderHook(() => useNetworkTopology(), { wrapper });

        await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        });

        // Mock a change in graph to trigger layout change or just check refresh calls fetch
        const newMockGraph = {
             ...mockGraph,
             clients: [{ id: 'client-1', type: 'NODE_TYPE_CLIENT', status: 'NODE_STATUS_ACTIVE', label: 'Client 1' }]
        };

        global.fetch = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => newMockGraph,
            text: async () => JSON.stringify(newMockGraph)
        });

        await act(async () => {
            result.current.refreshTopology();
        });

        await waitFor(() => {
             expect(result.current.nodes.length).toBeGreaterThan(1);
        }, { timeout: 2000 });

        expect(result.current.edges.length).toBeGreaterThan(0);
    });

    it('should reset node positions on auto-layout', async () => {
         const { result } = renderHook(() => useNetworkTopology(), { wrapper });

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
        const { result } = renderHook(() => useNetworkTopology(), { wrapper });

        await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        });

        const initialNodes = result.current.nodes;

        // refreshTopology calls fetchData (in context).
        // Mock returns same mockGraph object.
        await act(async () => {
            result.current.refreshTopology();
        });

        // Use waitFor because state update might happen but result in same object reference if optimized
        await waitFor(() => {
             // Just ensuring no crash and potentially checking reference equality if the optimization works as intended
             // However, ServiceHealthProvider updates latestTopology state.
             // If latestTopology changes (even if deep equal), useNetworkTopology effect runs.
             // Inside useNetworkTopology, we check structure hash.
             // So setNodes should NOT be called if structure hash is same.
        });

        expect(result.current.nodes).toBe(initialNodes);
    });
});
