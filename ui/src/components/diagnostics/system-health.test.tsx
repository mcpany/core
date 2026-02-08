/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { SystemHealth } from './system-health';
import { vi } from 'vitest';
import { apiClient } from '@/lib/client';

// Mock the API client
vi.mock('@/lib/client', () => ({
  apiClient: {
    getDoctorStatus: vi.fn(),
    getServiceHealth: vi.fn(),
  },
}));

describe('SystemHealth', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        (apiClient.getServiceHealth as any).mockResolvedValue({ services: [], history: {} });
    });

    it('renders loading state initially', () => {
        // Mock to return promise that doesn't resolve immediately
        (apiClient.getDoctorStatus as any).mockReturnValue(new Promise(() => {}));
        (apiClient.getServiceHealth as any).mockReturnValue(new Promise(() => {}));

        render(<SystemHealth />);
        expect(screen.getByText(/Running diagnostics.../i)).toBeInTheDocument();
    });

    it('renders healthy report correctly', async () => {
        const mockReport = {
            status: 'healthy',
            timestamp: new Date().toISOString(),
            checks: {
                'Core System': { status: 'ok', message: 'System operational', latency: '2ms' },
                'Database': { status: 'ok', message: 'Connected', latency: '10ms' },
            }
        };
        (apiClient.getDoctorStatus as any).mockResolvedValue(mockReport);

        render(<SystemHealth />);

        await waitFor(() => {
            // "Healthy" appears as text in the Badge
            // Note: getAllByText might be needed if "Healthy" appears multiple times
            expect(screen.getAllByText('Healthy')[0]).toBeInTheDocument();
            expect(screen.getByText('Core System')).toBeInTheDocument();
            expect(screen.getByText('System operational')).toBeInTheDocument();
            expect(screen.getByText('Database')).toBeInTheDocument();
        });
    });

    it('renders error state when api fails', async () => {
        (apiClient.getDoctorStatus as any).mockRejectedValue(new Error('Network error'));
        (apiClient.getServiceHealth as any).mockResolvedValue({ services: [], history: {} });

        render(<SystemHealth />);

        await waitFor(() => {
            expect(screen.getByText('Diagnostics Failed')).toBeInTheDocument();
            expect(screen.getByText(/Failed to retrieve diagnostics report/i)).toBeInTheDocument();
        });
    });

    it('renders degraded state correctly', async () => {
         const mockReport = {
            status: 'degraded',
            timestamp: new Date().toISOString(),
            checks: {
                'Core System': { status: 'ok', message: 'System operational', latency: '2ms' },
                'Upstream API': { status: 'degraded', message: 'High latency', latency: '500ms' },
            }
        };
        (apiClient.getDoctorStatus as any).mockResolvedValue(mockReport);

        render(<SystemHealth />);

         await waitFor(() => {
            // "Degraded" appears as text in the Badge
            expect(screen.getAllByText('Degraded')[0]).toBeInTheDocument();
            expect(screen.getByText('Upstream API')).toBeInTheDocument();
            expect(screen.getByText('High latency')).toBeInTheDocument();
        });
    });

    it('renders service health history timeline', async () => {
        const mockReport = {
            status: 'healthy',
            timestamp: new Date().toISOString(),
            checks: {}
        };
        const mockHealth = {
            services: [
                { id: 'srv1', name: 'My Service', status: 'healthy', latency: '10ms', uptime: '99%' }
            ],
            history: {
                'srv1': [
                    { timestamp: 1000, status: 'healthy' },
                    { timestamp: 2000, status: 'unhealthy' },
                    { timestamp: 3000, status: 'healthy' }
                ]
            }
        };

        (apiClient.getDoctorStatus as any).mockResolvedValue(mockReport);
        (apiClient.getServiceHealth as any).mockResolvedValue(mockHealth);

        render(<SystemHealth />);

        await waitFor(() => {
            expect(screen.getByText('Service Health History')).toBeInTheDocument();
            expect(screen.getByText('My Service')).toBeInTheDocument();
            // Check if timeline points are rendered (we can check for class names or tooltips if we could hover)
            // Ideally we check for elements with appropriate status colors
            // But since we can't easily check styles in basic jsdom without computing styles,
            // we assume presence of the container or elements is enough for this level.
        });
    });
});
