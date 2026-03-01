/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { SecretsManager } from './secrets-manager';
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

// Mock the API client
vi.mock('@/lib/client', () => ({
    apiClient: {
        listSecrets: vi.fn().mockResolvedValue([]),
        saveSecret: vi.fn().mockResolvedValue({}),
        deleteSecret: vi.fn().mockResolvedValue({})
    }
}));

describe('SecretsManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders without crashing', async () => {
    render(<SecretsManager />);
    await waitFor(() => {
        expect(screen.getByText('Manage secure credentials for your upstream services.')).toBeInTheDocument();
    });
  });

  it('has an Import .env button', async () => {
    render(<SecretsManager />);
    await waitFor(() => {
        expect(screen.getByRole('button', { name: /Import \.env/i })).toBeInTheDocument();
    });
  });
});
