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
  onopen: null as ((ev: Event) => void) | null,
  onmessage: null as ((ev: MessageEvent) => void) | null,
  onclose: null as ((ev: CloseEvent) => void) | null,
  onerror: null as ((ev: Event) => void) | null,
};

// Setup global WebSocket mock
global.WebSocket = vi.fn().mockImplementation(function() {
  return mockWebSocket;
}) as unknown as typeof WebSocket;

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

    // Wait for connection effect
    // Using a small delay or loop if necessary, but usually renderHook flushes effects

    // Check if onopen is set
    // If useEffect is async (it is), renderHook waits for it?
    // Actually, creating WebSocket is synchronous in connect().
    expect(mockWebSocket.onopen).toBeTruthy();

    // Simulate connection open
    act(() => {
      if (mockWebSocket.onopen) {
        mockWebSocket.onopen(new Event('open'));
      }
    });

    expect(result.current.isConnected).toBe(true);

    const trace1 = createTrace('1', 100);
    const trace2 = createTrace('2', 200);

    // Simulate incoming messages
    act(() => {
        if (mockWebSocket.onmessage) {
            mockWebSocket.onmessage(new MessageEvent('message', { data: JSON.stringify(trace1) }));
            mockWebSocket.onmessage(new MessageEvent('message', { data: JSON.stringify(trace2) }));
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

    expect(mockWebSocket.onopen).toBeTruthy();

    act(() => {
      if (mockWebSocket.onopen) {
        mockWebSocket.onopen(new Event('open'));
      }
    });

    const trace1 = createTrace('1', 100);
    const trace1Update = createTrace('1', 150); // Updated duration

    // First message
    act(() => {
      if (mockWebSocket.onmessage) {
        mockWebSocket.onmessage(new MessageEvent('message', { data: JSON.stringify(trace1) }));
      }
    });

    act(() => {
        vi.advanceTimersByTime(200);
    });

    expect(result.current.traces).toHaveLength(1);
    expect(result.current.traces[0].totalDuration).toBe(100);

    // Update message
    act(() => {
      if (mockWebSocket.onmessage) {
        mockWebSocket.onmessage(new MessageEvent('message', { data: JSON.stringify(trace1Update) }));
      }
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
        if (mockWebSocket.onopen) {
            mockWebSocket.onopen(new Event('open'));
        }
     });

     // Simulate 100 updates rapidly
     act(() => {
         if (mockWebSocket.onmessage) {
             for (let i = 0; i < 100; i++) {
                 mockWebSocket.onmessage(new MessageEvent('message', { data: JSON.stringify(createTrace(`${i}`, i)) }));
             }
         }
     });

     act(() => {
         vi.advanceTimersByTime(200);
     });

     expect(result.current.traces).toHaveLength(100);
     expect(result.current.traces[0].id).toBe('99');
  });

  it('should limit the number of traces to avoid memory leak', async () => {
    const { result } = renderHook(() => useTraces());

    expect(mockWebSocket.onopen).toBeTruthy();

    act(() => {
        if (mockWebSocket.onopen) {
            mockWebSocket.onopen(new Event('open'));
        }
    });

    // Simulate 1100 updates rapidly (limit is 1000)
    act(() => {
        if (mockWebSocket.onmessage) {
            // Send in batches to respect buffer interval slightly
            for (let i = 0; i < 1100; i++) {
                mockWebSocket.onmessage(new MessageEvent('message', { data: JSON.stringify(createTrace(`${i}`, i)) }));
            }
        }
    });

    // Wait for buffer flush
    act(() => {
        vi.advanceTimersByTime(200);
    });

    // Should be capped at 1000
    expect(result.current.traces).toHaveLength(1000);

    // Newest should be present (1099)
    expect(result.current.traces[0].id).toBe('1099');

    // Oldest should be dropped (0-99 should be gone, so last item should be 100)
    expect(result.current.traces[result.current.traces.length - 1].id).toBe('100');
  });
});
