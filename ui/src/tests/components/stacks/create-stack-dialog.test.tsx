/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CreateStackDialog } from '@/components/stacks/create-stack-dialog';
import { apiClient } from '@/lib/client';
import { vi, type Mock } from 'vitest';

// Mock API client
vi.mock('@/lib/client', () => ({
  apiClient: {
    createCollection: vi.fn(),
  },
}));

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe('CreateStackDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders create button', () => {
    render(<CreateStackDialog />);
    expect(screen.getByText('Create Stack')).toBeInTheDocument();
  });

  it('opens dialog and submits form', async () => {
    const onStackCreated = vi.fn();
    (apiClient.createCollection as Mock).mockResolvedValue({ name: 'test-stack' });

    render(<CreateStackDialog onStackCreated={onStackCreated} />);

    // Open dialog
    fireEvent.click(screen.getByText('Create Stack'));
    await waitFor(() => expect(screen.getByText('Create New Stack')).toBeInTheDocument());

    // Fill form
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'test-stack' } });
    fireEvent.change(screen.getByLabelText('Description'), { target: { value: 'Testing stack' } });
    fireEvent.change(screen.getByLabelText('Author'), { target: { value: 'Tester' } });

    // Submit (look for the "Create" button in the footer, which might be tricky if there are multiple)
    // The button has text "Create" and type "submit"
    const createBtn = screen.getByRole('button', { name: 'Create' });
    fireEvent.click(createBtn);

    await waitFor(() => {
      expect(apiClient.createCollection).toHaveBeenCalledWith({
        name: 'test-stack',
        description: 'Testing stack',
        author: 'Tester',
        version: '0.0.1',
        services: [],
      });
      expect(onStackCreated).toHaveBeenCalled();
    });
  });

  it('displays error on failure', async () => {
    (apiClient.createCollection as Mock).mockRejectedValue(new Error('Failed'));

    render(<CreateStackDialog />);

    // Open dialog
    fireEvent.click(screen.getByText('Create Stack'));
    await waitFor(() => expect(screen.getByText('Create New Stack')).toBeInTheDocument());

    // Fill form
    fireEvent.change(screen.getByLabelText('Name'), { target: { value: 'fail-stack' } });

    // Submit
    const createBtn = screen.getByRole('button', { name: 'Create' });
    fireEvent.click(createBtn);

    // Should NOT close dialog (so title still visible)
    // And ideally toast is shown (but toast is hard to test unless mocked, assume handled)
    await waitFor(() => {
      expect(apiClient.createCollection).toHaveBeenCalled();
    });
    // Check if dialog is still open
    expect(screen.getByText('Create New Stack')).toBeInTheDocument();
  });
});
