/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import WebhooksPage from './page';
import { apiClient } from '@/lib/client';

// Mock apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    getGlobalSettings: vi.fn(),
    saveGlobalSettings: vi.fn(),
  },
}));

// Mock useToast hook since it's used in the component
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

describe('WebhooksPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders and saves webhooks configuration', async () => {
    // Mock initial settings
    (apiClient.getGlobalSettings as any).mockResolvedValue({
      alerts: { enabled: false, webhook_url: '' },
      audit: { enabled: false, webhook_url: '', storage_type: 0 },
    });

    render(<WebhooksPage />);

    // Wait for loading to finish (Loader should disappear)
    await waitFor(() => expect(screen.queryByRole('heading', { name: /System Webhooks/i })).toBeInTheDocument());

    // Check if toggles are present. shadcn Switch uses role="switch".
    // We expect two switches: Alerts and Audit.
    const switches = screen.getAllByRole('switch');
    expect(switches).toHaveLength(2);
    const alertSwitch = switches[0];
    const auditSwitch = switches[1];

    expect(alertSwitch).not.toBeChecked();
    expect(auditSwitch).not.toBeChecked();

    // Enable Alerts
    fireEvent.click(alertSwitch);

    // Set Alert URL
    // Use getByLabelText or placeholder. Inputs have IDs matching Labels.
    const alertInput = screen.getByLabelText(/Webhook URL/i, { selector: '#alerts-url' });
    fireEvent.change(alertInput, { target: { value: 'http://alert.com' } });

    // Enable Audit
    fireEvent.click(auditSwitch);

    // Set Audit URL
    const auditInput = screen.getByLabelText(/Webhook URL/i, { selector: '#audit-url' });
    fireEvent.change(auditInput, { target: { value: 'http://audit.com' } });

    // Save
    const saveButton = screen.getByText(/Save Changes/i);
    fireEvent.click(saveButton);

    // Verify save call
    await waitFor(() => {
      expect(apiClient.saveGlobalSettings).toHaveBeenCalledWith(expect.objectContaining({
        alerts: expect.objectContaining({
          enabled: true,
          webhook_url: 'http://alert.com'
        }),
        audit: expect.objectContaining({
          enabled: true, // Audit toggle controls global enabled
          webhook_url: 'http://audit.com',
          storage_type: 4 // STORAGE_TYPE_WEBHOOK
        })
      }));
    });
  });
});
