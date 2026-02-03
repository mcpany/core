/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AnalyticsDashboard } from '../../components/stats/analytics-dashboard';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { apiClient } from '../../lib/client';

// Mock Recharts
vi.mock('recharts', async () => {
    const actual = await vi.importActual('recharts');
    return {
        ...actual,
        ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
        PieChart: ({ children }: any) => <div data-testid="pie-chart">{children}</div>,
        Pie: ({ children }: any) => <div>{children}</div>,
        Cell: () => <div />,
        Tooltip: () => <div />,
        AreaChart: ({ children }: any) => <div>{children}</div>,
        Area: () => <div />,
        XAxis: () => <div />,
        YAxis: () => <div />,
        CartesianGrid: () => <div />,
        BarChart: ({ children }: any) => <div>{children}</div>,
        Bar: () => <div />,
        LineChart: ({ children }: any) => <div>{children}</div>,
        Line: () => <div />,
        Legend: () => <div />,
    };
});

// Mock apiClient
vi.mock('../../lib/client', () => ({
    apiClient: {
        getDashboardTraffic: vi.fn(),
        getTopTools: vi.fn(),
        listTools: vi.fn(),
        getToolUsage: vi.fn(),
        setToolStatus: vi.fn(),
    },
}));

// Mock ResizeObserver for Recharts
global.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
};

describe('AnalyticsDashboard', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('should render Context Usage tab and data', async () => {
        const user = userEvent.setup();

        // Mock API responses
        (apiClient.getDashboardTraffic as any).mockResolvedValue([
            { time: "10:00", requests: 100, latency: 50, errors: 2 }
        ]);
        (apiClient.getTopTools as any).mockResolvedValue([
            { name: "test_tool", count: 10 }
        ]);
        (apiClient.listTools as any).mockResolvedValue({
            tools: [
                { name: "heavy_tool", description: "A very heavy tool", serviceId: "service_a", inputSchema: { type: "object", properties: { huge: { type: "string" } } } },
                { name: "light_tool", description: "Light", serviceId: "service_b", inputSchema: { type: "object" } }
            ]
        });
        (apiClient.getToolUsage as any).mockResolvedValue([]);

        render(<AnalyticsDashboard />);

        // Wait for data to load
        await waitFor(() => {
            expect(screen.getByText('Analytics & Stats')).toBeInTheDocument();
        });

        // Check for Context Usage tab trigger
        const contextTabTrigger = screen.getByText('Context Usage');
        expect(contextTabTrigger).toBeInTheDocument();

        // Click the tab
        await user.click(contextTabTrigger);

        // Wait for Context content to appear
        await waitFor(() => {
            expect(screen.getByText('Total System Prompt')).toBeInTheDocument();
        });

        expect(screen.getByText('Context Usage by Service')).toBeInTheDocument();
        expect(screen.getByText('Heaviest Tools')).toBeInTheDocument();

        // Verify heaviest tools are listed
        expect(screen.getByText('heavy_tool')).toBeInTheDocument();
        expect(screen.getByText('light_tool')).toBeInTheDocument();
        expect(screen.getByText('service_a')).toBeInTheDocument();
        expect(screen.getByText('service_b')).toBeInTheDocument();

        // Verify API was called
        expect(apiClient.listTools).toHaveBeenCalled();
    });

    it('should render Optimization tab and identify ghost tools', async () => {
        const user = userEvent.setup();

        // Mock API responses
        (apiClient.getDashboardTraffic as any).mockResolvedValue([]);
        (apiClient.getTopTools as any).mockResolvedValue([]);
        (apiClient.listTools as any).mockResolvedValue({
            tools: [
                {
                    name: "ghost_tool",
                    description: "Heavy and unused",
                    serviceId: "service_a",
                    inputSchema: {
                        type: "object",
                        properties: {
                            huge: { type: "string", description: "x".repeat(3000) }
                        }
                    }
                },
                { name: "used_tool", description: "Used", serviceId: "service_b", inputSchema: { type: "object" } }
            ]
        });
        (apiClient.getToolUsage as any).mockResolvedValue([
            { name: "used_tool", serviceId: "service_b", totalCalls: 100, successRate: 100 },
            { name: "ghost_tool", serviceId: "service_a", totalCalls: 0, successRate: 0 }
        ]);

        render(<AnalyticsDashboard />);

        await waitFor(() => {
            expect(screen.getByText('Analytics & Stats')).toBeInTheDocument();
        });

        const optTab = screen.getByText('Optimization');
        await user.click(optTab);

        await waitFor(() => {
             expect(screen.getByText('Context Efficiency')).toBeInTheDocument();
        });

        // Should find ghost_tool in the list
        expect(screen.getByText('ghost_tool')).toBeInTheDocument();

        // Check for Disable button
        const disableButtons = screen.getAllByText('Disable');
        expect(disableButtons.length).toBeGreaterThan(0);

        // Click disable
        await user.click(disableButtons[0]);

        expect(apiClient.setToolStatus).toHaveBeenCalledWith("ghost_tool", true);
    });
});
