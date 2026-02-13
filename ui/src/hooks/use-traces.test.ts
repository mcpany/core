/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// @vitest-environment jsdom

import { renderHook, waitFor } from '@testing-library/react';
import { useTraces } from './use-traces';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock Trace type
interface Trace {
  id: string;
  timestamp: string;
  // other fields...
}

describe('useTraces', () => {
  let mockSocketInstance: any;

  beforeEach(() => {
    mockSocketInstance = {
      onopen: null,
      onmessage: null,
      onclose: null,
      onerror: null,
      close: vi.fn(),
    };

    // Mock WebSocket constructor
    vi.stubGlobal('WebSocket', class {
        constructor() {
            setTimeout(() => {
                if (mockSocketInstance.onopen) mockSocketInstance.onopen();
            }, 0);
            return mockSocketInstance;
        }
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('should connect and receive traces', async () => {
    const { result } = renderHook(() => useTraces());

    // Initially loading
    expect(result.current.loading).toBe(true);

    // Wait for connection
    await waitFor(() => expect(result.current.isConnected).toBe(true));
    expect(result.current.loading).toBe(false);

    // Simulate incoming trace
    const trace1 = { id: 't1', timestamp: '2023-01-01T00:00:00Z', name: 'trace1' };

    if (mockSocketInstance.onmessage) {
        mockSocketInstance.onmessage({ data: JSON.stringify(trace1) });
    }

    // Expect trace to be added
    await waitFor(() => {
        expect(result.current.traces).toHaveLength(1);
        expect(result.current.traces[0].id).toBe('t1');
    });

    // Simulate another trace
    const trace2 = { id: 't2', timestamp: '2023-01-01T00:00:01Z', name: 'trace2' };
    if (mockSocketInstance.onmessage) {
        mockSocketInstance.onmessage({ data: JSON.stringify(trace2) });
    }

    await waitFor(() => {
        expect(result.current.traces).toHaveLength(2);
        expect(result.current.traces[0].id).toBe('t2'); // Latest first
    });

    // Simulate update to existing trace
    const trace1Updated = { ...trace1, name: 'trace1-updated' };
    if (mockSocketInstance.onmessage) {
        mockSocketInstance.onmessage({ data: JSON.stringify(trace1Updated) });
    }

    await waitFor(() => {
        expect(result.current.traces).toHaveLength(2);
        const t1 = result.current.traces.find((t: any) => t.id === 't1');
        expect(t1.name).toBe('trace1-updated');
    });
  });

  it('should handle rapid updates correctly (batching simulation)', async () => {
    const { result } = renderHook(() => useTraces());
    await waitFor(() => expect(result.current.isConnected).toBe(true));

    // Simulate 100 messages rapidly
    for (let i = 0; i < 100; i++) {
        const trace = { id: `t${i}`, timestamp: new Date().toISOString(), name: `trace${i}` };
        if (mockSocketInstance.onmessage) {
            mockSocketInstance.onmessage({ data: JSON.stringify(trace) });
        }
    }

    // Wait for state to settle (with batching, this might take up to 100ms)
    await waitFor(() => {
        expect(result.current.traces.length).toBe(100);
    }, { timeout: 1000 });
  });
});
