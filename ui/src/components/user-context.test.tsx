/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor, act } from '@testing-library/react';
import { UserProvider, useUser } from './user-context';
import { apiClient } from '@/lib/client';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import React from 'react';

vi.mock('@/lib/client', () => ({
  apiClient: {
    getCurrentUser: vi.fn(),
    login: vi.fn(),
  },
}));

const TestComponent = () => {
  const { user, loading, login, logout } = useUser();
  if (loading) return <div>Loading...</div>;
  if (!user) return <button onClick={() => login('admin', 'password')}>Login</button>;
  return (
    <div>
      <div data-testid="user-name">{user.name}</div>
      <button onClick={logout}>Logout</button>
    </div>
  );
};

describe('UserContext', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it('loads user from token on mount', async () => {
    localStorage.setItem('mcp_auth_token', 'token');
    (apiClient.getCurrentUser as any).mockResolvedValue({
      id: '1',
      name: 'Test User',
      role: 'admin',
    });

    render(
      <UserProvider>
        <TestComponent />
      </UserProvider>
    );

    expect(screen.getByText('Loading...')).toBeInTheDocument();
    await waitFor(() => expect(screen.getByTestId('user-name')).toHaveTextContent('Test User'));
  });

  it('handles login', async () => {
    (apiClient.login as any).mockResolvedValue({ token: 'token' });
    (apiClient.getCurrentUser as any).mockResolvedValue({
      id: '1',
      name: 'Test User',
      role: 'admin',
    });

    render(
      <UserProvider>
        <TestComponent />
      </UserProvider>
    );

    // Initial state (no user)
    const loginButton = await screen.findByText('Login');

    await act(async () => {
        loginButton.click();
    });

    await waitFor(() => expect(screen.getByTestId('user-name')).toHaveTextContent('Test User'));
    expect(localStorage.getItem('mcp_auth_token')).toBe('token');
  });
});
