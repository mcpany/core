/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { SmartResultRenderer } from './smart-result-renderer';
import { describe, it, expect } from 'vitest';

describe('SmartResultRenderer', () => {
  it('renders text content as table/JSON if parsed', () => {
    const result = {
      content: [{ type: 'text', text: '[{"id": 1, "name": "foo"}]' }]
    };
    render(<SmartResultRenderer result={result} />);
    // Should find table headers or cell content
    expect(screen.getByText('foo')).toBeInTheDocument();
  });

  it('renders image content as img tag', () => {
    const result = {
      content: [{
        type: 'image',
        data: 'base64data',
        mimeType: 'image/png'
      }]
    };
    render(<SmartResultRenderer result={result} />);
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });

  it('renders mixed content (text + image)', () => {
    const result = {
      content: [
        { type: 'text', text: 'Some explanation' },
        { type: 'image', data: 'base64data', mimeType: 'image/png' }
      ]
    };
    render(<SmartResultRenderer result={result} />);
    expect(screen.getByText('Some explanation')).toBeInTheDocument();
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });

  it('renders command output with nested JSON image content', () => {
    const nestedJson = JSON.stringify([
      { type: 'image', data: 'base64data', mimeType: 'image/png' }
    ]);
    const result = {
      stdout: nestedJson
    };
    render(<SmartResultRenderer result={result} />);
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });
});
