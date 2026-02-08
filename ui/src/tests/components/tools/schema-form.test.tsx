/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { SchemaForm } from '@/components/tools/schema-form';
import { TooltipProvider } from "@/components/ui/tooltip";

describe('SchemaForm', () => {
  it('renders a text input for string type', () => {
    const handleChange = vi.fn();
    const schema = { type: 'string' };
    render(
      <TooltipProvider>
        <SchemaForm schema={schema} value="" onChange={handleChange} />
      </TooltipProvider>
    );
    const input = screen.getByRole('textbox');
    expect(input).toBeDefined();

    fireEvent.change(input, { target: { value: 'hello' } });
    expect(handleChange).toHaveBeenCalledWith('hello');
  });

  it('renders a number input for number type', () => {
    const handleChange = vi.fn();
    const schema = { type: 'number' };
    render(
      <TooltipProvider>
        <SchemaForm schema={schema} value={0} onChange={handleChange} />
      </TooltipProvider>
    );
    const input = screen.getByRole('spinbutton');
    expect(input).toBeDefined();

    fireEvent.change(input, { target: { value: '42' } });
    expect(handleChange).toHaveBeenCalledWith(42);
  });

  it('renders nested object fields', () => {
    const handleChange = vi.fn();
    const schema = {
        type: 'object',
        properties: {
            name: { type: 'string' },
            age: { type: 'number' }
        }
    };
    const value = { name: 'Alice', age: 30 };
    render(
      <TooltipProvider>
        <SchemaForm schema={schema as any} value={value} onChange={handleChange} />
      </TooltipProvider>
    );

    expect(screen.getByDisplayValue('Alice')).toBeDefined();
    expect(screen.getByDisplayValue('30')).toBeDefined();
  });
});
