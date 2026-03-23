/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { RichResultViewer } from './rich-result-viewer';

describe('RichResultViewer', () => {
  it('renders a single JSON object as a Key-Value table', () => {
    render(
      <RichResultViewer
        result={{
          name: 'test-tool',
          status: 'ok'
        }}
      />,
    );

    // "Table" tab should be present
    expect(screen.getByRole('tab', { name: /Table/i })).toBeInTheDocument();

    // Key and Value headers should be present
    expect(screen.getByText('Key')).toBeInTheDocument();
    expect(screen.getByText('Value')).toBeInTheDocument();

    // Data should be present
    expect(screen.getByText('name')).toBeInTheDocument();
    expect(screen.getByText('test-tool')).toBeInTheDocument();
    expect(screen.getByText('status')).toBeInTheDocument();
    expect(screen.getByText('ok')).toBeInTheDocument();
  });

  it('renders stdout JSON arrays containing image content', () => {
    render(
      <RichResultViewer
        result={{
          stdout: JSON.stringify([
            {
              type: 'image',
              data: 'base64data',
              mimeType: 'image/png',
            },
          ]),
        }}
      />,
    );

    expect(screen.getByRole('img')).toHaveAttribute('src', 'data:image/png;base64,base64data');
  });
});
