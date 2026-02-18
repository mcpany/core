/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, act } from '@testing-library/react';
import { useTraces } from './use-traces';
import { Trace } from '@/types/trace';
import { vi, describe, it, expect, beforeEach, afterEach, beforeAll, afterAll } from 'vitest';

// Define a type for the mock WebSocket to avoid 'any' usage
type MockWebSocket = {
  send: ReturnType<typeof vi.fn>;
  close: ReturnType<typeof vi.fn>;
  onopen: ((this: WebSocket, ev: Event) => unknown) | null;
  onmessage: ((this: WebSocket, ev: MessageEvent) => unknown) | null;
  onclose: ((this: WebSocket, ev: CloseEvent) => unknown) | null;
  onerror: ((this: WebSocket, ev: Event) => unknown) | null;
};

// Mock WebSocket
const mockWebSocket: MockWebSocket = {
  send: vi.fn(),
  close: vi.fn(),
  onopen: null,
  onmessage: null,
  onclose: null,
  onerror: null,
};

describe('useTraces Hook', () => {
  let originalWebSocket: typeof WebSocket;

  beforeAll(() => {
    // Save original WebSocket
    originalWebSocket = global.WebSocket;
    // Mock global WebSocket
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    global.WebSocket = vi.fn().mockImplementation(function() { return mockWebSocket; }) as any;
  });

  afterAll(() => {
    // Restore original WebSocket
    global.WebSocket = originalWebSocket;
  });

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

    // Simulate connection open
    act(() => {
      if (mockWebSocket.onopen) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        mockWebSocket.onopen.call(mockWebSocket as any, {} as any);
      }
    });

    // We can't easily check isConnected state inside hook without exposing it or checking side effects
    // But the hook exposes isConnected
    expect(result.current.isConnected).toBe(true);

    const trace1 = createTrace('1', 100);
    const trace2 = createTrace('2', 200);

    // Simulate incoming messages
    act(() => {
        if (mockWebSocket.onmessage) {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            mockWebSocket.onmessage.call(mockWebSocket as any, { data: JSON.stringify(trace1) } as any);
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            mockWebSocket.onmessage.call(mockWebSocket as any, { data: JSON.stringify(trace2) } as any);
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

    act(() => {
      if (mockWebSocket.onopen) {
         // eslint-disable-next-line @typescript-eslint/no-explicit-any
         mockWebSocket.onopen.call(mockWebSocket as any, {} as any);
      }
    });

    const trace1 = createTrace('1', 100);
    const trace1Update = createTrace('1', 150); // Updated duration

    // First message
    act(() => {
      if (mockWebSocket.onmessage) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        mockWebSocket.onmessage.call(mockWebSocket as any, { data: JSON.stringify(trace1) } as any);
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
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        mockWebSocket.onmessage.call(mockWebSocket as any, { data: JSON.stringify(trace1Update) } as any);
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

     act(() => {
        if (mockWebSocket.onopen) {
             // eslint-disable-next-line @typescript-eslint/no-explicit-any
             mockWebSocket.onopen.call(mockWebSocket as any, {} as any);
        }
     });

     // Simulate 100 updates rapidly
     act(() => {
         if (mockWebSocket.onmessage) {
             for (let i = 0; i < 100; i++) {
                 // eslint-disable-next-line @typescript-eslint/no-explicit-any
                 mockWebSocket.onmessage.call(mockWebSocket as any, { data: JSON.stringify(createTrace(`${i}`, i)) } as any);
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

    act(() => {
      if (mockWebSocket.onopen) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        mockWebSocket.onopen.call(mockWebSocket as any, {} as any);
      }
    });

    // Simulate 1100 updates rapidly (over limit of 1000)
    act(() => {
      if (mockWebSocket.onmessage) {
        // Create traces with IDs 0 to 1099
        for (let i = 0; i < 1100; i++) {
           // eslint-disable-next-line @typescript-eslint/no-explicit-any
           mockWebSocket.onmessage.call(mockWebSocket as any, { data: JSON.stringify(createTrace(`${i}`, i)) } as any);
        }
      }
    });

    act(() => {
      vi.advanceTimersByTime(200);
    });

    // Should be capped at 1000
    expect(result.current.traces).toHaveLength(1000);
    // Newest trace should be at index 0. The last trace sent was 1099.
    expect(result.current.traces[0].id).toBe('1099');
    // The oldest trace should be 100 (since 0-99 fell off)
    expect(result.current.traces[999].id).toBe('100');
  });
});
