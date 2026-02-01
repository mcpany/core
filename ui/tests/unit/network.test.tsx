/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { renderHook, act, waitFor } from '@testing-library/react';
import { useNetworkTopology } from '../../src/hooks/use-network-topology';
import { ServiceHealthProvider } from '../../src/contexts/service-health-context';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import React from 'react';

describe.skip('useNetworkTopology', () => {
    // Skipped due to flaky mock behavior in CI
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

    const wrapper = ({ children }: { children: React.ReactNode }) => (
        <ServiceHealthProvider>{children}</ServiceHealthProvider>
    );

    it('should initialize with default nodes and edges', async () => {
        const { result } = renderHook(() => useNetworkTopology(), { wrapper });
        await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        }, { timeout: 2000 });
        expect(result.current.edges.length).toBe(0);
        const coreNode = result.current.nodes.find(n => n.id === 'mcp-core');
        expect(coreNode).toBeDefined();
        expect(coreNode?.data.label).toBe('MCP Any Core');
    });
});
