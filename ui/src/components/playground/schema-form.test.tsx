/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SchemaForm } from './schema-form';
import { vi } from 'vitest';

describe('SchemaForm', () => {
  it('renders a file input for base64 encoded strings and handles file upload', async () => {
    const schema = {
      type: 'object',
      properties: {
        fileData: {
          type: 'string',
          contentEncoding: 'base64',
        },
      },
    };

    const handleChange = vi.fn();

    render(<SchemaForm schema={schema} value={{}} onChange={handleChange} />);

    const fileInput = screen.getByLabelText(/fileData/i);
    expect(fileInput).toHaveAttribute('type', 'file');

    const file = new File(['hello world'], 'hello.txt', { type: 'text/plain' });

    // Simulate file upload
    await userEvent.upload(fileInput, file);

    await waitFor(() => {
      // "hello world" in base64 is aGVsbG8gd29ybGQ=
      expect(handleChange).toHaveBeenCalledWith(
        expect.objectContaining({
          fileData: 'aGVsbG8gd29ybGQ=',
        })
      );
    });
  });

  it('renders a file input for binary formatted strings', async () => {
    const schema = {
      type: 'object',
      properties: {
        binaryData: {
          type: 'string',
          format: 'binary',
        },
      },
    };

    const handleChange = vi.fn();

    render(<SchemaForm schema={schema} value={{}} onChange={handleChange} />);

    const fileInput = screen.getByLabelText(/binaryData/i);
    expect(fileInput).toHaveAttribute('type', 'file');
  });
});
