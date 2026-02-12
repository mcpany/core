/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { DashboardShell } from '../../components/dashboard/dashboard-shell';
import { apiClient } from '../../lib/client';

// Mock apiClient
vi.mock('../../lib/client', () => ({
    apiClient: {
        listServices: vi.fn(),
        registerService: vi.fn(), // Needed if OnboardingView uses it
    },
}));

// Mock child components to isolate shell logic
vi.mock('../../components/dashboard/dashboard-grid', () => ({
    DashboardGrid: () => <div data-testid="dashboard-grid">Dashboard Grid</div>,
}));

vi.mock('../../components/dashboard/onboarding-view', () => ({
    OnboardingView: () => <div data-testid="onboarding-view">Onboarding View</div>,
}));

// Mock UI components used in Shell (Loader)
vi.mock('lucide-react', () => ({
    Loader2: () => <div data-testid="loader">Loading...</div>,
}));

describe('DashboardShell', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('should render loading state initially', () => {
        // Return a promise that never resolves immediately to test loading state?
        // Or just check that before waitFor, it shows loader.
        // Actually, useEffect runs after render.
        (apiClient.listServices as any).mockImplementation(() => new Promise(() => {}));
        render(<DashboardShell />);
        expect(screen.getByTestId('loader')).toBeInTheDocument();
    });

    it('should render OnboardingView when no services exist', async () => {
        (apiClient.listServices as any).mockResolvedValue({ services: [] });

        render(<DashboardShell />);

        await waitFor(() => {
            expect(screen.queryByTestId('loader')).not.toBeInTheDocument();
        });

        expect(screen.getByTestId('onboarding-view')).toBeInTheDocument();
        expect(screen.queryByTestId('dashboard-grid')).not.toBeInTheDocument();
    });

    it('should render DashboardGrid when services exist', async () => {
        (apiClient.listServices as any).mockResolvedValue({
            services: [{ name: 'test-service' }]
        });

        render(<DashboardShell />);

        await waitFor(() => {
            expect(screen.queryByTestId('loader')).not.toBeInTheDocument();
        });

        expect(screen.getByTestId('dashboard-grid')).toBeInTheDocument();
        expect(screen.queryByTestId('onboarding-view')).not.toBeInTheDocument();
    });

    it('should fallback to DashboardGrid on error', async () => {
        (apiClient.listServices as any).mockRejectedValue(new Error('API Error'));

        render(<DashboardShell />);

        await waitFor(() => {
            expect(screen.queryByTestId('loader')).not.toBeInTheDocument();
        });

        // Error state usually falls back to dashboard grid to show error widgets or empty state
        expect(screen.getByTestId('dashboard-grid')).toBeInTheDocument();
    });
});
