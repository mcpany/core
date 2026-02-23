/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { GET } from './route';
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock process.env
process.env.BACKEND_URL = 'http://backend:8080';

describe('GET /api/dashboard/health', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    global.fetch = vi.fn();
  });

  it('should proxy to backend and return data', async () => {
    const mockData = {
      services: [{ id: '1', name: 'test', status: 'healthy', latency: '10ms', uptime: '100%' }],
      history: { '1': [] }
    };

    (global.fetch as any).mockResolvedValue({
      ok: true,
      json: async () => mockData,
    });

    const request = new Request('http://localhost/api/dashboard/health');
    const response = await GET(request);
    const json = await response.json();

    expect(global.fetch).toHaveBeenCalledWith(
      'http://backend:8080/api/v1/dashboard/health',
      expect.objectContaining({
        cache: 'no-store'
      })
    );
    expect(json).toEqual(mockData);
  });

  it('should handle backend errors gracefully', async () => {
    (global.fetch as any).mockResolvedValue({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error'
    });

    const request = new Request('http://localhost/api/dashboard/health');
    const response = await GET(request);
    const json = await response.json();

    expect(response.status).toBe(500);
    expect(json).toEqual({ error: "Failed to fetch service health" });
  });

  it('should handle network errors', async () => {
    (global.fetch as any).mockRejectedValue(new Error('Network Error'));

    const request = new Request('http://localhost/api/dashboard/health');
    const response = await GET(request);
    const json = await response.json();

    // The catch block returns a specific fallback object
    expect(json).toEqual({ services: [], history: {} });
  });
});
