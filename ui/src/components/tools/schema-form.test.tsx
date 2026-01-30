/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SchemaForm, Schema } from './schema-form';
import { TooltipProvider } from '@/components/ui/tooltip';

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <TooltipProvider>
      {ui}
    </TooltipProvider>
  );
};

describe('SchemaForm', () => {
  it('renders a string input', () => {
    const schema: Schema = { type: 'string', title: 'Name' };
    const onChange = vi.fn();
    renderWithProviders(<SchemaForm schema={schema} value="" onChange={onChange} name="username" />);

    // With title='Name', description is 'Name'. Placeholder becomes "Enter username..."
    const input = screen.getByPlaceholderText('Enter username...');
    expect(input).toBeInTheDocument();

    fireEvent.change(input, { target: { value: 'test' } });
    expect(onChange).toHaveBeenCalledWith('test');
  });

  it('renders a number input', () => {
    const schema: Schema = { type: 'number', title: 'Age' };
    const onChange = vi.fn();
    renderWithProviders(<SchemaForm schema={schema} value={10} onChange={onChange} name="age" />);

    const input = screen.getByRole('spinbutton'); // Input type=number renders as spinbutton
    expect(input).toBeInTheDocument();
    expect(input).toHaveValue(10);

    fireEvent.change(input, { target: { value: '20' } });
    expect(onChange).toHaveBeenCalledWith(20);
  });

  it('renders a boolean switch', () => {
    const schema: Schema = { type: 'boolean', title: 'Active' };
    const onChange = vi.fn();
    renderWithProviders(<SchemaForm schema={schema} value={false} onChange={onChange} name="active" />);

    const switchEl = screen.getByRole('switch');
    expect(switchEl).toBeInTheDocument();

    fireEvent.click(switchEl);
    expect(onChange).toHaveBeenCalledWith(true);
  });

  it('renders an enum select', () => {
    const schema: Schema = { enum: ['A', 'B'], title: 'Type' };
    const onChange = vi.fn();
    renderWithProviders(<SchemaForm schema={schema} value="A" onChange={onChange} name="type" />);

    const trigger = screen.getByRole('combobox');
    expect(trigger).toBeInTheDocument();
    expect(trigger).toHaveTextContent('A');
  });

  it('renders nested object', () => {
    const schema: Schema = {
      type: 'object',
      properties: {
        field1: { type: 'string', title: 'Field 1' } // Added title so placeholder is generated
      }
    };
    const onChange = vi.fn();
    const value = { field1: 'val' };

    renderWithProviders(<SchemaForm schema={schema} value={value} onChange={onChange} root={true} />);

    const input = screen.getByPlaceholderText('Enter field1...');
    expect(input).toHaveValue('val');

    fireEvent.change(input, { target: { value: 'new' } });
    expect(onChange).toHaveBeenCalledWith({ field1: 'new' });
  });

  it('renders array', () => {
      const schema: Schema = {
          type: 'array',
          items: { type: 'string', title: 'Item' }
      };
      const onChange = vi.fn();
      const value = ['item1'];

      renderWithProviders(<SchemaForm schema={schema} value={value} onChange={onChange} name="list" />);

      const input = screen.getByDisplayValue('item1');
      expect(input).toBeInTheDocument();

      const addButton = screen.getByText('Add Item');
      fireEvent.click(addButton);
      // Expect onChange to be called with new array
      expect(onChange).toHaveBeenCalledWith(['item1', '']);
  });
});
