/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, waitFor, act } from '@testing-library/react';
import { ConfigHealthBanner } from './config-health-banner';
import { apiClient } from '@/lib/client';

// Mock the API client
jest.mock('@/lib/client', () => ({
  apiClient: {
    getDoctorStatus: jest.fn(),
  },
}));

describe('ConfigHealthBanner', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('renders nothing when health is ok', async () => {
    (apiClient.getDoctorStatus as jest.Mock).mockResolvedValue({
      status: 'ok',
      checks: {
        configuration: { status: 'ok' },
      },
    });

    render(<ConfigHealthBanner />);

    // Should not find the alert
    await waitFor(() => {
        expect(screen.queryByText('Configuration Error')).not.toBeInTheDocument();
    });
  });

  it('renders banner when configuration is degraded', async () => {
    (apiClient.getDoctorStatus as jest.Mock).mockResolvedValue({
      status: 'degraded',
      checks: {
        configuration: {
            status: 'degraded',
            message: 'YAML parse error'
        },
      },
    });

    render(<ConfigHealthBanner />);

    await waitFor(() => {
      expect(screen.getByText('Configuration Error')).toBeInTheDocument();
      expect(screen.getByText('YAML parse error')).toBeInTheDocument();
    });
  });

  it('polls for updates', async () => {
    (apiClient.getDoctorStatus as jest.Mock).mockResolvedValue({
        status: 'ok',
        checks: { configuration: { status: 'ok' } },
    });

    render(<ConfigHealthBanner />);

    expect(apiClient.getDoctorStatus).toHaveBeenCalledTimes(1);

    // Advance time by 5 seconds
    act(() => {
        jest.advanceTimersByTime(5000);
    });

    expect(apiClient.getDoctorStatus).toHaveBeenCalledTimes(2);
  });
});
