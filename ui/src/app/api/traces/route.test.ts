/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { GET } from './route';
import { NextResponse } from 'next/server';

// Mock global fetch
global.fetch = jest.fn();

describe('GET /api/traces', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    process.env.BACKEND_URL = 'http://test-backend';
  });

  it('should fetch traces from backend and return them sorted', async () => {
    const mockTraces = [
      { id: '1', timestamp: '2023-01-01T10:00:00Z' },
      { id: '2', timestamp: '2023-01-01T11:00:00Z' },
    ];

    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => mockTraces,
    });

    const req = new Request('http://localhost/api/traces');
    const res = await GET(req);
    const json = await res.json();

    expect(global.fetch).toHaveBeenCalledWith(
      'http://test-backend/api/v1/traces',
      expect.objectContaining({
        headers: expect.any(Headers),
      })
    );
    expect(json).toHaveLength(2);
    expect(json[0].id).toBe('2'); // Newest first
    expect(json[1].id).toBe('1');
  });

  it('should handle backend errors', async () => {
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
    });

    const req = new Request('http://localhost/api/traces');
    const res = await GET(req);
    const json = await res.json();

    expect(json).toEqual([]);
  });
});
