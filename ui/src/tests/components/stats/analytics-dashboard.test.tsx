/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { render, screen } from '@testing-library/react';
import React from 'react';
import { AnalyticsDashboard } from '../../../components/stats/analytics-dashboard';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import { apiClient } from '@/lib/client';

// Mock apiClient
vi.mock("@/lib/client", () => ({
  apiClient: {
    getDashboardTraffic: vi.fn(),
    getTopTools: vi.fn(),
    listTools: vi.fn(),
  },
}));

// Mock recharts to avoid rendering issues in JSDOM
vi.mock('recharts', async () => {
  const OriginalModule = await vi.importActual('recharts');
  return {
    ...OriginalModule,
    ResponsiveContainer: (props: any) => React.createElement('div', { style: { width: 800, height: 800 } }, props.children),
  };
});

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe('AnalyticsDashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (apiClient.getDashboardTraffic as any).mockResolvedValue([]);
    (apiClient.getTopTools as any).mockResolvedValue([]);
    (apiClient.listTools as any).mockResolvedValue({ tools: [] });
  });

  it('renders the dashboard title', () => {
    render(<AnalyticsDashboard />);
    expect(screen.getByText('Analytics & Stats')).toBeInTheDocument();
    expect(screen.getByText('Real-time insights into your MCP infrastructure.')).toBeInTheDocument();
  });

  it('renders key metrics cards', () => {
    render(<AnalyticsDashboard />);
    expect(screen.getByText('Total Requests')).toBeInTheDocument();
    expect(screen.getByText('Avg Throughput')).toBeInTheDocument();
    expect(screen.getByText('Avg Latency')).toBeInTheDocument();
    expect(screen.getByText('Error Rate')).toBeInTheDocument();
  });

  it('renders the time range selector', () => {
    render(<AnalyticsDashboard />);
    expect(screen.getByText('Last 1 Hour')).toBeInTheDocument();
  });

  it('renders tabs', () => {
    render(<AnalyticsDashboard />);
    expect(screen.getByText('Overview')).toBeInTheDocument();
    expect(screen.getByText('Performance')).toBeInTheDocument();
    expect(screen.getByText('Errors')).toBeInTheDocument();
  });
});
