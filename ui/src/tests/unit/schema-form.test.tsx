/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from '@testing-library/react';
import { SchemaForm } from '@/components/tools/schema-form';
import { describe, it, expect, vi } from 'vitest';

// Mock Tooltip components to avoid issues with Radix UI in test environment if needed
// But Radix usually works fine in JSDOM. If it fails, I'll mock it.

describe('SchemaForm', () => {
  it('renders a text input for string type', () => {
    const schema = { type: 'string' };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value="" onChange={onChange} name="test_field" />);

    const input = screen.getByRole('textbox');
    expect(input).toBeDefined();

    fireEvent.change(input, { target: { value: 'hello' } });
    expect(onChange).toHaveBeenCalledWith('hello');
  });

  it('renders a number input for number type', () => {
    const schema = { type: 'number' };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value={10} onChange={onChange} name="age" />);

    // type="number" input is a spinbutton
    const input = screen.getByRole('spinbutton');
    expect(input).toBeDefined();

    fireEvent.change(input, { target: { value: '20' } });
    expect(onChange).toHaveBeenCalledWith(20);
  });

  it('renders a switch for boolean type', () => {
    const schema = { type: 'boolean' };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value={false} onChange={onChange} name="is_active" />);

    const switchBtn = screen.getByRole('switch');
    expect(switchBtn).toBeDefined();

    fireEvent.click(switchBtn);
    expect(onChange).toHaveBeenCalledWith(true);
  });

  it('renders a select for enum', () => {
    // Radix UI Select is tricky to test because it uses portals.
    // However, we can test that the trigger renders.
    const schema = { type: 'string', enum: ['A', 'B'] };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value="" onChange={onChange} name="category" />);

    const trigger = screen.getByRole('combobox');
    expect(trigger).toBeDefined();
    // Interacting with Radix Select requires more setup (pointer events),
    // verifying existence is enough for unit test of SchemaForm logic mapping.
  });

  it('renders nested object fields', () => {
    const schema = {
      type: 'object',
      properties: {
        nested_field: { type: 'string' }
      }
    };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value={{}} onChange={onChange} name="root" />);

    // Label for nested field should be present
    expect(screen.getByText('nested_field')).toBeDefined();
    const input = screen.getByRole('textbox');
    fireEvent.change(input, { target: { value: 'nested_val' } });

    expect(onChange).toHaveBeenCalledWith({ nested_field: 'nested_val' });
  });

  it('renders array with add button', () => {
      const schema = {
          type: 'array',
          items: { type: 'string' }
      };
      const onChange = vi.fn();
      render(<SchemaForm schema={schema} value={[]} onChange={onChange} name="tags" />);

      expect(screen.getByText('tags')).toBeDefined();
      // Initially empty
      expect(screen.queryByRole('textbox')).toBeNull();

      // Find Add button (it has a Plus icon, usually accessible name might be missing or tricky,
      // but in our code: <span className="font-medium text-sm">{name || "List"}</span>... <Button ...><Plus/></Button>
      // We didn't give aria-label to the button explicitly in SchemaForm.
      // But we can find by role button.
      const buttons = screen.getAllByRole('button');
      // The Add button is likely the first one or we can search by class or nearby text.
      // But simpler: just click the one that looks like add?
      // In this isolated render, there's only the Add button.
      // Wait, there might be Tooltip triggers? No, only description has tooltip.
      // We have no description in schema.
      const addBtn = buttons[0];

      fireEvent.click(addBtn);
      // Should call onChange with one undefined item
      expect(onChange).toHaveBeenCalledWith([undefined]);
  });
});
