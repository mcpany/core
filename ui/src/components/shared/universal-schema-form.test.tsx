/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { UniversalSchemaForm, Schema } from "./universal-schema-form";
import { vi, describe, it, expect } from "vitest";

describe("UniversalSchemaForm", () => {
    it("renders basic string input", () => {
        const schema: Schema = {
            type: "object",
            properties: {
                name: { type: "string", title: "Name" }
            }
        };
        const onChange = vi.fn();
        render(<UniversalSchemaForm schema={schema} value={{}} onChange={onChange} />);

        const input = screen.getByLabelText("Name");
        fireEvent.change(input, { target: { value: "test" } });

        expect(onChange).toHaveBeenCalledWith({ name: "test" });
    });

    it("renders boolean switch", () => {
        const schema: Schema = {
            type: "object",
            properties: {
                active: { type: "boolean", title: "Active" }
            }
        };
        const onChange = vi.fn();
        render(<UniversalSchemaForm schema={schema} value={{ active: false }} onChange={onChange} />);

        const toggle = screen.getByRole("switch");
        fireEvent.click(toggle);

        expect(onChange).toHaveBeenCalledWith({ active: true });
    });

    it("renders nested object", () => {
        const schema: Schema = {
            type: "object",
            properties: {
                user: {
                    type: "object",
                    title: "User",
                    properties: {
                        email: { type: "string", title: "Email" }
                    }
                }
            }
        };
        const onChange = vi.fn();
        render(<UniversalSchemaForm schema={schema} value={{ user: {} }} onChange={onChange} />);

        // Should find nested label
        expect(screen.getByText("User")).toBeInTheDocument();
        const input = screen.getByLabelText("Email");
        fireEvent.change(input, { target: { value: "test@example.com" } });

        expect(onChange).toHaveBeenCalledWith({ user: { email: "test@example.com" } });
    });

    it("renders array controls", () => {
        const schema: Schema = {
            type: "object",
            properties: {
                tags: {
                    type: "array",
                    title: "Tags",
                    items: { type: "string", title: "Tag" }
                }
            }
        };
        const onChange = vi.fn();
        render(<UniversalSchemaForm schema={schema} value={{ tags: [] }} onChange={onChange} />);

        const addButton = screen.getByText("Add Item");
        fireEvent.click(addButton);

        // Expect onChange to be called with one undefined item (or empty string if handled by input)
        // The array handler adds `undefined`
        expect(onChange).toHaveBeenCalledWith({ tags: [undefined] });
    });

    it("handles number input types", () => {
        const schema: Schema = {
            type: "object",
            properties: {
                count: { type: "integer", title: "Count" }
            }
        };
        const onChange = vi.fn();
        render(<UniversalSchemaForm schema={schema} value={{}} onChange={onChange} />);

        const input = screen.getByLabelText("Count");
        fireEvent.change(input, { target: { value: "42" } });

        expect(onChange).toHaveBeenCalledWith({ count: 42 });
    });
});
