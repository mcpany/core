/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi } from 'vitest';
import WebhooksPage from './page';
import { apiClient } from '@/lib/client';
import { AuditConfig_StorageType } from '@proto/config/v1/config';

// Mock apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    getGlobalSettings: vi.fn(),
    saveGlobalSettings: vi.fn(),
  },
}));

describe('WebhooksPage', () => {
  const mockSettings = {
    alerts: {
      enabled: false,
      webhookUrl: '',
    },
    audit: {
      enabled: false,
      webhookUrl: '',
      storageType: AuditConfig_StorageType.STORAGE_TYPE_UNSPECIFIED,
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    (apiClient.getGlobalSettings as any).mockResolvedValue(mockSettings);
  });

  it('renders loading state initially', () => {
    // Mock delayed response
    (apiClient.getGlobalSettings as any).mockImplementation(() => new Promise(() => {}));
    render(<WebhooksPage />);
    // Since loading is true by default, we just check if content is not there
    expect(screen.queryByText('System Integrations')).not.toBeInTheDocument();
  });

  it('renders settings after loading', async () => {
    render(<WebhooksPage />);

    await waitFor(() => {
      expect(screen.getByText('System Integrations')).toBeInTheDocument();
    });

    expect(screen.getByText('System Alerts')).toBeInTheDocument();
    expect(screen.getByText('Audit Logging')).toBeInTheDocument();
  });

  it('updates alert settings and saves', async () => {
    render(<WebhooksPage />);
    await waitFor(() => screen.getByText('System Integrations'));

    const alertSwitch = screen.getAllByRole('switch')[0]; // First switch is Alerts
    const alertInput = screen.getByLabelText('Webhook URL');

    // Enable alerts
    fireEvent.click(alertSwitch);
    // Input URL
    fireEvent.change(alertInput, { target: { value: 'https://alert.example.com' } });

    // Save
    const saveButton = screen.getByText('Save Changes');
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(apiClient.saveGlobalSettings).toHaveBeenCalledWith(expect.objectContaining({
        alerts: {
          enabled: true,
          webhook_url: 'https://alert.example.com',
        },
      }));
    });
  });

  it('updates audit settings and enforces storage type', async () => {
    render(<WebhooksPage />);
    await waitFor(() => screen.getByText('System Integrations'));

    const auditSwitch = screen.getAllByRole('switch')[1]; // Second switch is Audit
    const auditInput = screen.getAllByLabelText(/URL/)[1]; // Second input (Collector URL)

    // Enable audit
    fireEvent.click(auditSwitch);
    // Input URL
    fireEvent.change(auditInput, { target: { value: 'https://audit.example.com' } });

    // Save
    const saveButton = screen.getByText('Save Changes');
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(apiClient.saveGlobalSettings).toHaveBeenCalledWith(expect.objectContaining({
        audit: expect.objectContaining({
          enabled: true,
          webhook_url: 'https://audit.example.com',
          storage_type: AuditConfig_StorageType.STORAGE_TYPE_WEBHOOK,
        }),
      }));
    });
  });
});
