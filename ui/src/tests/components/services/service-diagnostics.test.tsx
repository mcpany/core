/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ServiceDiagnostics } from '@/components/services/editor/service-diagnostics';
import { UpstreamServiceConfig } from '@/lib/client';
import { vi } from 'vitest';

// Mock apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    validateService: vi.fn(),
    getServiceStatus: vi.fn(),
    listTools: vi.fn(),
  },
}));

import { apiClient } from '@/lib/client';

describe('ServiceDiagnostics', () => {
  const mockService: UpstreamServiceConfig = {
    id: 'test-id',
    name: 'test-service',
    version: '1.0.0',
    disable: false,
    httpService: { address: 'http://localhost' },
    priority: 0,
    loadBalancingStrategy: 0,
  } as unknown as UpstreamServiceConfig;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders correctly', () => {
    render(<ServiceDiagnostics service={mockService} />);
    expect(screen.getByText('Service Diagnostics')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /run diagnostics/i })).toBeInTheDocument();
  });

  it('runs diagnostics successfully', async () => {
    const user = userEvent.setup();

    // Mock successful responses
    (apiClient.validateService as any).mockResolvedValue({ valid: true });
    (apiClient.getServiceStatus as any).mockResolvedValue({ status: 'Active' });
    (apiClient.listTools as any).mockResolvedValue({
      tools: [{ name: 'test-tool', serviceId: 'test-service' }]
    });

    render(<ServiceDiagnostics service={mockService} />);

    await user.click(screen.getByRole('button', { name: /run diagnostics/i }));

    await waitFor(() => {
        expect(apiClient.validateService).toHaveBeenCalledWith(mockService);
        expect(apiClient.getServiceStatus).toHaveBeenCalledWith('test-service');
        expect(apiClient.listTools).toHaveBeenCalled();
    });

    // Check results
    expect(screen.getByText('Configuration Validation')).toBeInTheDocument();
    expect(screen.getByText('Configuration is valid.')).toBeInTheDocument();
    expect(screen.getByText('Runtime Status')).toBeInTheDocument();
    expect(screen.getByText('Service is Active.')).toBeInTheDocument();
    expect(screen.getByText('Tool Discovery')).toBeInTheDocument();
    expect(screen.getByText('Discovered 1 tool(s).')).toBeInTheDocument();
  });

  it('handles validation failure', async () => {
    const user = userEvent.setup();
    (apiClient.validateService as any).mockResolvedValue({ valid: false, errors: ['Invalid URL'] });

    render(<ServiceDiagnostics service={mockService} />);

    await user.click(screen.getByRole('button', { name: /run diagnostics/i }));

    await waitFor(() => {
        expect(screen.getByText('Configuration is invalid.')).toBeInTheDocument();
    });
    expect(screen.getByText('Invalid URL')).toBeInTheDocument();
  });

  it('skips runtime checks if service is unsaved', async () => {
     const user = userEvent.setup();
     const unsavedService = { ...mockService, id: '', name: 'new-service' };
     (apiClient.validateService as any).mockResolvedValue({ valid: true });

     render(<ServiceDiagnostics service={unsavedService} />);
     await user.click(screen.getByRole('button', { name: /run diagnostics/i }));

     await waitFor(() => {
        expect(apiClient.validateService).toHaveBeenCalled();
        expect(apiClient.getServiceStatus).not.toHaveBeenCalled();
     });

     expect(screen.getByText('Skipped (Service not saved yet).')).toBeInTheDocument();
  });
});
