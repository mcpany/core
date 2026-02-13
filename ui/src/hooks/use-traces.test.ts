/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, act, waitFor } from '@testing-library/react';
import { useTraces } from './use-traces';
import { Trace } from '@/types/trace';
import { vi, describe, it, expect, beforeEach, afterEach, Mock } from 'vitest';

// Mock WebSocket
const mockWebSocket = {
  send: vi.fn(),
  close: vi.fn(),
  onopen: null as any,
  onmessage: null as any,
  onclose: null as any,
  onerror: null as any,
  readyState: WebSocket.OPEN,
};

// Global WebSocket mock
const originalWebSocket = global.WebSocket;

describe('useTraces', () => {
  beforeEach(() => {
    // Reset mocks
    mockWebSocket.send.mockClear();
    mockWebSocket.close.mockClear();
    mockWebSocket.onopen = null;
    mockWebSocket.onmessage = null;
    mockWebSocket.onclose = null;
    mockWebSocket.onerror = null;

    global.WebSocket = class {
        constructor() { return mockWebSocket; }
    } as any;
  });

  afterEach(() => {
    global.WebSocket = originalWebSocket;
  });

  it('should connect and receive traces', async () => {
    const { result } = renderHook(() => useTraces());

    // Simulate connection open
    act(() => {
      mockWebSocket.onopen?.({} as any);
    });

    expect(result.current.isConnected).toBe(true);

    // Simulate a trace message
    const trace: Trace = {
      id: 'trace-1',
      timestamp: new Date().toISOString(),
      source: 'test',
      operation: 'test-op',
      status: 'success',
      duration: 100,
      spans: []
    };

    act(() => {
      mockWebSocket.onmessage?.({ data: JSON.stringify(trace) } as any);
    });

    // Wait for update (since we might batch later, waitFor is good)
    await waitFor(() => {
        expect(result.current.traces).toHaveLength(1);
        expect(result.current.traces[0].id).toBe('trace-1');
    });
  });

  it('should handle rapid updates without crashing (performance simulation)', async () => {
    const { result } = renderHook(() => useTraces());

    act(() => {
      mockWebSocket.onopen?.({} as any);
    });

    // Simulate 1000 messages rapidly
    const messages = Array.from({ length: 1000 }, (_, i) => ({
      id: `trace-${i}`,
      timestamp: new Date().toISOString(),
      source: 'test',
      operation: 'test-op',
      status: 'success',
      duration: 100,
      spans: []
    }));

    await act(async () => {
        messages.forEach(msg => {
            mockWebSocket.onmessage?.({ data: JSON.stringify(msg) } as any);
        });
    });

    // With batching, this should eventually reflect 1000 traces.
    // Without batching, it also should, but might be slower (hard to test speed in unit test without benchmarks).
    // We mainly verify correctness here.

    await waitFor(() => {
        expect(result.current.traces.length).toBe(1000);
    }, { timeout: 3000 });
  });

  it('should deduplicate traces by ID', async () => {
    const { result } = renderHook(() => useTraces());

    act(() => {
      mockWebSocket.onopen?.({} as any);
    });

    const trace1: Trace = {
      id: 'trace-1',
      timestamp: '2023-01-01T00:00:00Z',
      source: 'test',
      operation: 'op1',
      status: 'pending',
      duration: 0,
      spans: []
    };

    const trace1Update: Trace = {
      id: 'trace-1',
      timestamp: '2023-01-01T00:00:01Z', // Updated
      source: 'test',
      operation: 'op1',
      status: 'success',
      duration: 100,
      spans: []
    };

    act(() => {
      mockWebSocket.onmessage?.({ data: JSON.stringify(trace1) } as any);
    });

    await waitFor(() => {
        expect(result.current.traces).toHaveLength(1);
        expect(result.current.traces[0].status).toBe('pending');
    });

    act(() => {
      mockWebSocket.onmessage?.({ data: JSON.stringify(trace1Update) } as any);
    });

    await waitFor(() => {
        expect(result.current.traces).toHaveLength(1);
        expect(result.current.traces[0].status).toBe('success');
    });
  });

  it('should respect pause state', async () => {
      const { result } = renderHook(() => useTraces({ initialPaused: true }));

      act(() => {
        mockWebSocket.onopen?.({} as any);
      });

      expect(result.current.isPaused).toBe(true);

      const trace: Trace = {
          id: 'trace-1',
          timestamp: new Date().toISOString(),
          source: 'test',
          operation: 'test-op',
          status: 'success',
          duration: 100,
          spans: []
        };

        act(() => {
          mockWebSocket.onmessage?.({ data: JSON.stringify(trace) } as any);
        });

        // Should NOT receive trace
        // Wait a bit to be sure
        await new Promise(r => setTimeout(r, 200));

        expect(result.current.traces).toHaveLength(0);

        // Unpause
        act(() => {
            result.current.setIsPaused(false);
        });

        // Send another
        act(() => {
             mockWebSocket.onmessage?.({ data: JSON.stringify(trace) } as any);
        });

        await waitFor(() => {
            expect(result.current.traces).toHaveLength(1);
        });
  });
});
