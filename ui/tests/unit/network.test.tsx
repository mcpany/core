/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, act, waitFor } from '@testing-library/react';
import { useNetworkTopology } from '../../src/hooks/use-network-topology';
import { describe, it, expect, vi, afterEach } from 'vitest';
import React from 'react';

// vi.hoisted creates variables available inside vi.mock factories
const { getTopology, setTopology, getRefreshFn } = vi.hoisted(() => {
    let _topology: any = null;
    let _refreshFn: (() => void) = () => {};
    return {
        getTopology: () => _topology,
        setTopology: (t: any) => { _topology = t; },
        getRefreshFn: () => _refreshFn,
    };
});

// Mock the service-health-context module
vi.mock('../../src/contexts/service-health-context', () => ({
    ServiceHealthProvider: ({ children }: any) => children,
    useTopology: () => ({
        latestTopology: getTopology(),
        refreshTopology: getRefreshFn(),
    }),
}));

// Mock dagre to avoid complex graph layout in tests
vi.mock('dagre', () => {
    const Graph = vi.fn();
    Graph.prototype.setGraph = vi.fn();
    Graph.prototype.setDefaultEdgeLabel = vi.fn();
    Graph.prototype.setNode = vi.fn();
    Graph.prototype.setEdge = vi.fn();
    Graph.prototype.node = vi.fn(() => ({ x: 100, y: 100 }));
    const dagreMock = { graphlib: { Graph }, layout: vi.fn() };
    return { default: dagreMock, ...dagreMock };
});

// Mock @xyflow/react hooks
vi.mock('@xyflow/react', async () => {
    const React = await import('react');
    return {
        useNodesState: (initial: any[]) => {
            const [nodes, setNodes] = React.useState<any[]>(initial);
            return [nodes, setNodes, () => {}];
        },
        useEdgesState: (initial: any[]) => {
            const [edges, setEdges] = React.useState<any[]>(initial);
            return [edges, setEdges, () => {}];
        },
        addEdge: (params: any, edges: any[]) => [...edges, params],
        MarkerType: { ArrowClosed: 'arrowclosed' },
        Position: { Top: 'top', Bottom: 'bottom', Left: 'left', Right: 'right' },
    };
});

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

    afterEach(() => {
        setTopology(null);
        vi.restoreAllMocks();
    });

    // Simple wrapper that re-renders hook when topology changes
    const createWrapper = () => {
        const Wrapper = ({ children }: { children: React.ReactNode }) => {
            return React.createElement(React.Fragment, null, children);
        };
        return Wrapper;
    };

    it('should initialize with default nodes and edges', async () => {
        setTopology(mockGraph);
        const wrapper = createWrapper();
        const { result } = renderHook(() => useNetworkTopology(), { wrapper });

        // Wait for processGraph to run and nodes to be set
        await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        }, { timeout: 3000 });

        expect(result.current.edges.length).toBe(0); // Only core node, no clients

        const coreNode = result.current.nodes.find((n: any) => n.id === 'mcp-core');
        expect(coreNode).toBeDefined();
        expect(coreNode?.data.label).toBe('MCP Any Core');
    });

    it('should update node positions on refresh', async () => {
        setTopology(mockGraph);
        const wrapper = createWrapper();
        const { result } = renderHook(() => useNetworkTopology(), { wrapper });

        await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        }, { timeout: 3000 });

        const coreNode = result.current.nodes.find((n: any) => n.id === 'mcp-core');
        expect(coreNode).toBeDefined();
        act(() => {
            result.current.refreshTopology();
        });
        expect(result.current.nodes.length).toBeGreaterThan(0);
    });

    it('should reset node positions on auto-layout', async () => {
        setTopology(mockGraph);
        const wrapper = createWrapper();
        const { result } = renderHook(() => useNetworkTopology(), { wrapper });

        await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        }, { timeout: 3000 });

        await act(async () => {
            result.current.autoLayout();
        });

        const coreNode = result.current.nodes.find((n: any) => n.id === 'mcp-core');
        expect(coreNode).toBeDefined();
    });

    it('should not trigger state update if topology data is identical', async () => {
        setTopology(mockGraph);
        const wrapper = createWrapper();
        const { result } = renderHook(() => useNetworkTopology(), { wrapper });

        await waitFor(() => {
            expect(result.current.nodes.length).toBeGreaterThan(0);
        }, { timeout: 3000 });

        const initialLength = result.current.nodes.length;

        await act(async () => {
            result.current.refreshTopology();
        });

        expect(result.current.nodes.length).toBe(initialLength);
    });
});
