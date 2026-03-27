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
  it('renders JSON string correctly', async () => {
    const data = { key: 'value' };
    render(<JsonView data={data} />);

    // Switch to Raw view to test raw syntax rendering
    const jsonBtn = screen.getByText('Raw');
    await act(async () => {
      fireEvent.click(jsonBtn);
    });

    // SyntaxHighlighter might break it up into spans, so we search for text parts
    expect(await screen.findByText(/"key"/)).toBeInTheDocument();
    expect(screen.getByText(/"value"/)).toBeInTheDocument();
  });

  it('renders null correctly', () => {
    render(<JsonView data={null} />);
    expect(screen.getByText('null')).toBeInTheDocument();
  });

  it('copies to clipboard', async () => {
    const data = { foo: 'bar' };
    render(<JsonView data={data} />);

    // Switch to Raw view which has the copy button for the entire JSON payload
    const jsonBtn = screen.getByText('Raw');
    await act(async () => {
      fireEvent.click(jsonBtn);
    });

    // The copy button is initially hidden (opacity 0) but present in DOM
    const copyButton = await screen.findByTitle('Copy JSON');
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

  it('collapses long content', async () => {
      // We can't easily test visual height in jsdom, but we can check if the collapse button renders
      // and toggles state.
      const data = { key: 'very long content' };
      // maxHeight defaults to 400.

      render(<JsonView data={data} maxHeight={100} />);

      // The properties view doesn't have a collapse button by default because it's a flat table.
      // We need to switch to "Raw" view first to test the collapse logic.
      const jsonBtn = screen.getByText('Raw');
      await act(async () => {
        fireEvent.click(jsonBtn);
      });

      // Now "Show More" should be present.
      expect(await screen.findByText('Show More')).toBeInTheDocument();

      fireEvent.click(screen.getByText('Show More'));
      expect(screen.getByText('Show Less')).toBeInTheDocument();
  });
});
