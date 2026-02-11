/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import SystemIntegrationsPage from './page';
import { apiClient } from '@/lib/client';

// Mock apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    getGlobalSettings: vi.fn(),
    saveGlobalSettings: vi.fn(),
  },
}));

// Mock hooks
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

describe('SystemIntegrationsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads settings on mount', async () => {
    (apiClient.getGlobalSettings as any).mockResolvedValue({
      alerts: { enabled: true, webhook_url: 'http://alert.com' },
      audit: { enabled: false, webhook_url: '' },
    });

    render(<SystemIntegrationsPage />);

    await waitFor(() => {
      expect(screen.getByText('System Integrations')).toBeInTheDocument();
      expect(screen.getByDisplayValue('http://alert.com')).toBeInTheDocument();
    });

    expect(apiClient.getGlobalSettings).toHaveBeenCalledTimes(1);
  });

  it('updates state when toggles are clicked', async () => {
    (apiClient.getGlobalSettings as any).mockResolvedValue({
      alerts: { enabled: false, webhook_url: '' },
      audit: { enabled: false, webhook_url: '' },
    });

    render(<SystemIntegrationsPage />);

    await waitFor(() => {
      expect(screen.getByText('System Integrations')).toBeInTheDocument();
    });

    // Find the toggle for Alerts (using label associated with switch)
    // The switch has id="alerts-enabled".
    // We can find by role="switch". The first one is alerts.
    const switches = screen.getAllByRole('switch');
    expect(switches.length).toBeGreaterThan(0);
    const alertsSwitch = switches[0];

    fireEvent.click(alertsSwitch);

    // Expect input to appear (since it only shows when enabled)
    await waitFor(() => {
      expect(screen.getByPlaceholderText('https://hooks.slack.com/services/...')).toBeInTheDocument();
    });
  });

  it('calls saveGlobalSettings with correct payload on save', async () => {
    (apiClient.getGlobalSettings as any).mockResolvedValue({
      alerts: { enabled: true, webhook_url: 'http://old.com' },
      audit: { enabled: false, webhook_url: '' },
    });

    render(<SystemIntegrationsPage />);

    await waitFor(() => {
      expect(screen.getByDisplayValue('http://old.com')).toBeInTheDocument();
    });

    // Update URL
    const input = screen.getByDisplayValue('http://old.com');
    fireEvent.change(input, { target: { value: 'http://new.com' } });

    // Enable Audit
    const switches = screen.getAllByRole('switch');
    const auditSwitch = switches[1]; // Second switch is audit
    fireEvent.click(auditSwitch);

    // Set Audit URL to ensure storage_type logic triggers (if applicable)
    // Note: Input appears after switch click
    await waitFor(() => {
        expect(screen.getByPlaceholderText('https://collector.example.com/v1/audit')).toBeInTheDocument();
    });
    const auditInput = screen.getByPlaceholderText('https://collector.example.com/v1/audit');
    fireEvent.change(auditInput, { target: { value: 'http://audit.com' } });

    // Click Save
    const saveButton = screen.getByText('Save Changes');
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(apiClient.saveGlobalSettings).toHaveBeenCalledWith(expect.objectContaining({
        alerts: { enabled: true, webhook_url: 'http://new.com' },
        audit: expect.objectContaining({
            enabled: true,
            webhook_url: 'http://audit.com',
            storage_type: 4 // Verify logic for auto-setting storage type
        })
      }));
    });
  });
});
