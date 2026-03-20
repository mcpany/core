/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { RichResultViewer } from './rich-result-viewer';
import * as useToastHook from '@/hooks/use-toast';

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

  it('renders a table for tabular data and handles sorting and CSV export', () => {
    const mockToast = vi.fn();
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    vi.spyOn(useToastHook, 'useToast').mockReturnValue({ toast: mockToast } as any);

    // Mock URL.createObjectURL and revokeObjectURL
    const mockCreateObjectURL = vi.fn().mockReturnValue('blob:test-url');
    const mockRevokeObjectURL = vi.fn();
    global.URL.createObjectURL = mockCreateObjectURL;
    global.URL.revokeObjectURL = mockRevokeObjectURL;

    // Mock document.createElement to intercept the anchor click
    const originalCreateElement = document.createElement.bind(document);
    const mockClick = vi.fn();
    vi.spyOn(document, 'createElement').mockImplementation((tagName: string) => {
        if (tagName === 'a') {
            const el = originalCreateElement(tagName);
            el.click = mockClick;
            return el;
        }
        return originalCreateElement(tagName);
    });

    const tabularData = [
        { id: 2, name: 'Bob', active: false },
        { id: 1, name: 'Alice', active: true },
        { id: 3, name: 'Charlie', active: true }
    ];

    render(<RichResultViewer result={tabularData} />);

    // Verify it renders the table tab
    expect(screen.getByRole('tab', { name: /Table/i })).toBeInTheDocument();

    // Sort by ID
    const idHeader = screen.getByRole('columnheader', { name: /id/i });
    fireEvent.click(idHeader);

    // After ascending sort, Alice (id 1) should be the first row (after header)
    const cellsAfterSort = screen.getAllByRole('cell');
    expect(cellsAfterSort[0]).toHaveTextContent('1');

    // Click again for descending sort
    fireEvent.click(idHeader);
    const cellsAfterDescSort = screen.getAllByRole('cell');
    expect(cellsAfterDescSort[0]).toHaveTextContent('3');

    // Test CSV Export
    const exportButton = screen.getByRole('button', { name: /Export CSV/i });
    fireEvent.click(exportButton);

    expect(mockCreateObjectURL).toHaveBeenCalled();
    expect(mockClick).toHaveBeenCalled();
    expect(mockToast).toHaveBeenCalledWith({ title: "Exported to CSV", description: "Your result has been successfully exported." });
  });
});
