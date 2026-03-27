/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { RichResultViewer } from './rich-result-viewer';

describe('RichResultViewer', () => {
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

  it('detects nested arrays for tables', () => {
    const { getByText } = render(
      <RichResultViewer
        result={{
          files: [
            { id: 1, name: 'foo' },
            { id: 2, name: 'bar' }
          ]
        }}
      />
    );
    expect(getByText('Table')).toBeInTheDocument();
  });
});
