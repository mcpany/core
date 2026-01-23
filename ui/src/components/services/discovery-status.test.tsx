import { render, screen, waitFor } from '@testing-library/react';
import { DiscoveryStatus } from './discovery-status';
import { apiClient } from '@/lib/client';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Mock apiClient
vi.mock('@/lib/client', () => ({
    apiClient: {
        getDiscoveryStatus: vi.fn(),
    },
}));

describe('DiscoveryStatus', () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders nothing when no providers are returned', async () => {
        (apiClient.getDiscoveryStatus as any).mockResolvedValue({ providers: [] });
        const { container } = render(<DiscoveryStatus />);
        await waitFor(() => {
            expect(container).toBeEmptyDOMElement();
        });
    });

    it('renders providers when returned', async () => {
        (apiClient.getDiscoveryStatus as any).mockResolvedValue({
            providers: [
                {
                    name: 'Ollama',
                    status: 'OK',
                    discoveredCount: 5,
                    lastRunAt: new Date().toISOString(),
                },
                {
                    name: 'Local',
                    status: 'ERROR',
                    lastError: 'Connection refused',
                    discoveredCount: 0,
                    lastRunAt: new Date().toISOString(),
                }
            ],
        });

        render(<DiscoveryStatus />);

        await waitFor(() => {
            expect(screen.getByText('Auto-Discovery Status')).toBeInTheDocument();
            expect(screen.getByText('Ollama')).toBeInTheDocument();
            expect(screen.getByText('(5 tools)')).toBeInTheDocument();
            expect(screen.getByText('Local')).toBeInTheDocument();
            expect(screen.getByText('Connection refused')).toBeInTheDocument();
        });
    });

    it('renders error state', async () => {
        (apiClient.getDiscoveryStatus as any).mockRejectedValue(new Error('Failed'));

        render(<DiscoveryStatus />);

        await waitFor(() => {
            expect(screen.getByText('Failed to load discovery status')).toBeInTheDocument();
        });
    });
});
