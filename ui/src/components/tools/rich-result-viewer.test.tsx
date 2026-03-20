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

  it('renders flattened nested JSON object as a table', () => {
    render(
      <RichResultViewer
        result={[
          {
            id: 1,
            user: {
              profile: {
                name: 'Alice',
                age: 30,
              },
            },
          },
        ]}
      />,
    );

    // In a flat table, "user.profile.name" and "user.profile.age" should be the column headers
    expect(screen.getByText('user.profile.name')).toBeInTheDocument();
    expect(screen.getByText('user.profile.age')).toBeInTheDocument();
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('30')).toBeInTheDocument();
  });
});
