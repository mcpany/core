/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { SchemaForm } from './schema-form';
import { vi } from 'vitest';

describe('SchemaForm', () => {
    const mockSchema = {
        type: "object",
        properties: {
            name: { type: "string", title: "Name", description: "Enter name" },
            enabled: { type: "boolean", title: "Enabled" },
            category: { type: "string", enum: ["A", "B"], title: "Category" },
            token: { type: "string", title: "Secret Token" }
        },
        required: ["name"]
    };

    it('renders all fields', () => {
        const onChange = vi.fn();
        render(<SchemaForm schema={mockSchema} value={{}} onChange={onChange} />);

        expect(screen.getByLabelText(/Name/)).toBeInTheDocument();
        expect(screen.getByLabelText(/Enabled/)).toBeInTheDocument();
        expect(screen.getByText(/Enter name/)).toBeInTheDocument(); // Description
    });

    it('handles text input changes', () => {
        const onChange = vi.fn();
        render(<SchemaForm schema={mockSchema} value={{}} onChange={onChange} />);

        const input = screen.getByLabelText(/Name/);
        fireEvent.change(input, { target: { value: 'Test Name' } });

        expect(onChange).toHaveBeenCalledWith({ name: 'Test Name' });
    });

    it('handles boolean input changes', () => {
        const onChange = vi.fn();
        render(<SchemaForm schema={mockSchema} value={{}} onChange={onChange} />);

        const checkbox = screen.getByLabelText(/Enabled/);
        fireEvent.click(checkbox);

        expect(onChange).toHaveBeenCalledWith({ enabled: 'true' });
    });

    it('renders password field for sensitive keys', () => {
        const onChange = vi.fn();
        render(<SchemaForm schema={mockSchema} value={{}} onChange={onChange} />);

        const input = screen.getByLabelText(/Secret Token/);
        expect(input).toHaveAttribute('type', 'password');
    });

    it('shows required indicator', () => {
        const onChange = vi.fn();
        render(<SchemaForm schema={mockSchema} value={{}} onChange={onChange} />);

        // The asterisk is in a span, getting by label text might include it or not depending on impl
        // Label text is "Name *" usually in standard screen readers if aria-hidden is not used
        // But our component: <Label> {title} {isRequired && span} </Label>
        // testing-library handles this well usually.
        const label = screen.getByText(/Name/);
        expect(label).toContainHTML('<span class="text-destructive">*</span>');
    });
});
