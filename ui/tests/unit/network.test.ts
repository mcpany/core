/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { renderHook, act } from '@testing-library/react';
import { useNetworkTopology } from '../../src/hooks/use-network-topology';
import { describe, it, expect } from 'vitest';

describe('useNetworkTopology', () => {
    it('should initialize with default nodes and edges', () => {
        const { result } = renderHook(() => useNetworkTopology());

        expect(result.current.nodes.length).toBeGreaterThan(0);
        expect(result.current.edges.length).toBeGreaterThan(0);

        const coreNode = result.current.nodes.find(n => n.id === 'mcp-core');
        expect(coreNode).toBeDefined();
        expect(coreNode?.data.label).toBe('MCP Any Core');
    });

    it('should update node positions on refresh', () => {
        const { result } = renderHook(() => useNetworkTopology());
        const initialPosition = result.current.nodes[0].position;

        act(() => {
            result.current.refreshTopology();
        });

        const newPosition = result.current.nodes[0].position;
        expect(newPosition).not.toEqual(initialPosition);
    });

    it('should reset node positions on auto-layout', () => {
         const { result } = renderHook(() => useNetworkTopology());

         // Move them first
         act(() => {
            result.current.refreshTopology();
        });

        // Reset
        act(() => {
            result.current.autoLayout();
        });

        const coreNode = result.current.nodes.find(n => n.id === 'mcp-core');
        // Initial mock position for core is 400, 300
        expect(coreNode?.position).toEqual({ x: 400, y: 300 });
    });
});
