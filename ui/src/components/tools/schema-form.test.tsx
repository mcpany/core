/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { SchemaForm } from "./schema-form";
import { describe, it, expect, vi } from "vitest";
import { TooltipProvider } from "@/components/ui/tooltip";

const renderWithTooltip = (component: React.ReactNode) => {
    return render(
        <TooltipProvider>
            {component}
        </TooltipProvider>
    );
};

describe("SchemaForm (Tools)", () => {
  it("renders basic string input", () => {
    const schema = {
      type: "string",
      description: "Enter text"
    };
    const onChange = vi.fn();
    renderWithTooltip(<SchemaForm schema={schema} value="" onChange={onChange} label="test-input" />);

    expect(screen.getByText("test-input")).toBeDefined();
    expect(screen.getByPlaceholderText("Enter text")).toBeDefined();
  });

  it("renders number input and updates with number type", () => {
    const schema = {
      type: "number"
    };
    const onChange = vi.fn();
    renderWithTooltip(<SchemaForm schema={schema} value={10} onChange={onChange} label="age" />);

    const input = screen.getByPlaceholderText("0");
    fireEvent.change(input, { target: { value: "20" } });
    expect(onChange).toHaveBeenCalledWith(20);
  });

  it("renders boolean switch", () => {
    const schema = {
      type: "boolean"
    };
    const onChange = vi.fn();
    renderWithTooltip(<SchemaForm schema={schema} value={false} onChange={onChange} label="isEnabled" />);

    expect(screen.getByText("isEnabled")).toBeDefined();
    const switchEl = screen.getByRole("switch");
    fireEvent.click(switchEl);
    expect(onChange).toHaveBeenCalledWith(true);
  });

  it("renders enum select", async () => {
    const schema = {
      type: "string",
      enum: ["A", "B"]
    };
    const onChange = vi.fn();
    renderWithTooltip(<SchemaForm schema={schema} value="A" onChange={onChange} label="category" />);

    // Trigger select
    const trigger = screen.getByRole("combobox");
    expect(trigger).toBeDefined();
    // Interacting with Select in tests can be tricky without userEvent, but basic render check is good.
  });

  it("renders nested object properties", () => {
    const schema = {
      type: "object",
      properties: {
        nested: { type: "string" }
      }
    };
    const onChange = vi.fn();
    renderWithTooltip(<SchemaForm schema={schema} value={{}} onChange={onChange} />);

    expect(screen.getByText("nested")).toBeDefined();
  });

  it("updates nested object value correctly", () => {
    const schema = {
      type: "object",
      properties: {
        nested: { type: "string" }
      }
    };
    const onChange = vi.fn();
    renderWithTooltip(<SchemaForm schema={schema} value={{}} onChange={onChange} />);

    const input = screen.getByRole("textbox");
    fireEvent.change(input, { target: { value: "hello" } });

    expect(onChange).toHaveBeenCalledWith({ nested: "hello" });
  });
});
