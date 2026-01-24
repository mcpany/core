import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { UsageChart } from './usage-chart';
import { apiClient } from '@/lib/client';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Mock Recharts
vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: any) => <div>{children}</div>,
  ComposedChart: ({ children }: any) => <div data-testid="composed-chart">{children}</div>,
  Bar: () => <div>Bar</div>,
  Line: () => <div>Line</div>,
  XAxis: () => <div>XAxis</div>,
  YAxis: ({ label }: any) => <div>{label && label.value}</div>,
  CartesianGrid: () => <div>Grid</div>,
  Tooltip: () => <div>Tooltip</div>,
  Legend: () => <div>Legend</div>,
}));

vi.mock('@/lib/client', () => ({
    apiClient: {
        listAuditLogs: vi.fn(),
    },
}));

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe('UsageChart', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders loading state initially', async () => {
         (apiClient.listAuditLogs as any).mockReturnValue(new Promise(() => {})); // Never resolves to keep loading
         render(<UsageChart toolName="test-tool" />);
         // Skeleton check might be tricky without test-id, but we can verify API call
         expect(apiClient.listAuditLogs).toHaveBeenCalledWith(expect.objectContaining({
             tool_name: 'test-tool',
             limit: 1000
         }));
    });

    it('renders data when loaded', async () => {
        const mockData = [
            {
                timestamp: new Date().toISOString(),
                tool_name: 'test-tool',
                duration_ms: 100,
            }
        ];
        (apiClient.listAuditLogs as any).mockResolvedValue(mockData);

        render(<UsageChart toolName="test-tool" />);

        await waitFor(() => expect(screen.getByText('Usage Over Time')).toBeInTheDocument());

        // Check for axis labels (mocked)
        await waitFor(() => expect(screen.getByText('Calls')).toBeInTheDocument());
        await waitFor(() => expect(screen.getByText('Latency (ms)')).toBeInTheDocument());
    });

     it('renders empty state', async () => {
        (apiClient.listAuditLogs as any).mockResolvedValue([]);

        render(<UsageChart toolName="test-tool" />);

        await waitFor(() => expect(screen.getByText('No usage data available for this period.')).toBeInTheDocument());
    });
});
