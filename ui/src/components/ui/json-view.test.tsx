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

  it('collapses long content', () => {
      const data = { key: 'very long content' };
      render(<JsonView data={data} maxHeight={100} />);
      expect(screen.getByText('Show More')).toBeInTheDocument();
      fireEvent.click(screen.getByText('Show More'));
      expect(screen.getByText('Show Less')).toBeInTheDocument();
  });
});
