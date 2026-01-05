/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { render, screen } from '@testing-library/react';
import React from 'react';
import { AnalyticsDashboard } from '../../../src/components/stats/analytics-dashboard';
import { vi } from 'vitest';

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
  it('renders the dashboard title', async () => {
    render(<AnalyticsDashboard />);
    expect(await screen.findByText('Analytics & Stats')).toBeInTheDocument();
    expect(screen.getByText('Real-time insights into your MCP infrastructure.')).toBeInTheDocument();
  });

  it('renders key metrics cards', async () => {
    render(<AnalyticsDashboard />);
    expect(await screen.findByText('Total Requests')).toBeInTheDocument();
    expect(screen.getByText('Avg Latency')).toBeInTheDocument();
    expect(screen.getByText('Error Rate')).toBeInTheDocument();
    // 'Active Services' might be different or dynamic, check component code? It is hardcoded in tabs overview.
    // Wait, Card title 'Active Services' is not in the code I read?
    // I read: Total Requests, Avg Throughput, Avg Latency, Error Rate.
    // Active Services is not in the top cards. It might be elsewhere?
    // Let's remove Active Services check if it's not there, or check 'Avg Throughput'.
    expect(screen.getByText('Avg Throughput')).toBeInTheDocument();
  });

  it('renders the time range selector', async () => {
    render(<AnalyticsDashboard />);
    expect(await screen.findByText('Last 1 Hour')).toBeInTheDocument();
  });

  it('renders tabs', async () => {
    render(<AnalyticsDashboard />);
    expect(await screen.findByText('Overview')).toBeInTheDocument();
    expect(screen.getByText('Performance')).toBeInTheDocument();
    expect(screen.getByText('Errors')).toBeInTheDocument();
  });
});
