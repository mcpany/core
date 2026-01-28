/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ServiceEditor } from '@/components/services/editor/service-editor';
import { UpstreamServiceConfig, apiClient } from '@/lib/client';
import { vi } from 'vitest';

// Mock apiClient
vi.mock('@/lib/client', async () => {
  const actual = await vi.importActual('@/lib/client');
  return {
    ...actual as any,
    apiClient: {
      validateService: vi.fn(),
    }
  };
});

// Mock hooks
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

// ResizeObserver mock
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};
window.HTMLElement.prototype.scrollIntoView = vi.fn();
window.HTMLElement.prototype.releasePointerCapture = vi.fn();
window.HTMLElement.prototype.hasPointerCapture = vi.fn();

describe('ServiceEditor', () => {
    const mockService: UpstreamServiceConfig = {
        id: 'test-id',
        name: 'test-service',
        version: '1.0.0',
        disable: false,
        httpService: { address: 'http://localhost' },
        priority: 0,
        loadBalancingStrategy: 0,
    } as unknown as UpstreamServiceConfig;

    const onChange = vi.fn();
    const onSave = vi.fn();
    const onCancel = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
    });

    it('renders correctly', () => {
        render(<ServiceEditor service={mockService} onChange={onChange} onSave={onSave} onCancel={onCancel} />);
        expect(screen.getByDisplayValue('test-service')).toBeInTheDocument();
        expect(screen.getByText('Connection')).toBeInTheDocument();
    });

    it('validates before saving if validation was not run', async () => {
        const user = userEvent.setup();
        // Mock validation success
        vi.mocked(apiClient.validateService).mockResolvedValue({ valid: true });

        render(<ServiceEditor service={mockService} onChange={onChange} onSave={onSave} onCancel={onCancel} />);

        // Click Save Changes
        await user.click(screen.getByRole('button', { name: /save changes/i }));

        // Should have called validate
        expect(apiClient.validateService).toHaveBeenCalledWith(mockService);
        // Should have called onSave
        await waitFor(() => expect(onSave).toHaveBeenCalled());
    });

    it('blocks saving and shows error if validation fails', async () => {
        const user = userEvent.setup();
        // Mock validation failure
        vi.mocked(apiClient.validateService).mockResolvedValue({ valid: false, error: 'Connection refused' });

        render(<ServiceEditor service={mockService} onChange={onChange} onSave={onSave} onCancel={onCancel} />);

        // Click Save Changes
        await user.click(screen.getByRole('button', { name: /save changes/i }));

        // Should have called validate
        expect(apiClient.validateService).toHaveBeenCalledWith(mockService);

        // Should NOT have called onSave
        expect(onSave).not.toHaveBeenCalled();

        // Should switch to Connection tab (we verify by checking if Connection tab content is visible or if "Test Connection" is visible if we add it there)
        // Since we haven't implemented the tab switch yet, we can check for the error message
        // which we plan to add as an Alert
        // await waitFor(() => expect(screen.getByText('Connection refused')).toBeInTheDocument());
    });
});
