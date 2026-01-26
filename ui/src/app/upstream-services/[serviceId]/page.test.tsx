import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import UpstreamServiceDetailPage from './page';
import { apiClient } from '@/lib/client';

// Mock imports
vi.mock('next/navigation', () => ({
  useParams: () => ({ serviceId: 'test-service' }),
  useRouter: () => ({ push: vi.fn() }),
}));

vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({ toast: vi.fn() }),
}));

vi.mock('@/lib/client', () => ({
  apiClient: {
    getService: vi.fn(),
    validateService: vi.fn(),
    updateService: vi.fn(),
    setServiceStatus: vi.fn(),
    unregisterService: vi.fn(),
  },
}));

describe('UpstreamServiceDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (apiClient.getService as any).mockResolvedValue({
      service: {
        id: 'test-service',
        name: 'Test Service',
        httpService: { address: 'http://example.com' },
        lastError: null,
      },
    });
  });

  it('renders service details', async () => {
    render(<UpstreamServiceDetailPage />);
    await waitFor(() => {
        expect(screen.getByText('Test Service')).toBeInTheDocument();
    });
  });

  it('switches to diagnostics tab and runs validation', async () => {
    const user = userEvent.setup();
    (apiClient.validateService as any).mockResolvedValue({ valid: true });

    render(<UpstreamServiceDetailPage />);
    await waitFor(() => expect(screen.getByText('Test Service')).toBeInTheDocument());

    // Click Diagnostics Tab
    const tab = screen.getByRole('tab', { name: /diagnostics/i });
    await user.click(tab);

    // Click Run Diagnostics
    // Use findByRole to wait for it to appear
    const btn = await screen.findByRole('button', { name: /run diagnostics/i });
    await user.click(btn);

    await waitFor(() => {
        expect(apiClient.validateService).toHaveBeenCalled();
        expect(screen.getByText('Configuration Valid & Reachable')).toBeInTheDocument();
    });
  });
});
