/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ProfileManager } from './profile-manager';
import { apiClient } from '@/lib/client';

// Mock the apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    listProfiles: vi.fn(),
    createProfile: vi.fn(),
    updateProfile: vi.fn(),
    deleteProfile: vi.fn(),
  },
}));

// Mock toast
vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

// Mock ResizeObserver
class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
global.ResizeObserver = ResizeObserver;

describe('ProfileManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders profile list', async () => {
    const mockProfiles = [
      { name: 'dev', requiredRoles: ['admin'], selector: { tags: ['env:dev'] } },
      { name: 'prod', requiredRoles: [], selector: { tags: ['env:prod'] } },
    ];
    (apiClient.listProfiles as any).mockResolvedValue(mockProfiles);

    render(<ProfileManager />);

    expect(await screen.findByText('dev')).toBeDefined();
    expect(screen.getByText('prod')).toBeDefined();
    expect(screen.getByText('admin')).toBeDefined();
  });

  it('opens create dialog', async () => {
    (apiClient.listProfiles as any).mockResolvedValue([]);
    render(<ProfileManager />);

    // Wait for loading to finish (or mock it to be fast)
    // The component sets loading=true initially.
    await waitFor(() => expect(apiClient.listProfiles).toHaveBeenCalled());

    const createBtn = screen.getByText('Create Profile');
    fireEvent.click(createBtn);

    expect(await screen.findByText('Define a new execution profile.')).toBeDefined();
    expect(screen.getByLabelText('Profile Name')).toBeDefined();
  });
});
