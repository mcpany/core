/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { GET } from './route';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock NextResponse
vi.mock('next/server', () => ({
  NextResponse: {
    json: vi.fn((data) => data),
  },
}));

// Mock fetch
global.fetch = vi.fn();

describe('GET /api/traces', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('should parse JSON-RPC tool calls and create nested spans', async () => {
    const mockEntries = [
      {
        id: '1',
        timestamp: '2023-01-01T00:00:00Z',
        method: 'POST',
        path: '/mcp',
        status: 200,
        duration: 100000000, // 100ms
        request_headers: {},
        response_headers: {},
        request_body: JSON.stringify({
          jsonrpc: '2.0',
          method: 'tools/call',
          params: {
            name: 'my_tool',
            arguments: { arg1: 'val1' },
          },
          id: 1,
        }),
        response_body: JSON.stringify({
          jsonrpc: '2.0',
          result: { output: 'success' },
          id: 1,
        }),
      },
    ];

    (global.fetch as any).mockResolvedValue({
      ok: true,
      json: async () => mockEntries,
    });

    const request = {
      headers: {
        get: () => '',
      },
    } as any;

    const response = await GET(request);

    // In our mock, NextResponse.json returns the data directly
    const traces = response as any;

    expect(traces).toHaveLength(1);
    const trace = traces[0];
    expect(trace.rootSpan.name).toBe('Execute Request');
    expect(trace.rootSpan.children).toHaveLength(1);

    const toolSpan = trace.rootSpan.children[0];
    expect(toolSpan.type).toBe('tool');
    expect(toolSpan.name).toBe('my_tool');
    expect(toolSpan.input).toEqual({ arg1: 'val1' });
    expect(toolSpan.output).toEqual({ output: 'success' });
  });

  it('should handle SSE response body', async () => {
    const mockEntries = [
      {
        id: '2',
        timestamp: '2023-01-01T00:00:00Z',
        method: 'POST',
        path: '/mcp',
        status: 200,
        duration: 100000000,
        request_headers: {},
        response_headers: {},
        request_body: '{}',
        response_body: 'event: message\ndata: {"result": "sse_success"}\n\n',
      },
    ];

    (global.fetch as any).mockResolvedValue({
      ok: true,
      json: async () => mockEntries,
    });

    const request = { headers: { get: () => '' } } as any;
    const response = await GET(request);
    const traces = response as any;

    expect(traces[0].rootSpan.output).toEqual({ result: 'sse_success' });
  });
});
