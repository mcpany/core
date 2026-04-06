/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { JsonView } from './json-view';

// Avoid lazy-loading issues in jsdom – replace the real syntax highlighter with
// a lightweight synchronous stub so tests don't hit Suspense boundaries.
vi.mock('./optimized-syntax-highlighter', () => ({
  default: ({ children }: { children: React.ReactNode }) => <pre data-testid="syntax-highlighter">{children}</pre>,
}));

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

    // Switch to JSON (Raw) - use act to handle the Suspense boundary of the
    // lazily-loaded SyntaxHighlighter component
    const jsonBtn = screen.getByText('Raw');
    await act(async () => {
      fireEvent.click(jsonBtn);
    });
    // ⚡ BOLT: Wait for lazy-loaded SyntaxHighlighter
    expect(await screen.findByText(/"Alice"/)).toBeInTheDocument();
  });

  it('collapses long content conditionally based on height', () => {
      // Create a mock ref to simulate a large scroll height
      const data = { key: 'very long content' };

      // Override scrollHeight property getter for HTMLDivElement just for this test
      const originalScrollHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'scrollHeight');
      Object.defineProperty(HTMLElement.prototype, 'scrollHeight', {
          configurable: true,
          get() { return 500; } // Mock value > maxHeight (100)
      });

      render(<JsonView data={data} maxHeight={100} />);

      expect(screen.getByText('Show More')).toBeInTheDocument();

      fireEvent.click(screen.getByText('Show More'));
      expect(screen.getByText('Show Less')).toBeInTheDocument();

      // Restore the property
      if (originalScrollHeight) {
          Object.defineProperty(HTMLElement.prototype, 'scrollHeight', originalScrollHeight);
      }
  });

  it('does not show collapse button for short content', () => {
      const data = { key: 'short' };

      const originalScrollHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'scrollHeight');
      Object.defineProperty(HTMLElement.prototype, 'scrollHeight', {
          configurable: true,
          get() { return 50; } // Mock value < maxHeight (100)
      });

      render(<JsonView data={data} maxHeight={100} />);

      expect(screen.queryByText('Show More')).not.toBeInTheDocument();

      if (originalScrollHeight) {
          Object.defineProperty(HTMLElement.prototype, 'scrollHeight', originalScrollHeight);
      }
  });
});
