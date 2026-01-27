/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from '@testing-library/react';
import { SystemStatusBanner } from '@/components/system-status-banner';
import { vi, describe, it, expect, beforeEach, afterEach, type Mock } from 'vitest';
import { apiClient } from '@/lib/client';

// Mock apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    getDoctorStatus: vi.fn(),
  },
}));

describe('SystemStatusBanner', () => {
  beforeEach(() => {
    // Clear mocks before each test
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders nothing when status is healthy', async () => {
    (apiClient.getDoctorStatus as Mock).mockResolvedValue({ status: 'healthy', checks: {} });

    render(<SystemStatusBanner />);

    await waitFor(() => {
      expect(apiClient.getDoctorStatus).toHaveBeenCalled();
    });

    expect(screen.queryByText(/System Status/i)).not.toBeInTheDocument();
  });

  it('renders error banner when fetch fails', async () => {
    (apiClient.getDoctorStatus as Mock).mockRejectedValue(new Error('Network error'));

    render(<SystemStatusBanner />);

    await waitFor(() => {
      expect(screen.getByText(/Connection Error/i)).toBeInTheDocument();
      expect(screen.getByText(/Network error/i)).toBeInTheDocument();
    });
  });

  it('renders degraded banner when status is degraded', async () => {
    (apiClient.getDoctorStatus as Mock).mockResolvedValue({
      status: 'degraded',
      checks: {
        internet: { status: 'failed', message: 'No internet connection' },
        configuration: { status: 'ok' }
      },
    });

    render(<SystemStatusBanner />);

    await waitFor(() => {
      expect(screen.getByText(/System Status: Degraded/i)).toBeInTheDocument();
      expect(screen.getByText(/Internet: No internet connection/i)).toBeInTheDocument();
    });
  });

  it('polls repeatedly', async () => {
    vi.useFakeTimers();
    (apiClient.getDoctorStatus as Mock).mockResolvedValue({ status: 'healthy', checks: {} });

    render(<SystemStatusBanner />);

    expect(apiClient.getDoctorStatus).toHaveBeenCalledTimes(1);

    await vi.advanceTimersByTimeAsync(30000);

    // Should have called multiple times (interval is 5s)
    expect(apiClient.getDoctorStatus).toHaveBeenCalledTimes(7); // 1 initial + 6 polls (30/5)

    vi.useRealTimers();
  });
});
