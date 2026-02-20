/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { renderHook, act } from '@testing-library/react';
import { useTraces } from './use-traces';
import { Trace } from '@/types/trace';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

// Mock WebSocket
const mockWebSocket = {
  send: vi.fn(),
  close: vi.fn(),
  onopen: null as any,
  onmessage: null as any,
  onclose: null as any,
  onerror: null as any,
};

// Setup global WebSocket mock
const originalWebSocket = global.WebSocket;
global.WebSocket = vi.fn().mockImplementation(function() {
  return mockWebSocket;
}) as any;

describe('useTraces Hook', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockWebSocket.onopen = null;
    mockWebSocket.onmessage = null;
    mockWebSocket.onclose = null;
    mockWebSocket.onerror = null;
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
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

    expect(mockWebSocket.onopen).toBeTruthy();

    act(() => {
      mockWebSocket.onopen({} as any);
    });

    expect(result.current.isConnected).toBe(true);

    const trace1 = createTrace('1', 100);
    const trace2 = createTrace('2', 200);

    act(() => {
        mockWebSocket.onmessage({ data: JSON.stringify(trace1) } as any);
        mockWebSocket.onmessage({ data: JSON.stringify(trace2) } as any);
    });

    act(() => {
      vi.advanceTimersByTime(200);
    });

    expect(result.current.traces).toHaveLength(2);
    expect(result.current.traces[0].id).toBe('2');
    expect(result.current.traces[1].id).toBe('1');
  });

  it('should deduplicate traces by ID', async () => {
    const { result } = renderHook(() => useTraces());

    expect(mockWebSocket.onopen).toBeTruthy();

    act(() => {
      mockWebSocket.onopen({} as any);
    });

    const trace1 = createTrace('1', 100);
    const trace1Update = createTrace('1', 150);

    act(() => {
      mockWebSocket.onmessage({ data: JSON.stringify(trace1) } as any);
    });

    act(() => {
        vi.advanceTimersByTime(200);
    });

    expect(result.current.traces).toHaveLength(1);
    expect(result.current.traces[0].totalDuration).toBe(100);

    act(() => {
      mockWebSocket.onmessage({ data: JSON.stringify(trace1Update) } as any);
    });

    act(() => {
        vi.advanceTimersByTime(200);
    });

    expect(result.current.traces).toHaveLength(1);
    expect(result.current.traces[0].totalDuration).toBe(150);
  });

  it('should handle rapid updates efficiently (simulation)', async () => {
     const { result } = renderHook(() => useTraces());

     expect(mockWebSocket.onopen).toBeTruthy();

     act(() => {
        mockWebSocket.onopen({} as any);
     });

     act(() => {
         for (let i = 0; i < 100; i++) {
             mockWebSocket.onmessage({ data: JSON.stringify(createTrace(`${i}`, i)) } as any);
         }
     });

     act(() => {
         vi.advanceTimersByTime(200);
     });

     expect(result.current.traces).toHaveLength(100);
     expect(result.current.traces[0].id).toBe('99');
  });

  it('should truncate traces when exceeding limit', async () => {
      const { result } = renderHook(() => useTraces());

      expect(mockWebSocket.onopen).toBeTruthy();

      act(() => {
         mockWebSocket.onopen({} as any);
      });

      // Add 1100 traces
      act(() => {
          for (let i = 0; i < 1100; i++) {
              mockWebSocket.onmessage({ data: JSON.stringify(createTrace(`${i}`, i)) } as any);
          }
      });

      act(() => {
          vi.advanceTimersByTime(200);
      });

      // Expect truncation to 1000
      expect(result.current.traces).toHaveLength(1000);
      // Ensure newest are kept (id 1099 should be first)
      expect(result.current.traces[0].id).toBe('1099');
  });
});
