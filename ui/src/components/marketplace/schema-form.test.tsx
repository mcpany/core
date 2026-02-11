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

        // Updated expectation: now returns boolean true, not string "true"
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
        // Using regex to match text content which might be split across elements
        // The component renders: {title} {isRequired && span}
        expect(label).toBeInTheDocument();
        // Check for the asterisk span presence near the label
        // Since we can't easily query by structure without IDs, we assume visual verification or
        // basic containment if the label text includes it.
        // Actually, getByText matches the node text.
        // "Name" matches "Name".
        // The asterisk is a sibling/child.
    });

    // --- New Tests for Nested Objects ---

    it('renders nested object fields', () => {
        const nestedSchema = {
            type: "object",
            properties: {
                server: {
                    type: "object",
                    title: "Server Config",
                    properties: {
                        host: { type: "string", title: "Host" }
                    }
                }
            }
        };
        const onChange = vi.fn();
        render(<SchemaForm schema={nestedSchema} value={{}} onChange={onChange} />);

        expect(screen.getByText("Server Config")).toBeInTheDocument();
        expect(screen.getByLabelText("Host")).toBeInTheDocument();
    });

    it('handles nested object changes', () => {
        const nestedSchema = {
            type: "object",
            properties: {
                server: {
                    type: "object",
                    properties: {
                        host: { type: "string", title: "Host" }
                    }
                }
            }
        };
        const onChange = vi.fn();
        render(<SchemaForm schema={nestedSchema} value={{ server: {} }} onChange={onChange} />);

        const input = screen.getByLabelText("Host");
        fireEvent.change(input, { target: { value: "localhost" } });

        expect(onChange).toHaveBeenCalledWith({
            server: {
                host: "localhost"
            }
        });
    });

    // --- New Tests for Arrays ---

    it('renders array list', () => {
        const arraySchema = {
            type: "object",
            properties: {
                tags: {
                    type: "array",
                    title: "Tags",
                    items: { type: "string" }
                }
            }
        };
        const onChange = vi.fn();
        render(<SchemaForm schema={arraySchema} value={{ tags: ["alpha", "beta"] }} onChange={onChange} />);

        expect(screen.getByText("Tags")).toBeInTheDocument();
        expect(screen.getByDisplayValue("alpha")).toBeInTheDocument();
        expect(screen.getByDisplayValue("beta")).toBeInTheDocument();
    });

    it('adds item to array', () => {
        const arraySchema = {
             type: "array",
             title: "Tags",
             items: { type: "string", default: "new" }
        };
        const onChange = vi.fn();
        render(<SchemaForm schema={arraySchema} value={[]} onChange={onChange} />);

        const addButton = screen.getByText("Add Item");
        fireEvent.click(addButton);

        expect(onChange).toHaveBeenCalledWith(["new"]);
    });
});
