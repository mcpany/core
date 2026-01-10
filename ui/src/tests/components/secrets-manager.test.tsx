/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SecretsManager } from '../../components/settings/secrets-manager';
import { apiClient } from '../../lib/client';

// Mock the apiClient
vi.mock('../../lib/client', () => ({
  apiClient: {
    listSecrets: vi.fn(),
    saveSecret: vi.fn(),
    deleteSecret: vi.fn(),
  },
}));

// Mock useToast
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

describe('SecretsManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders correctly and loads secrets', async () => {
    const mockSecrets = [
      { id: '1', name: 'Test Secret', key: 'TEST_KEY', value: 'secret-value', provider: 'custom', createdAt: '2023-01-01' },
    ];
    (apiClient.listSecrets as any).mockResolvedValue(mockSecrets);

    render(<SecretsManager />);

    expect(screen.getByText('Loading secrets...')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Test Secret')).toBeInTheDocument();
    });
    expect(screen.getByText('TEST_KEY')).toBeInTheDocument();
  });

  it('allows adding a new secret', async () => {
    (apiClient.listSecrets as any).mockResolvedValue([]);
    render(<SecretsManager />);

    await waitFor(() => {
        expect(screen.queryByText('Loading secrets...')).not.toBeInTheDocument();
    });

    const user = userEvent.setup();

    // Open dialog
    await user.click(screen.getByText('Add Secret'));

    // Fill form
    await user.type(screen.getByPlaceholderText('e.g. Production OpenAI Key'), 'New API Key');
    await user.type(screen.getByPlaceholderText('e.g. OPENAI_API_KEY'), 'OPENAI_KEY');
    await user.type(screen.getByPlaceholderText('sk-...'), 'sk-12345');

    // Save
    await user.click(screen.getByText('Save Secret'));

    await waitFor(() => {
      expect(apiClient.saveSecret).toHaveBeenCalledWith(expect.objectContaining({
        name: 'New API Key',
        key: 'OPENAI_KEY',
        value: 'sk-12345',
        provider: 'custom'
      }));
    });
  });

  it('allows deleting a secret', async () => {
     const mockSecrets = [
      { id: '1', name: 'Delete Me', key: 'DELETE_KEY', value: 'secret-value', provider: 'custom', createdAt: '2023-01-01' },
    ];
    (apiClient.listSecrets as any).mockResolvedValue(mockSecrets);

    render(<SecretsManager />);

    await waitFor(() => {
      expect(screen.getByText('Delete Me')).toBeInTheDocument();
    });

    const user = userEvent.setup();
    const deleteBtn = screen.getByLabelText('Delete secret');
    await user.click(deleteBtn);

    await waitFor(() => {
        expect(apiClient.deleteSecret).toHaveBeenCalledWith('1');
    });
  });
});
