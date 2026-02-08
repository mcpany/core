/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { SmartResultRenderer } from './smart-result-renderer';
import React from 'react';

// Mock JsonView because it's lazy-loaded or complex
vi.mock('@/components/ui/json-view', () => ({
  JsonView: ({ data }: { data: any }) => (
    <div data-testid="json-view">{JSON.stringify(data)}</div>
  ),
}));

describe('SmartResultRenderer', () => {
  it('renders raw JSON when data is simple object', () => {
    const data = { foo: 'bar' };
    render(<SmartResultRenderer result={data} />);
    expect(screen.getByTestId('json-view')).toHaveTextContent('{"foo":"bar"}');
  });

  it('renders table when data is array of objects', () => {
    const data = [
      { id: 1, name: 'Alice' },
      { id: 2, name: 'Bob' },
    ];
    render(<SmartResultRenderer result={data} />);
    // Check for table headers
    expect(screen.getByText('id')).toBeInTheDocument();
    expect(screen.getByText('name')).toBeInTheDocument();
    // Check for table content
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
  });

  it('unwraps text content from CallToolResult and renders table', () => {
    const data = {
      content: [
        {
          type: 'text',
          text: '[{"id": 1, "name": "Charlie"}]',
        },
      ],
    };
    render(<SmartResultRenderer result={data} />);
    expect(screen.getByText('Charlie')).toBeInTheDocument();
  });

  it('unwraps stdout from Command output and renders table', () => {
    const data = {
      stdout: '[{"id": 1, "name": "Dave"}]',
    };
    render(<SmartResultRenderer result={data} />);
    expect(screen.getByText('Dave')).toBeInTheDocument();
  });

  // New features tests (expected to fail initially or need implementation)
  it('renders image when content contains image type', () => {
    const data = {
      content: [
        {
          type: 'image',
          data: 'base64data',
          mimeType: 'image/png',
        },
      ],
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
          text: 'Here is an image:',
        },
        {
          type: 'image',
          data: 'base64data',
          mimeType: 'image/jpeg',
        },
      ],
    };
    render(<SmartResultRenderer result={data} />);
    expect(screen.getByText('Here is an image:')).toBeInTheDocument();
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'data:image/jpeg;base64,base64data');
  });

  it('renders image from nested command stdout JSON', () => {
     // This simulates a command line tool returning a JSON string that IS a CallToolResult content array
     const innerJson = JSON.stringify([
         {
             type: 'image',
             data: 'base64data',
             mimeType: 'image/png'
         }
     ]);
     const data = {
         stdout: innerJson
     };

     render(<SmartResultRenderer result={data} />);
     const img = screen.getByRole('img');
     expect(img).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });
});
