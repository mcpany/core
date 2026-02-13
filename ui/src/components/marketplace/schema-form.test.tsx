/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, fireEvent, within } from '@testing-library/react';
import { SchemaForm } from './schema-form';
import { vi } from 'vitest';

// Mock UI components to simplify testing
vi.mock("@/components/ui/input", () => ({
  Input: ({ value, onChange, placeholder, type, id }: any) => (
    <input
      id={id}
      data-testid="input"
      type={type}
      value={value}
      onChange={onChange}
      placeholder={placeholder}
    />
  ),
}));

vi.mock("@/components/ui/label", () => ({
  Label: ({ children, htmlFor }: any) => <label htmlFor={htmlFor}>{children}</label>,
}));

vi.mock("@/components/ui/checkbox", () => ({
  Checkbox: ({ checked, onCheckedChange, id }: any) => (
    <input
      id={id}
      type="checkbox"
      checked={checked}
      onChange={(e) => onCheckedChange(e.target.checked)}
    />
  ),
}));

vi.mock("@/components/ui/select", () => ({
  Select: ({ value, onValueChange, children }: any) => (
    <div data-testid="select">
      <select value={value} onChange={(e) => onValueChange(e.target.value)}>
        {React.Children.map(children, (child) => {
            // Traverse down to find SelectItem/SelectContent if structure is complex
            // But here we'll mock SelectContent/SelectItem separately
            return <option value="mocked">Mocked</option> // Simple fallback
        })}
        {/* We need to render the children so the mock sub-components render */}
        {children}
      </select>
    </div>
  ),
  SelectTrigger: ({ children }: any) => <div>{children}</div>,
  SelectValue: () => <span>Select Value</span>,
  SelectContent: ({ children }: any) => <div>{children}</div>,
  SelectItem: ({ value, children }: any) => <option value={value}>{children}</option>,
}));

vi.mock("@/components/ui/card", () => ({
  Card: ({ children }: any) => <div data-testid="card" className="border-dashed">{children}</div>,
  CardHeader: ({ children }: any) => <div data-testid="card-header">{children}</div>,
  CardTitle: ({ children }: any) => <div data-testid="card-title">{children}</div>,
  CardContent: ({ children }: any) => <div data-testid="card-content">{children}</div>,
}));

vi.mock("@/components/ui/button", () => ({
  Button: ({ children, onClick }: any) => (
    <button onClick={onClick}>{children}</button>
  ),
}));

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

    it('renders nested object schema', () => {
        const schema = {
          type: "object",
          properties: {
            server: {
              type: "object",
              title: "Server Details",
              properties: {
                host: { type: "string" },
                port: { type: "integer" }
              }
            }
          }
        };

        const onChange = vi.fn();
        render(<SchemaForm schema={schema} value={{}} onChange={onChange} />);

        // Check for nested title
        expect(screen.getByText("Server Details")).toBeInTheDocument();
        // Check for nested field labels
        expect(screen.getByText("host")).toBeInTheDocument();
        expect(screen.getByText("port")).toBeInTheDocument();
    });

    it('updates nested object value', () => {
        const schema = {
          type: "object",
          properties: {
            server: {
              type: "object",
              properties: {
                host: { type: "string" }
              }
            }
          }
        };

        // We need to simulate parent state management
        let value = {};
        const onChange = vi.fn((newVal) => { value = newVal; });
        const { rerender } = render(<SchemaForm schema={schema} value={value} onChange={onChange} />);

        // Find inputs. There should be one input for 'host'.
        // Since we mocked Input with data-testid="input"
        const input = screen.getByLabelText(/host/);
        fireEvent.change(input, { target: { value: "localhost" } });

        expect(onChange).toHaveBeenCalledWith({
          server: { host: "localhost" }
        });
    });

    it('renders array of items', () => {
        const schema = {
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
        render(<SchemaForm schema={schema} value={{ tags: ["tag1"] }} onChange={onChange} />);

        expect(screen.getByText("Tags")).toBeInTheDocument();
        const inputs = screen.getAllByTestId("input");
        expect(inputs[0]).toHaveValue("tag1");
    });

    it('adds item to array', () => {
        const schema = {
          type: "object",
          properties: {
            tags: {
              type: "array",
              title: "Tags",
              items: { type: "string" }
            }
          }
        };

        let value = { tags: [] };
        const onChange = vi.fn((newVal) => { value = newVal; });
        render(<SchemaForm schema={schema} value={value} onChange={onChange} />);

        const addButton = screen.getByText("Add Item");
        fireEvent.click(addButton);

        expect(onChange).toHaveBeenCalledWith({
          tags: [""]
        });
    });
});
