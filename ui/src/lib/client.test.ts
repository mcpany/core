/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { apiClient } from './client';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('apiClient Request Deduplication', () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    fetchMock.mockReset();
    fetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({}),
      text: async () => "{}",
    } as Response);
    vi.stubGlobal('fetch', fetchMock);

    // Clear localStorage mock if needed
    vi.stubGlobal('localStorage', {
      getItem: vi.fn(),
      setItem: vi.fn(),
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('should call fetch ONCE for concurrent requests AFTER optimization', async () => {
    // This test documents optimized behavior (deduplication)
    // We expect fetch to be called ONCE.

    const p1 = apiClient.getTopology();
    const p2 = apiClient.getTopology();

    await Promise.all([p1, p2]);

    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
