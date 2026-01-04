
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ServiceMarketplace } from './service-marketplace';
import { apiClient } from '@/lib/client';
import { toast } from 'sonner';

import { vi, describe, it, expect } from 'vitest';

// Mock dependencies
vi.mock('@/lib/client', () => ({
    apiClient: {
        registerService: vi.fn().mockResolvedValue({}),
    },
}));

vi.mock('sonner', () => ({
    toast: {
        success: vi.fn(),
        error: vi.fn(),
    },
}));

describe('ServiceMarketplace', () => {
    it('renders marketplace items', () => {
        render(<ServiceMarketplace />);

        // Check if some items are rendered
        expect(screen.getByText('Filesystem')).toBeInTheDocument();
        expect(screen.getByText('SQLite')).toBeInTheDocument();
    });

    it('filters items based on search query', () => {
        render(<ServiceMarketplace />);

        const searchInput = screen.getByPlaceholderText('Search services...');
        fireEvent.change(searchInput, { target: { value: 'SQLite' } });

        expect(screen.getByText('SQLite')).toBeInTheDocument();
        expect(screen.queryByText('Filesystem')).not.toBeInTheDocument();
    });

    it('opens install dialog when install button is clicked', async () => {
        render(<ServiceMarketplace />);

        const installButtons = screen.getAllByText('Install');
        fireEvent.click(installButtons[0]); // Click the first one (Filesystem)

        // Dialog title should appear
        expect(screen.getByText('Installing Filesystem')).toBeInTheDocument();
        expect(screen.getByText('ALLOWED_PATH')).toBeInTheDocument();
    });

    it('installs service with configuration', async () => {
        const onInstallComplete = vi.fn();
        render(<ServiceMarketplace onInstallComplete={onInstallComplete} />);

        // Find SQLite item and click Install
        // We can find by traversing or just knowing the order.
        // Let's filter first to be sure
        const searchInput = screen.getByPlaceholderText('Search services...');
        fireEvent.change(searchInput, { target: { value: 'SQLite' } });

        const installButton = screen.getByText('Install');
        fireEvent.click(installButton);

        // Fill in the required env var
        const dbPathInput = screen.getByLabelText(/DB_PATH/);
        fireEvent.change(dbPathInput, { target: { value: '/tmp/test.db' } });

        // Click Install in Dialog
        const confirmButton = screen.getByText('Install Service');
        fireEvent.click(confirmButton);

        await waitFor(() => {
            expect(apiClient.registerService).toHaveBeenCalled();
        });

        // Verify the arguments passed to registerService
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const calledArg = (apiClient.registerService as any).mock.calls[0][0];
        expect(calledArg.command_line_service.args).toContain('/tmp/test.db');
        expect(calledArg.command_line_service.env).toEqual({ DB_PATH: '/tmp/test.db' });

        expect(onInstallComplete).toHaveBeenCalled();
        expect(toast.success).toHaveBeenCalled();
    });
});
