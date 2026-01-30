/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SchemaForm } from '../../components/tools/schema-form';

// Mock UI components that might be problematic in JSDOM if not fully mocked,
// but basic inputs should be fine. Radix UI usually works with setup.

describe('SchemaForm', () => {
  it('renders string input', () => {
    const onChange = vi.fn();
    const schema = { type: 'string', description: 'Enter name' };
    render(<SchemaForm schema={schema} value="test" onChange={onChange} name="Name" />);

    expect(screen.getByLabelText('Name')).toBeDefined();
    expect(screen.getByPlaceholderText('Enter name')).toBeDefined();

    const input = screen.getByDisplayValue('test');
    fireEvent.change(input, { target: { value: 'new value' } });

    expect(onChange).toHaveBeenCalledWith('new value');
  });

  it('renders boolean switch', () => {
    const onChange = vi.fn();
    const schema = { type: 'boolean' };
    render(<SchemaForm schema={schema} value={false} onChange={onChange} name="Enabled" />);

    expect(screen.getByText('Enabled')).toBeDefined();
    const switchEl = screen.getByRole('switch');
    fireEvent.click(switchEl);

    expect(onChange).toHaveBeenCalledWith(true);
  });

  it('renders nested object', () => {
     const onChange = vi.fn();
     const schema = {
         type: 'object',
         properties: {
             child: { type: 'string' }
         }
     };
     const value = { child: 'val' };
     render(<SchemaForm schema={schema} value={value} onChange={onChange} name="Parent" />);

     expect(screen.getByText('Parent')).toBeDefined();
     expect(screen.getByDisplayValue('val')).toBeDefined();
  });
});
