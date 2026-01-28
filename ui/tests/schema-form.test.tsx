/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { SchemaForm } from "../src/components/playground/schema-form";
import { describe, it, expect, vi } from "vitest";

describe("SchemaForm", () => {
  it("renders basic string input", () => {
    const schema = {
      type: "object",
      properties: {
        name: { type: "string", description: "Your name" },
      },
    };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value={{}} onChange={onChange} />);

    expect(screen.getByText("name")).toBeDefined();
    expect(screen.getByPlaceholderText("name")).toBeDefined();
  });

  it("renders nested object", () => {
    const schema = {
      type: "object",
      properties: {
        user: {
          type: "object",
          properties: {
            first: { type: "string" },
            last: { type: "string" },
          },
        },
      },
    };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value={{}} onChange={onChange} />);

    expect(screen.getByText("user")).toBeDefined();
    expect(screen.getByText("first")).toBeDefined();
    expect(screen.getByText("last")).toBeDefined();
  });

  it("updates values correctly", () => {
    const schema = {
      type: "object",
      properties: {
        age: { type: "integer" },
      },
    };
    const onChange = vi.fn();
    render(<SchemaForm schema={schema} value={{}} onChange={onChange} />);

    const input = screen.getByPlaceholderText("0");
    fireEvent.change(input, { target: { value: "25" } });

    expect(onChange).toHaveBeenCalledWith({ age: 25 });
  });

  it("handles array adding items", () => {
    const schema = {
      type: "object",
      properties: {
        tags: {
          type: "array",
          items: { type: "string" },
        },
      },
    };
    const onChange = vi.fn();
    // Start with empty value
    render(<SchemaForm schema={schema} value={{}} onChange={onChange} />);

    const addButton = screen.getByText("Add Item");
    fireEvent.click(addButton);

    // Should call onChange with new array containing undefined/empty
    expect(onChange).toHaveBeenCalledWith({ tags: [undefined] });
  });
});
