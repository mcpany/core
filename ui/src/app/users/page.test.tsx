/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import UsersPage from './page';
import { apiClient } from '@/lib/client';
import { vi, Mock } from 'vitest';

// Mock apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    listUsers: vi.fn(),
    createUser: vi.fn(),
    updateUser: vi.fn(),
    deleteUser: vi.fn(),
  },
}));

// Mock crypto for API key generation
Object.defineProperty(global, 'crypto', {
  value: {
    getRandomValues: (arr: Uint8Array) => {
        for (let i = 0; i < arr.length; i++) {
            arr[i] = i % 256;
        }
        return arr;
    }
  }
});

// Mock confirm
global.confirm = vi.fn(() => true);

describe('UsersPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders users list', async () => {
    (apiClient.listUsers as Mock).mockResolvedValue([
      { id: 'testuser', roles: ['admin'] }
    ]);

    render(<UsersPage />);

    expect(screen.getByText('Users')).toBeInTheDocument();
    await waitFor(() => expect(screen.getByText('testuser')).toBeInTheDocument());
    expect(screen.getByText('admin')).toBeInTheDocument();
  });

  it('opens add user dialog', async () => {
    (apiClient.listUsers as Mock).mockResolvedValue([]);
    render(<UsersPage />);

    const addButton = screen.getByText('Add User');
    fireEvent.click(addButton);

    // shadcn dialog uses a portal, but jsdom usually handles finding text in document
    await waitFor(() => expect(screen.getByText('Create User')).toBeInTheDocument());
    expect(screen.getByText('Role')).toBeInTheDocument();
  });

  it('generates api key', async () => {
    const user = { id: 'testuser', roles: ['admin'], authentication: { api_key: { key_value: 'old' } } };
    (apiClient.listUsers as Mock).mockResolvedValue([user]);
    (apiClient.updateUser as Mock).mockResolvedValue({});

    render(<UsersPage />);

    await waitFor(() => expect(screen.getByText('testuser')).toBeInTheDocument());

    const generateBtn = screen.getByTitle('Generate API Key');
    fireEvent.click(generateBtn);

    expect(global.confirm).toHaveBeenCalled();

    await waitFor(() => {
        expect(apiClient.updateUser).toHaveBeenCalled();
    });

    // Expect the dialog title to appear
    expect(screen.getByText('API Key Generated')).toBeInTheDocument();
  });
});
