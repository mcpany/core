/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { SmartResultRenderer } from './smart-result-renderer';
import { describe, it, expect, vi } from 'vitest';

// Mock JsonView because it might use canvas or complex DOM
vi.mock('@/components/ui/json-view', () => ({
  JsonView: ({ data }: { data: any }) => <div data-testid="json-view">{JSON.stringify(data)}</div>,
}));

describe('SmartResultRenderer', () => {
  it('renders raw JSON for simple text content', () => {
    const result = { foo: 'bar' };
    render(<SmartResultRenderer result={result} />);
    expect(screen.getByTestId('json-view')).toHaveTextContent('{"foo":"bar"}');
  });

  it('renders table for array of objects', () => {
    const result = [
      { id: 1, name: 'Alice' },
      { id: 2, name: 'Bob' },
    ];
    render(<SmartResultRenderer result={result} />);
    // Check for table elements
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
    // It should NOT render JsonView if table mode is active (default)
    expect(screen.queryByTestId('json-view')).not.toBeInTheDocument();
  });

  it('renders image for MCP image content', () => {
    const result = {
      content: [
        {
          type: 'image',
          data: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=',
          mimeType: 'image/png',
        },
      ],
    };
    render(<SmartResultRenderer result={result} />);

    // Look for image tag
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=');
  });

  it('renders mixed content (text + image)', () => {
    const result = {
      content: [
        { type: 'text', text: 'Here is an image:' },
        {
          type: 'image',
          data: 'base64data',
          mimeType: 'image/png',
        },
      ],
    };
    render(<SmartResultRenderer result={result} />);

    expect(screen.getByText('Here is an image:')).toBeInTheDocument();
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });

  it('renders image from nested JSON in stdout (Command Output)', () => {
    const nestedContent = JSON.stringify([
      {
        type: 'image',
        data: 'base64data',
        mimeType: 'image/png',
      },
    ]);

    const result = {
      command: 'echo ...',
      stdout: nestedContent,
    };

    render(<SmartResultRenderer result={result} />);

    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });
});
