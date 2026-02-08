/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { SmartResultRenderer } from './smart-result-renderer';
import { describe, it, expect } from 'vitest';
import React from 'react';

describe('SmartResultRenderer', () => {
  it('renders text content as table if parsed as JSON array', () => {
    const result = {
      content: [
        {
          type: 'text',
          text: JSON.stringify([{ id: 1, name: 'Test' }])
        }
      ]
    };
    render(<SmartResultRenderer result={result} />);
    // Should verify table headers or cells
    expect(screen.getByText('id')).toBeInTheDocument();
    expect(screen.getByText('name')).toBeInTheDocument();
    expect(screen.getByText('Test')).toBeInTheDocument();
  });

  it('renders image content as img tag', () => {
    const result = {
      content: [
        {
          type: 'image',
          data: 'base64data',
          mimeType: 'image/png'
        }
      ]
    };
    render(<SmartResultRenderer result={result} />);
    const img = screen.getByRole('img');
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });

  it('renders mixed content', () => {
    const result = {
      content: [
        {
          type: 'text',
          text: 'Some text explanation'
        },
        {
          type: 'image',
          data: 'base64data',
          mimeType: 'image/png'
        }
      ]
    };
    render(<SmartResultRenderer result={result} />);
    expect(screen.getByText('Some text explanation')).toBeInTheDocument();
    const img = screen.getByRole('img');
    expect(img).toBeInTheDocument();
  });

  it('renders command output with nested JSON image content as img tag', () => {
    const nestedContent = JSON.stringify([
       {
          type: 'image',
          data: 'base64data',
          mimeType: 'image/png'
       }
    ]);

    const result = {
      command: 'test-cmd',
      stdout: nestedContent
    };

    render(<SmartResultRenderer result={result} />);
    const img = screen.getByRole('img');
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });
});
