/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { SchemaForm } from '@/components/tools/schema-form';

// Mock components that might be problematic in JSDOM or just to simplify testing
vi.mock('@/components/ui/select', () => ({
  Select: ({ children, onValueChange }: any) => <div onClick={() => onValueChange('option1')}>{children}</div>,
  SelectTrigger: ({ children }: any) => <button>{children}</button>,
  SelectValue: ({ placeholder }: any) => <span>{placeholder}</span>,
  SelectContent: ({ children }: any) => <div>{children}</div>,
  SelectItem: ({ children, value }: any) => <div data-value={value}>{children}</div>,
}));

vi.mock('@/components/ui/switch', () => ({
  Switch: ({ checked, onCheckedChange }: any) => (
    <button role="switch" aria-checked={checked} onClick={() => onCheckedChange(!checked)} />
  ),
}));

describe('SchemaForm', () => {
  it('renders nothing if schema is missing', () => {
    const { container } = render(<SchemaForm schema={undefined as any} value={{}} onChange={() => {}} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('renders a text input for string type', () => {
    const schema = { type: 'string' };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value="" onChange={onChange} name="testField" />);

    expect(screen.getByText('testField')).toBeInTheDocument();
    const input = screen.getByRole('textbox');
    expect(input).toBeInTheDocument();

    fireEvent.change(input, { target: { value: 'hello' } });
    expect(onChange).toHaveBeenCalledWith('hello');
  });

  it('renders a number input for number type', () => {
    const schema = { type: 'number' };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value={0} onChange={onChange} name="age" />);

    expect(screen.getByText('age')).toBeInTheDocument();
    const input = screen.getByRole('spinbutton'); // input type="number" has role spinbutton
    expect(input).toBeInTheDocument();

    fireEvent.change(input, { target: { value: '42' } });
    expect(onChange).toHaveBeenCalledWith(42);
  });

  it('renders a switch for boolean type', () => {
    const schema = { type: 'boolean' };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value={false} onChange={onChange} name="active" />);

    expect(screen.getByText('active')).toBeInTheDocument();
    const switchBtn = screen.getByRole('switch');
    expect(switchBtn).toBeInTheDocument();
    expect(switchBtn).toHaveAttribute('aria-checked', 'false');

    fireEvent.click(switchBtn);
    expect(onChange).toHaveBeenCalledWith(true);
  });

  it('renders nested object fields', () => {
    const schema = {
      type: 'object',
      properties: {
        nestedField: { type: 'string' }
      }
    };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value={{}} onChange={onChange} name="parent" />);

    expect(screen.getByText('nestedField')).toBeInTheDocument();
  });
});
