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

        expect(onChange).toHaveBeenCalledWith({ enabled: true });
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

        const label = screen.getByText(/Name/);
        expect(label).toContainHTML('<span class="text-destructive">*</span>');
    });

    describe("Nested Objects and Arrays", () => {
        const complexSchema = {
            type: "object",
            properties: {
                server: {
                    type: "object",
                    title: "Server Details",
                    properties: {
                        host: { type: "string", title: "Host" },
                        port: { type: "integer", title: "Port" }
                    }
                },
                features: {
                    type: "array",
                    title: "Features",
                    items: { type: "string", title: "Feature" }
                }
            }
        };

        it("renders nested object fields", () => {
            const onChange = vi.fn();
            render(<SchemaForm schema={complexSchema} value={{}} onChange={onChange} />);

            expect(screen.getByText("Server Details")).toBeInTheDocument();
            expect(screen.getByLabelText("Host")).toBeInTheDocument();
            expect(screen.getByLabelText("Port")).toBeInTheDocument();
        });

        it("renders array controls", () => {
            const onChange = vi.fn();
            render(<SchemaForm schema={complexSchema} value={{}} onChange={onChange} />);

            expect(screen.getByText("Add Feature")).toBeInTheDocument();
        });

        it("updates nested value", () => {
            const onChange = vi.fn();
            render(<SchemaForm schema={complexSchema} value={{}} onChange={onChange} />);

            const hostInput = screen.getByLabelText("Host");
            fireEvent.change(hostInput, { target: { value: "example.com" } });

            expect(onChange).toHaveBeenCalledWith({
                server: { host: "example.com" }
            });
        });
    });
});
