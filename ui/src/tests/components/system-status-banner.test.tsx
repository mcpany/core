/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from '@testing-library/react';
import { SystemStatusBanner } from '@/components/system-status-banner';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('SystemStatusBanner', () => {
  beforeEach(() => {
    // Clear mocks before each test
    vi.clearAllMocks();
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders nothing when status is healthy', async () => {
    (global.fetch as any).mockResolvedValue({
      ok: true,
      json: async () => ({ status: 'healthy', checks: {} }),
    });

    render(<SystemStatusBanner />);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith('/doctor');
    });

    expect(screen.queryByText(/System Status/i)).not.toBeInTheDocument();
  });

  it('renders error banner when fetch fails', async () => {
    (global.fetch as any).mockRejectedValue(new Error('Network error'));

    render(<SystemStatusBanner />);

    await waitFor(() => {
      expect(screen.getByText(/Connection Error/i)).toBeInTheDocument();
      expect(screen.getByText(/Network error/i)).toBeInTheDocument();
    });
  });

  it('renders degraded banner when status is degraded', async () => {
    (global.fetch as any).mockResolvedValue({
      ok: true,
      json: async () => ({
        status: 'degraded',
        checks: {
          internet: { status: 'failed', message: 'No internet connection' },
          configuration: { status: 'ok' }
        },
      }),
    });

    render(<SystemStatusBanner />);

    await waitFor(() => {
      expect(screen.getByText(/System Status: Degraded/i)).toBeInTheDocument();
      expect(screen.getByText(/Internet: No internet connection/i)).toBeInTheDocument();
    });
  });

  it('polls repeatedly', async () => {
    vi.useFakeTimers();
    (global.fetch as any).mockResolvedValue({
      ok: true,
      json: async () => ({ status: 'healthy', checks: {} }),
    });

    render(<SystemStatusBanner />);

    expect(global.fetch).toHaveBeenCalledTimes(1);

    await vi.advanceTimersByTimeAsync(30000);

    expect(global.fetch).toHaveBeenCalledTimes(2);

    vi.useRealTimers();
  });
});
