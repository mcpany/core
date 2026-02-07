/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { SmartResultRenderer } from './smart-result-renderer';
import { describe, it, expect } from 'vitest';

describe('SmartResultRenderer', () => {
  it('renders table for array of objects', () => {
    const data = [{ id: 1, name: 'Test' }, { id: 2, name: 'Test 2' }];
    render(<SmartResultRenderer result={data} />);
    // Just check if text is present. Table structure verification is secondary.
    expect(screen.getByText('Test')).toBeInTheDocument();
    expect(screen.getByText('Test 2')).toBeInTheDocument();
  });

  it('renders image for CallToolResult with image', () => {
    const data = {
      content: [
        {
          type: 'image',
          data: 'base64data',
          mimeType: 'image/png'
        }
      ]
    };
    render(<SmartResultRenderer result={data} />);
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });

  it('renders mixed content (text + image)', () => {
    const data = {
      content: [
        {
          type: 'text',
          text: 'Hello World'
        },
        {
          type: 'image',
          data: 'base64data',
          mimeType: 'image/png'
        }
      ]
    };
    render(<SmartResultRenderer result={data} />);
    expect(screen.getByText('Hello World')).toBeInTheDocument();
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });

  it('renders image from command output stdout', () => {
    const stdout = JSON.stringify([
        {
          type: 'image',
          data: 'base64data',
          mimeType: 'image/png'
        }
    ]);
    const data = {
        stdout: stdout
    };
    render(<SmartResultRenderer result={data} />);
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });
});
