/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, act } from '@testing-library/react';
import { useTraces } from './use-traces';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('useTraces', () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    let mockWebSocket: any;
    const originalWebSocket = global.WebSocket;

    beforeEach(() => {
        mockWebSocket = {
            send: vi.fn(),
            close: vi.fn(),
            onopen: null,
            onmessage: null,
            onclose: null,
            onerror: null,
        };

        // Mock WebSocket constructor
        global.WebSocket = class {
            constructor() {
                return mockWebSocket;
            }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } as any;

        // Mock window.location
        Object.defineProperty(window, 'location', {
            value: {
                protocol: 'http:',
                host: 'localhost:3000',
            },
            writable: true,
        });
    });

    afterEach(() => {
        global.WebSocket = originalWebSocket;
    });

    it('should add new traces', async () => {
        const { result } = renderHook(() => useTraces());

        act(() => {
            if (mockWebSocket.onopen) mockWebSocket.onopen();
        });

        const trace1 = { id: '1', timestamp: '2024-01-01T00:00:00Z', message: 'test1' };

        act(() => {
            if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(trace1) });
        });

        expect(result.current.traces).toHaveLength(1);
        expect(result.current.traces[0]).toEqual(trace1);
    });

    it('should update existing traces (deduplication)', async () => {
        const { result } = renderHook(() => useTraces());

        act(() => {
            if (mockWebSocket.onopen) mockWebSocket.onopen();
        });

        const trace1 = { id: '1', timestamp: '2024-01-01T00:00:00Z', message: 'test1' };

        act(() => {
            if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(trace1) });
        });

        const trace1Updated = { ...trace1, message: 'updated' };

        act(() => {
            if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(trace1Updated) });
        });

        expect(result.current.traces).toHaveLength(1);
        expect(result.current.traces[0].message).toBe('updated');
    });

    it('should limit the number of traces', async () => {
        const { result } = renderHook(() => useTraces());

        act(() => {
            if (mockWebSocket.onopen) mockWebSocket.onopen();
        });

        const MAX_TRACES = 1000;

        // Add MAX_TRACES + 10 traces
        await act(async () => {
            for (let i = 0; i < MAX_TRACES + 10; i++) {
                const trace = { id: `trace-${i}`, timestamp: '2024-01-01T00:00:00Z', message: `msg-${i}` };
                if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(trace) });
            }
        });

        // This expectation should fail until the fix is implemented
        expect(result.current.traces.length).toBeLessThanOrEqual(MAX_TRACES);
    });
});
