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
  },
}));

describe('SystemHealth', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders loading state initially', () => {
        // Mock to return promise that doesn't resolve immediately
        (apiClient.getDoctorStatus as any).mockReturnValue(new Promise(() => {}));

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
});
