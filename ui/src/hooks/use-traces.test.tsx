/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, act, waitFor } from '@testing-library/react';
import { useTraces } from './use-traces';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { Trace } from '@/types/trace';

// Mock types
const MOCK_TRACE_1: Trace = {
    id: "tr_1",
    rootSpan: { id: "sp_1", name: "test_tool_1", type: "tool", startTime: 1000, endTime: 1100, status: "success" },
    timestamp: new Date(1000).toISOString(),
    totalDuration: 100,
    status: "success",
    trigger: "user"
};

const MOCK_TRACE_2: Trace = {
    id: "tr_2",
    rootSpan: { id: "sp_2", name: "test_tool_2", type: "tool", startTime: 2000, endTime: 2200, status: "error" },
    timestamp: new Date(2000).toISOString(),
    totalDuration: 200,
    status: "error",
    trigger: "webhook"
};

describe('useTraces', () => {
    let mockWebSocket: any;
    let globalFetch: any;

    beforeEach(() => {
        // Mock Fetch
        globalFetch = vi.fn(() => Promise.resolve({
            ok: true,
            json: () => Promise.resolve([MOCK_TRACE_1])
        }));
        global.fetch = globalFetch;

        // Mock WebSocket
        mockWebSocket = {
            send: vi.fn(),
            close: vi.fn(),
            onopen: null,
            onmessage: null,
            onclose: null,
            onerror: null,
        };
        global.WebSocket = vi.fn(function() { return mockWebSocket; }) as any;
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    it('fetches history on mount', async () => {
        const { result } = renderHook(() => useTraces({ fetchHistory: true }));

        expect(result.current.loading).toBe(true);

        await waitFor(() => expect(globalFetch).toHaveBeenCalledTimes(1));

        // Wait for state update
        await waitFor(() => expect(result.current.traces).toHaveLength(1));
        expect(result.current.traces[0].id).toBe("tr_1");
        expect(result.current.loading).toBe(false);
    });

    it('connects to websocket', async () => {
        const { result } = renderHook(() => useTraces({ fetchHistory: false }));

        await waitFor(() => expect(global.WebSocket).toHaveBeenCalled());

        // Simulate open
        act(() => {
            if (mockWebSocket.onopen) mockWebSocket.onopen();
        });

        expect(result.current.isConnected).toBe(true);
    });

    it('merges live traces via websocket', async () => {
        const { result } = renderHook(() => useTraces({ fetchHistory: true }));

        await waitFor(() => expect(result.current.traces).toHaveLength(1));

        // Simulate WS message with NEW trace
        act(() => {
            if (mockWebSocket.onmessage) {
                 mockWebSocket.onmessage({ data: JSON.stringify(MOCK_TRACE_2) });
            }
        });

        await waitFor(() => expect(result.current.traces).toHaveLength(2));
        // Should be sorted by timestamp desc (MOCK_TRACE_2 is newer)
        expect(result.current.traces[0].id).toBe("tr_2");
    });

    it('deduplicates traces (history vs live)', async () => {
        const { result } = renderHook(() => useTraces({ fetchHistory: true }));

        await waitFor(() => expect(result.current.traces).toHaveLength(1));

        // Simulate WS message with SAME trace (updated or duplicate)
        const updatedTrace = { ...MOCK_TRACE_1, status: "error" }; // Changed status
        act(() => {
             if (mockWebSocket.onmessage) {
                mockWebSocket.onmessage({ data: JSON.stringify(updatedTrace) });
             }
        });

        // Should still be 1 trace, but updated
        await waitFor(() => expect(result.current.traces).toHaveLength(1));
        expect(result.current.traces[0].status).toBe("error");
    });

    it('refresh reconnects and fetches history', async () => {
         const { result } = renderHook(() => useTraces({ fetchHistory: true }));

         await waitFor(() => expect(result.current.traces).toHaveLength(1));

         act(() => {
             result.current.refresh();
         });

         // Fetch should be called again
         await waitFor(() => expect(globalFetch).toHaveBeenCalledTimes(2));
         expect(mockWebSocket.close).toHaveBeenCalled();
    });
});
