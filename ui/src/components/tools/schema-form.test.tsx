/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { SchemaForm, Schema } from "./schema-form";

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// Mock PointerEvent
global.PointerEvent = class extends Event {
  constructor(type: string, props: any) {
    super(type, props);
  }
} as any;

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

describe("SchemaForm", () => {
  it("renders a text input for string type", () => {
    const schema: Schema = { type: "string" };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value="" onChange={onChange} name="testField" />);

    const input = screen.getByLabelText(/testField/i);
    expect(input).toBeDefined();

    fireEvent.change(input, { target: { value: "hello" } });
    expect(onChange).toHaveBeenCalledWith("hello");
  });

  it("renders a number input for integer type", () => {
    const schema: Schema = { type: "integer" };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value={0} onChange={onChange} name="age" />);

    const input = screen.getByLabelText(/age/i) as HTMLInputElement;
    expect(input.type).toBe("number");

    fireEvent.change(input, { target: { value: "42" } });
    expect(onChange).toHaveBeenCalledWith(42);
  });

  it("renders a switch for boolean type", () => {
    const schema: Schema = { type: "boolean" };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value={false} onChange={onChange} name="isActive" />);

    const switchEl = screen.getByLabelText(/isActive/i);
    expect(switchEl).toBeDefined();

    // Radix Switch is a button
    fireEvent.click(switchEl);
    expect(onChange).toHaveBeenCalledWith(true);
  });

  it("renders nested object fields", () => {
    const schema: Schema = {
      type: "object",
      properties: {
        firstName: { type: "string" },
        lastName: { type: "string" }
      }
    };
    const onChange = vi.fn();
    const value = { firstName: "John", lastName: "Doe" };

    render(<SchemaForm schema={schema} value={value} onChange={onChange} />);

    const firstInput = screen.getByLabelText(/firstName/i);
    expect(firstInput).toBeDefined();

    fireEvent.change(firstInput, { target: { value: "Jane" } });
    expect(onChange).toHaveBeenCalledWith({ firstName: "Jane", lastName: "Doe" });
  });

  it("handles array adding items", async () => {
     const schema: Schema = {
        type: "array",
        items: { type: "string", default: "new item" }
     };
     const onChange = vi.fn();
     const value: string[] = [];

     render(<SchemaForm schema={schema} value={value} onChange={onChange} name="tags" />);

     const addButton = screen.getByText("Add Item");
     fireEvent.click(addButton);

     expect(onChange).toHaveBeenCalledWith(["new item"]);
  });
});
