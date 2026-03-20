/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SecretsManager } from '../../components/settings/secrets-manager';
import { apiClient } from '@/lib/client';

// Mock useToast
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

describe('SecretsManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    global.fetch = vi.fn();
  });

  it('renders correctly and loads secrets', async () => {
    const mockSecrets = [
      { id: '1', name: 'Test Secret', key: 'TEST_KEY', value: 'secret-value', provider: 'custom', createdAt: '2023-01-01' },
    ];
    (global.fetch as any).mockImplementation(async (url: string) => {
        if (url.includes('/secrets')) {
            return { ok: true, json: async () => ({ secrets: mockSecrets }) };
        }
        return { ok: true, json: async () => ({}) };
    });

    render(<SecretsManager />);

    expect(screen.getByText('Loading secrets...')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText('Test Secret')).toBeInTheDocument();
    });
    expect(screen.getByText('TEST_KEY')).toBeInTheDocument();
  });

  it('allows adding a new secret', async () => {
    (global.fetch as any).mockImplementation(async (url: string, options: any) => {
        if (url.includes('/secrets') && (!options || options.method === 'GET')) {
            return { ok: true, json: async () => ({ secrets: [] }) };
        }
        if (url.includes('/secrets') && options && options.method === 'POST') {
            return { ok: true, json: async () => ({ id: 'new_id' }) };
        }
        return { ok: true, json: async () => ({}) };
    });
    render(<SecretsManager />);

    await waitFor(() => {
      expect(screen.queryByText('Loading secrets...')).not.toBeInTheDocument();
    });

    const user = userEvent.setup();

    // Open dialog
    await user.click(screen.getByText('Add Secret'));
    const dialog = await screen.findByRole('dialog');

    // Fill form
    await user.type(within(dialog).getByPlaceholderText('e.g. Production OpenAI Key'), 'New API Key');
    await user.type(within(dialog).getByPlaceholderText('e.g. OPENAI_API_KEY'), 'OPENAI_KEY');
    await user.type(within(dialog).getByPlaceholderText('sk-...'), 'sk-12345');

    // Save
    await user.click(within(dialog).getByRole('button', { name: 'Save Secret' }));

    await waitFor(() => {
      // Mocking global.fetch verifies we tried to save, and checking UI confirms flow.
      expect(screen.getByText('New API Key')).toBeInTheDocument();
    }, { timeout: 15000 });
  }, 20000);

  it('allows deleting a secret', async () => {
    const mockSecrets = [
      { id: '1', name: 'Delete Me', key: 'DELETE_KEY', value: 'secret-value', provider: 'custom', createdAt: '2023-01-01' },
    ];
    (global.fetch as any).mockImplementation(async (url: string, options: any) => {
        if (url.includes('/secrets') && (!options || options.method === 'GET')) {
            return { ok: true, json: async () => ({ secrets: mockSecrets }) };
        }
        if (url.includes('/secrets/1') && options && options.method === 'DELETE') {
            return { ok: true, json: async () => ({}) };
        }
        return { ok: true, json: async () => ({}) };
    });

    render(<SecretsManager />);

    await waitFor(() => {
      expect(screen.getByText('Delete Me')).toBeInTheDocument();
    });

    const user = userEvent.setup();
    const deleteBtn = screen.getByLabelText('Delete secret');
    await user.click(deleteBtn);

    await waitFor(() => {
      expect(screen.queryByText('Delete Me')).not.toBeInTheDocument();
    });
  });
});
