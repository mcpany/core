/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { JsonTree } from './json-tree';

// Mock clipboard
const mockWriteText = vi.fn().mockResolvedValue(undefined);
Object.assign(navigator, {
  clipboard: {
    writeText: mockWriteText,
  },
});

describe('JsonTree', () => {
  it('renders primitive string correctly', () => {
    render(<JsonTree data="test string" />);
    expect(screen.getByText(/"test string"/)).toBeInTheDocument();
  });

  it('renders primitive number correctly', () => {
    render(<JsonTree data={123} />);
    expect(screen.getByText('123')).toBeInTheDocument();
  });

  it('renders null correctly', () => {
    render(<JsonTree data={null} />);
    expect(screen.getByText('null')).toBeInTheDocument();
  });

  it('renders object keys and values', () => {
    const data = { foo: 'bar', num: 42 };
    render(<JsonTree data={data} defaultExpandedLevel={1} />);
    expect(screen.getByText(/"foo":/)).toBeInTheDocument();
    expect(screen.getByText(/"bar"/)).toBeInTheDocument();
    expect(screen.getByText(/"num":/)).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
  });

  it('renders array elements', () => {
    const data = ['a', 'b'];
    render(<JsonTree data={data} defaultExpandedLevel={1} />);
    // Arrays don't show keys by default in my implementation (just values in order)
    expect(screen.getByText(/"a"/)).toBeInTheDocument();
    expect(screen.getByText(/"b"/)).toBeInTheDocument();
  });

  it('collapses and expands objects', () => {
    const data = { nested: { key: 'value' } };
    // Start collapsed (defaultExpandedLevel=0)
    render(<JsonTree data={data} defaultExpandedLevel={0} />);

    // Should show root object structure collapsed
    // The preview text containing "nested" should be visible
    expect(screen.getByText(/nested/)).toBeInTheDocument();

    // Value "value" should NOT be visible yet (it's inside nested object)
    expect(screen.queryByText(/"value"/)).not.toBeInTheDocument();

    // Click to expand root
    // Find the clickable header by finding the opening brace
    const expander = screen.getByText('{');
    fireEvent.click(expander);

    // Now root is expanded.
    // The nested object key "nested" is visible.
    // But nested object itself is collapsed (level 1 >= 0 defaultExpandedLevel) if logic matches.
    // Wait, level passed to child is level+1 (so 1). defaultExpandedLevel is 0.
    // 1 < 0 is false. So child is collapsed.

    // We should see the nested object key
    expect(screen.getByText(/"nested":/)).toBeInTheDocument();

    // But the inner value "value" should still be hidden because nested object is collapsed
    expect(screen.queryByText(/"value"/)).not.toBeInTheDocument();

    // Expand nested object. It has its own header.
    // There are multiple "{" now (root and nested).
    // The nested one is inside the tree.
    const braces = screen.getAllByText('{');
    // braces[0] is root (already expanded). braces[1] is nested (collapsed).
    fireEvent.click(braces[1]);

    // Now "value" should be visible
    expect(screen.getByText(/"value"/)).toBeInTheDocument();
  });

  it('handles defaultExpandedLevel', () => {
      const data = { level1: { level2: 'val' } };
      render(<JsonTree data={data} defaultExpandedLevel={2} />);

      // Level 0 (root) expanded (0 < 2)
      // Level 1 (level1) expanded (1 < 2)
      // Level 2 (string) is primitive, always visible if parent expanded.

      expect(screen.getByText(/"val"/)).toBeInTheDocument();
  });

  it('copies primitive value on click', () => {
      render(<JsonTree data="copy me" />);
      const copyButton = screen.getByTitle('Copy value');
      fireEvent.click(copyButton);
      expect(mockWriteText).toHaveBeenCalledWith("copy me");
  });

  it('copies object json on click', () => {
      const data = { a: 1 };
      render(<JsonTree data={data} />);
      const copyButton = screen.getByTitle('Copy JSON');
      fireEvent.click(copyButton);
      expect(mockWriteText).toHaveBeenCalledWith(JSON.stringify(data, null, 2));
  });
});
