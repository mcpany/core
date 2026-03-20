/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { RichResultViewer } from './rich-result-viewer';

describe('RichResultViewer', () => {
  it('renders stdout JSON arrays containing image content', () => {
    render(
      <RichResultViewer
        result={{
          stdout: JSON.stringify([
            {
              type: 'image',
              data: 'base64data',
              mimeType: 'image/png',
            },
          ]),
        }}
      />,
    );

    expect(screen.getByRole('img')).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });

  it('renders tabular data, filters by search, and sorts by column', () => {
    const data = [
      { id: 1, name: 'Alice', role: 'Admin' },
      { id: 2, name: 'Bob', role: 'User' },
      { id: 3, name: 'Charlie', role: 'User' },
    ];

    render(<RichResultViewer result={data} />);

    // Switch to table tab
    fireEvent.click(screen.getByRole('tab', { name: /table/i }));

    // Verify all rows are visible
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
    expect(screen.getByText('Charlie')).toBeInTheDocument();

    // Type in search box
    const searchInput = screen.getByPlaceholderText('Search table...');
    fireEvent.change(searchInput, { target: { value: 'alice' } });

    // Verify filtered rows
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.queryByText('Bob')).not.toBeInTheDocument();
    expect(screen.queryByText('Charlie')).not.toBeInTheDocument();

    // Clear search
    fireEvent.change(searchInput, { target: { value: '' } });

    // Click on "name" header to sort descending (first click asc, second desc, but 'asc' is default, wait... default is no sort, first click sorts ascending)
    const nameHeader = screen.getByText('name');
    fireEvent.click(nameHeader); // asc
    fireEvent.click(nameHeader); // desc

    // In a descending sort, Charlie should come before Alice in the DOM.
    // We can verify this by checking the order of rows.
    const rows = screen.getAllByRole('row');
    // rows[0] is header
    // rows[1] should be Charlie
    expect(rows[1]).toHaveTextContent('Charlie');
    expect(rows[3]).toHaveTextContent('Alice');
  });
});
