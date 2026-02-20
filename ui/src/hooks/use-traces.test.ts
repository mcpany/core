/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { renderHook, act } from '@testing-library/react';
import { useTraces, MAX_TRACES } from './use-traces';
import { Trace } from '@/types/trace';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

// Setup global WebSocket mock
const originalWebSocket = global.WebSocket;

describe('useTraces Hook', () => {
  let mockWebSocket: any;

  beforeEach(() => {
    mockWebSocket = {
      send: vi.fn(),
      close: vi.fn(),
      onopen: null,
      onmessage: null,
      onclose: null,
      onerror: null,
    };

    global.WebSocket = vi.fn().mockImplementation(function() {
      return mockWebSocket;
    }) as any;

    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    global.WebSocket = originalWebSocket;
    vi.clearAllMocks();
  });

  const createTrace = (id: string, duration: number): Trace => ({
    id,
    timestamp: new Date().toISOString(),
    totalDuration: duration,
    status: 'success',
    trigger: 'user',
    rootSpan: {
      id: `span-${id}`,
      name: 'test-operation',
      type: 'service',
      startTime: 0,
      endTime: duration,
      status: 'success'
    }
  });

  it('should accumulate incoming traces', async () => {
    const { result } = renderHook(() => useTraces());

    // Wait for connection effect
    expect(mockWebSocket.onopen).toBeDefined();

    // Simulate connection open
    act(() => {
      if (mockWebSocket.onopen) mockWebSocket.onopen({} as any);
    });

    expect(result.current.isConnected).toBe(true);

    const trace1 = createTrace('1', 100);
    const trace2 = createTrace('2', 200);

    // Simulate incoming messages
    act(() => {
        if (mockWebSocket.onmessage) {
            mockWebSocket.onmessage({ data: JSON.stringify(trace1) } as any);
            mockWebSocket.onmessage({ data: JSON.stringify(trace2) } as any);
        }
    });

    // Advance timers to trigger interval flush
    act(() => {
      vi.advanceTimersByTime(200);
    });

    expect(result.current.traces).toHaveLength(2);

    // Expect newest first (LIFO/prepended)
    expect(result.current.traces[0].id).toBe('2');
    expect(result.current.traces[1].id).toBe('1');
  });

  it('should deduplicate traces by ID', async () => {
    const { result } = renderHook(() => useTraces());

    expect(mockWebSocket.onopen).toBeDefined();

    act(() => {
      if (mockWebSocket.onopen) mockWebSocket.onopen({} as any);
    });

    const trace1 = createTrace('1', 100);
    const trace1Update = createTrace('1', 150); // Updated duration

    // First message
    act(() => {
      if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(trace1) } as any);
    });

    act(() => {
        vi.advanceTimersByTime(200);
    });

    expect(result.current.traces).toHaveLength(1);
    expect(result.current.traces[0].totalDuration).toBe(100);

    // Update message
    act(() => {
      if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(trace1Update) } as any);
    });

    act(() => {
        vi.advanceTimersByTime(200);
    });

    expect(result.current.traces).toHaveLength(1);
    expect(result.current.traces[0].totalDuration).toBe(150);
  });

  it('should handle rapid updates efficiently (simulation)', async () => {
     const { result } = renderHook(() => useTraces());

     expect(mockWebSocket.onopen).toBeDefined();

     act(() => {
        if (mockWebSocket.onopen) mockWebSocket.onopen({} as any);
     });

     // Simulate 100 updates rapidly
     act(() => {
         for (let i = 0; i < 100; i++) {
             if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(createTrace(`${i}`, i)) } as any);
         }
     });

     act(() => {
         vi.advanceTimersByTime(200);
     });

     expect(result.current.traces).toHaveLength(100);
     expect(result.current.traces[0].id).toBe('99');
  });

  it('should enforce MAX_TRACES limit', async () => {
    const { result } = renderHook(() => useTraces());

    expect(mockWebSocket.onopen).toBeDefined();

    act(() => {
      if (mockWebSocket.onopen) mockWebSocket.onopen({} as any);
    });

    // Simulate sending MAX_TRACES + 50 items
    act(() => {
      for (let i = 0; i < MAX_TRACES + 50; i++) {
        if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(createTrace(`${i}`, i)) } as any);
      }
    });

    act(() => {
      vi.advanceTimersByTime(200);
    });

    expect(result.current.traces).toHaveLength(MAX_TRACES);
    // Expect the newest items to be present.
    // The last inserted item was id `${MAX_TRACES + 49}`.
    // Since items are prepended (reversed), index 0 should be the newest.
    expect(result.current.traces[0].id).toBe(`${MAX_TRACES + 49}`);
  });
});
