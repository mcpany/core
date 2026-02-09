/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { StackList } from '@/components/stacks/stack-list';
import { ServiceCollection } from '@/lib/marketplace-service';
import { vi, describe, it, expect } from 'vitest';

describe('StackList', () => {
  const mockStacks: ServiceCollection[] = [
    {
      name: 'stack-1',
      description: 'Test Stack 1',
      author: 'Test',
      version: '1.0.0',
      services: [{ name: 'svc1' } as any]
    },
    {
      name: 'stack-2',
      description: 'Test Stack 2',
      author: 'Test',
      version: '1.0.0',
      services: []
    }
  ];

  it('renders a list of stacks', () => {
    render(<StackList stacks={mockStacks} onDelete={() => {}} onDeploy={() => {}} />);
    expect(screen.getByText('stack-1')).toBeInTheDocument();
    expect(screen.getByText('stack-2')).toBeInTheDocument();
    expect(screen.getByText('1 Services')).toBeInTheDocument();
  });

  it('calls onDeploy when deploy button is clicked', () => {
    const handleDeploy = vi.fn();
    render(<StackList stacks={mockStacks} onDelete={() => {}} onDeploy={handleDeploy} />);

    const deployButtons = screen.getAllByText('Deploy');
    fireEvent.click(deployButtons[0]);

    expect(handleDeploy).toHaveBeenCalledWith('stack-1');
  });

  it('calls onDelete when delete button is clicked', () => {
    const handleDelete = vi.fn();
    render(<StackList stacks={mockStacks} onDelete={handleDelete} onDeploy={() => {}} />);

    const deleteButtons = screen.getAllByLabelText('Delete');
    fireEvent.click(deleteButtons[0]);

    expect(handleDelete).toHaveBeenCalledWith('stack-1');
  });

  it('renders empty state when no stacks provided', () => {
      render(<StackList stacks={[]} onDelete={() => {}} onDeploy={() => {}} />);
      expect(screen.getByText(/No stacks found/i)).toBeInTheDocument();
  });
});
