/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { describe, it, expect } from 'vitest';

// Define minimal Trace interface for testing
interface Span {
  id: string;
  name: string;
  type: 'tool' | 'service' | 'resource' | 'prompt' | 'core';
  startTime: number;
  endTime: number;
  status: 'success' | 'error' | 'pending';
}

interface Trace {
  id: string;
  rootSpan: Span;
  timestamp: string;
  totalDuration: number;
  status: 'success' | 'error' | 'pending';
  trigger: 'user' | 'webhook' | 'scheduler' | 'system';
}

// Logic copied from InspectorPage for verification
function filterTraces(traces: Trace[], searchQuery: string, statusFilter: string, typeFilter: string) {
    return traces.filter((trace) => {
        // Filter by Status
        if (statusFilter !== "all" && trace.status !== statusFilter) return false;

        // Filter by Type (root span type)
        if (typeFilter !== "all" && trace.rootSpan.type !== typeFilter) return false;

        // Filter by Search (ID or Name)
        if (searchQuery) {
            const query = searchQuery.toLowerCase();
            return (
                trace.id.toLowerCase().includes(query) ||
                trace.rootSpan.name.toLowerCase().includes(query)
            );
        }

        return true;
    });
}

describe('Inspector Filter Logic', () => {
    const mockTraces: Trace[] = [
        {
            id: 'trace-1',
            status: 'success',
            timestamp: new Date().toISOString(),
            totalDuration: 100,
            trigger: 'user',
            rootSpan: { id: 'span-1', name: 'get_weather', type: 'tool', status: 'success', startTime: 0, endTime: 100 }
        },
        {
            id: 'trace-2',
            status: 'error',
            timestamp: new Date().toISOString(),
            totalDuration: 50,
            trigger: 'user',
            rootSpan: { id: 'span-2', name: 'fetch_data', type: 'tool', status: 'error', startTime: 0, endTime: 50 }
        },
        {
            id: 'trace-3',
            status: 'success',
            timestamp: new Date().toISOString(),
            totalDuration: 200,
            trigger: 'system',
            rootSpan: { id: 'span-3', name: 'health_check', type: 'service', status: 'success', startTime: 0, endTime: 200 }
        }
    ];

    it('should return all traces when no filters are applied', () => {
        const result = filterTraces(mockTraces, '', 'all', 'all');
        expect(result).toHaveLength(3);
    });

    it('should filter by status', () => {
        const result = filterTraces(mockTraces, '', 'error', 'all');
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('trace-2');
    });

    it('should filter by type', () => {
        const result = filterTraces(mockTraces, '', 'all', 'service');
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('trace-3');
    });

    it('should filter by search query (name)', () => {
        const result = filterTraces(mockTraces, 'weather', 'all', 'all');
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('trace-1');
    });

    it('should filter by search query (id)', () => {
        const result = filterTraces(mockTraces, 'trace-3', 'all', 'all');
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('trace-3');
    });

    it('should handle case-insensitive search', () => {
        const result = filterTraces(mockTraces, 'WEATHER', 'all', 'all');
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('trace-1');
    });

    it('should combine filters', () => {
        const result = filterTraces(mockTraces, 'trace', 'success', 'tool');
        // trace-1: success, tool -> match
        // trace-2: error, tool -> fail status
        // trace-3: success, service -> fail type
        expect(result).toHaveLength(1);
        expect(result[0].id).toBe('trace-1');
    });
});
