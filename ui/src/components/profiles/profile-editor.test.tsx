/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ProfileEditor } from './profile-editor';
import { apiClient } from '@/lib/client';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Mock apiClient
vi.mock('@/lib/client', () => ({
  apiClient: {
    listServices: vi.fn(),
  },
}));

describe('ProfileEditor', () => {
  const mockServices = [
    { name: 'service-a', tags: ['dev'] },
    { name: 'service-b', tags: ['prod'] },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    (apiClient.listServices as any).mockResolvedValue({ services: mockServices });
  });

  it('renders correctly with default state', async () => {
    render(<ProfileEditor onSave={vi.fn()} onCancel={vi.fn()} />);

    expect(screen.getByLabelText('Profile Name')).toBeInTheDocument();
    expect(screen.getByLabelText('Selector Tags')).toHaveValue('dev');

    await waitFor(() => {
      expect(screen.getByText('service-a')).toBeInTheDocument();
      expect(screen.getByText('service-b')).toBeInTheDocument();
    });
  });

  it('loads profile data for editing', async () => {
    const profile = {
      name: 'test-profile',
      selector: { tags: ['prod', 'custom'] },
      serviceConfig: {
        'service-a': { enabled: true },
        'service-b': { enabled: false },
      },
    };

    render(<ProfileEditor profile={profile as any} onSave={vi.fn()} onCancel={vi.fn()} />);

    expect(screen.getByLabelText('Profile Name')).toHaveValue('test-profile');
    expect(screen.getByLabelText('Profile Name')).toBeDisabled();
    expect(screen.getByLabelText('Selector Tags')).toHaveValue('prod, custom');

    // Wait for services to load
    await waitFor(() => screen.getByText('service-a'));
  });

  it('calls onSave with correct data', async () => {
    const onSave = vi.fn();
    render(<ProfileEditor onSave={onSave} onCancel={vi.fn()} />);

    // Wait for services
    await waitFor(() => screen.getByText('service-a'));

    fireEvent.change(screen.getByLabelText('Profile Name'), { target: { value: 'new-profile' } });
    fireEvent.change(screen.getByLabelText('Selector Tags'), { target: { value: 'tag1' } });

    fireEvent.click(screen.getByText('Save Profile'));

    await waitFor(() => {
      expect(onSave).toHaveBeenCalledWith({
        name: 'new-profile',
        selector: { tags: ['tag1'] },
        serviceConfig: {},
        requiredRoles: undefined,
        parentProfileIds: undefined,
        secrets: undefined,
      });
    });
  });
});
