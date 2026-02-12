/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { CollectionsSidebar, Collection } from '@/components/playground/collections-sidebar';
import { describe, it, expect, vi } from 'vitest';

describe('CollectionsSidebar', () => {
  const mockCollections: Collection[] = [
    {
      id: '1',
      name: 'Test Collection',
      requests: [
        { id: 'r1', name: 'Req 1', toolName: 'test_tool', toolArgs: { foo: 'bar' } }
      ]
    }
  ];

  const mockSetCollections = vi.fn();
  const mockOnRunRequest = vi.fn();

  it('renders collections', () => {
    render(
      <CollectionsSidebar
        collections={mockCollections}
        setCollections={mockSetCollections}
        onRunRequest={mockOnRunRequest}
      />
    );

    expect(screen.getByText('Test Collection')).toBeDefined();
  });

  it('expands collection and runs request', () => {
    render(
      <CollectionsSidebar
        collections={mockCollections}
        setCollections={mockSetCollections}
        onRunRequest={mockOnRunRequest}
      />
    );

    // Expand
    fireEvent.click(screen.getByText('Test Collection'));

    // Check request visible
    expect(screen.getByText('Req 1')).toBeDefined();

    // Click request to run
    fireEvent.click(screen.getByText('Req 1'));

    expect(mockOnRunRequest).toHaveBeenCalledWith('test_tool', { foo: 'bar' });
  });

  it('creates new collection', () => {
    render(
      <CollectionsSidebar
        collections={[]}
        setCollections={mockSetCollections}
        onRunRequest={mockOnRunRequest}
      />
    );

    // Click add (the first button is usually the Add button in the header)
    // Better: Find by Icon if possible or just use indices.
    // There are only a few buttons. The Plus icon button.
    const buttons = screen.getAllByRole('button');
    const addBtn = buttons[0]; // Assuming order

    fireEvent.click(addBtn);

    // Input should appear
    const input = screen.getByPlaceholderText('Name...');
    fireEvent.change(input, { target: { value: 'New Col' } });
    fireEvent.keyDown(input, { key: 'Enter', code: 'Enter' });

    // Expect setCollections to be called
    expect(mockSetCollections).toHaveBeenCalled();
    // Verify the argument contains the new collection
    const callArg = mockSetCollections.mock.calls[0][0];
    // Since we spread ...collections (which is empty), result is [newCol]
    expect(callArg).toHaveLength(1);
    expect(callArg[0].name).toBe('New Col');
  });
});
