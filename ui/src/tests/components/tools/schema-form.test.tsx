/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { SchemaForm, Schema } from '@/components/tools/schema-form';
import { TooltipProvider } from "@/components/ui/tooltip";

// Mock Tooltip components to simplify testing
vi.mock("@/components/ui/tooltip", () => ({
  Tooltip: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  TooltipTrigger: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  TooltipContent: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  TooltipProvider: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

describe('SchemaForm', () => {
  const mockOnChange = vi.fn();

  const renderForm = (schema: Schema, value?: any) => {
    return render(
        <SchemaForm schema={schema} value={value} onChange={mockOnChange} />
    );
  };

  it('renders a text input for string type', () => {
    const schema: Schema = { type: 'string', title: 'Test String' };
    renderForm(schema, 'initial');

    const input = screen.getByLabelText('Test String');
    expect(input).toBeDefined();
    expect((input as HTMLInputElement).value).toBe('initial');

    fireEvent.change(input, { target: { value: 'updated' } });
    expect(mockOnChange).toHaveBeenCalledWith('updated');
  });

  it('renders a number input for number type', () => {
    const schema: Schema = { type: 'number', title: 'Test Number' };
    renderForm(schema, 42);

    const input = screen.getByLabelText('Test Number');
    expect(input).toBeDefined();
    expect((input as HTMLInputElement).type).toBe('number');
    expect((input as HTMLInputElement).value).toBe('42');

    fireEvent.change(input, { target: { value: '100' } });
    expect(mockOnChange).toHaveBeenCalledWith(100);
  });

  it('renders nested object fields', () => {
    const schema: Schema = {
      type: 'object',
      properties: {
        nested: { type: 'string', title: 'Nested Field' }
      }
    };
    renderForm(schema, { nested: 'val' });

    const input = screen.getByLabelText('Nested Field');
    expect(input).toBeDefined();
    expect((input as HTMLInputElement).value).toBe('val');
  });
});
