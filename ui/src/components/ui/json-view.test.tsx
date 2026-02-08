/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { JsonView } from './json-view';

// Mock clipboard
const mockWriteText = vi.fn().mockResolvedValue(undefined);
Object.assign(navigator, {
  clipboard: {
    writeText: mockWriteText,
  },
});

// Mock ResizeObserver which might be used by SyntaxHighlighter or something internal
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

describe('JsonView', () => {
  it('renders JSON string correctly', () => {
    const data = { key: 'value' };
    render(<JsonView data={data} />);
    // SyntaxHighlighter might break it up into spans, so we search for text parts
    expect(screen.getByText(/"key"/)).toBeInTheDocument();
    expect(screen.getByText(/"value"/)).toBeInTheDocument();
  });

  it('renders null correctly', () => {
    render(<JsonView data={null} />);
    expect(screen.getByText('null')).toBeInTheDocument();
  });

  it('copies to clipboard', () => {
    const data = { foo: 'bar' };
    render(<JsonView data={data} />);

    // The copy button is initially hidden (opacity 0) but present in DOM
    const copyButton = screen.getByTitle('Copy JSON');
    fireEvent.click(copyButton);

    expect(mockWriteText).toHaveBeenCalledWith(JSON.stringify(data, null, 2));
  });

  it('supports smart table view', async () => {
    const data = [
        { id: 1, name: 'Alice' },
        { id: 2, name: 'Bob' }
    ];
    render(<JsonView data={data} smartTable={true} />);

    // Should render table button
    expect(screen.getByText('Table')).toBeInTheDocument();

    // Default mode is Smart (Table)
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();

    // Switch to JSON (Raw)
    const jsonBtn = screen.getByText('Raw');
    fireEvent.click(jsonBtn);
    // ⚡ BOLT: Wait for lazy-loaded SyntaxHighlighter
    expect(await screen.findByText(/"Alice"/)).toBeInTheDocument();
  });

  it('collapses long content', () => {
      // We can't easily test visual height in jsdom, but we can check if the collapse button renders
      // and toggles state.
      const data = { key: 'very long content' };
      // maxHeight defaults to 400.

      render(<JsonView data={data} maxHeight={100} />);

      // The button "Show More" should be present if we force it?
      // Wait, render logic says:
      // const showCollapse = maxHeight > 0;
      // ... {showCollapse && ( ... button ... )}

      // So the button is ALWAYS rendered if maxHeight > 0?
      // Yes, my implementation:
      /*
        {showCollapse && (
            <div className="...">
                <Button ...>
                    {isExpanded ? ... : ...}
                </Button>
            </div>
        )}
      */
      // Wait, checking my implementation:
      // It renders the button unconditionally if showCollapse is true?
      // Yes. It doesn't check if the content *actually* exceeds maxHeight.
      // This is a known limitation I accepted in comments:
      // "Calculate approximate lines to guess if we need expand button without rendering? Hard to do accurately."

      expect(screen.getByText('Show More')).toBeInTheDocument();

      fireEvent.click(screen.getByText('Show More'));
      expect(screen.getByText('Show Less')).toBeInTheDocument();
  });
});
