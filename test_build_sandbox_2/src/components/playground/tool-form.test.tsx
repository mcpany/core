/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ToolForm } from "./tool-form";
import { ToolDefinition } from "@/lib/client";
import { vi } from "vitest";

// Mock ToolDefinition
const mockTool: ToolDefinition = {
  name: "test-tool",
  description: "A test tool",
  inputSchema: {
    type: "object",
    properties: {
      username: { type: "string", description: "The username" },
      age: { type: "string", description: "The age" }, // Defined as string in schema, so input will be string
      isActive: { type: "boolean", description: "Is active status" },
      role: { type: "string", enum: ["admin", "user"], description: "The role" }
    },
    required: ["username"]
  },
  serviceId: "test-service",
  title: "Test Tool",
  isStream: false,
  readOnlyHint: false,
  destructiveHint: false,
  idempotentHint: false,
  openWorldHint: false,
  callId: "",
  profiles: [],
  tags: [],
  mergeStrategy: 0,
  disable: false
};

describe("ToolForm", () => {
  it("renders form fields based on schema", () => {
    const onSubmit = vi.fn();
    const onCancel = vi.fn();

    render(<ToolForm tool={mockTool} onSubmit={onSubmit} onCancel={onCancel} />);

    expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/age/i)).toBeInTheDocument();
    // boolean is a switch
    // The label "isActive" should be present.
    expect(screen.getByText(/isActive/i)).toBeInTheDocument();
  });

  it("calls onSubmit with correct data", async () => {
    const onSubmit = vi.fn();
    const onCancel = vi.fn();
    const user = userEvent.setup();

    render(<ToolForm tool={mockTool} onSubmit={onSubmit} onCancel={onCancel} />);

    // Fill username
    await user.type(screen.getByLabelText(/username/i), "jdoe");

    // Fill age
    await user.type(screen.getByLabelText(/age/i), "30");

    // Toggle isActive (switch)
    // Find switch by label association or role
    // The switch ID is "isActive", label for "isActive" points to it.
    // However, shadcn switch might be nested.
    // Let's try getting by label text "isActive" which is in the Label component.
    // Actually, simpler to find the switch role.
    const switchEl = screen.getByRole("switch");
    await user.click(switchEl);

    // Submit
    const submitBtn = screen.getByRole("button", { name: /build command/i });
    await user.click(submitBtn);

    await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith({
            username: "jdoe",
            age: "30", // Schema says string, so input returns string.
            isActive: true
        });
    });
  });

  it("shows validation errors for required fields", async () => {
    const onSubmit = vi.fn();
    const onCancel = vi.fn();
    const user = userEvent.setup();

    render(<ToolForm tool={mockTool} onSubmit={onSubmit} onCancel={onCancel} />);

    // Submit without filling anything
    const submitBtn = screen.getByRole("button", { name: /build command/i });
    await user.click(submitBtn);

    await waitFor(() => {
       // Check for validation messages
       const errors = screen.getAllByText(/this field is required/i);
       expect(errors.length).toBeGreaterThan(0);
       expect(onSubmit).not.toHaveBeenCalled();
    });
  });

  it("validates complex schema rules (pattern)", async () => {
     const complexTool: ToolDefinition = {
        ...mockTool,
        inputSchema: {
            type: "object",
            properties: {
                email: { type: "string", format: "email" },
                code: { type: "string", pattern: "^[A-Z]{3}$" }
            }
        }
     };

     const onSubmit = vi.fn();
     const onCancel = vi.fn();
     const user = userEvent.setup();

     render(<ToolForm tool={complexTool} onSubmit={onSubmit} onCancel={onCancel} />);

     // Invalid Email
     await user.type(screen.getByLabelText(/email/i), "not-an-email");
     // Invalid Code
     await user.type(screen.getByLabelText(/code/i), "abc");

     const submitBtn = screen.getByRole("button", { name: /build command/i });
     await user.click(submitBtn);

     await waitFor(() => {
         // Expect validation errors
         // AJV default error messages
         expect(screen.getByText(/must match format "email"/i)).toBeInTheDocument();
         expect(screen.getByText(/must match pattern "\^\[A-Z\]\{3\}\$"/i)).toBeInTheDocument();
         expect(onSubmit).not.toHaveBeenCalled();
     });

     // Fix data
     await user.clear(screen.getByLabelText(/email/i));
     await user.type(screen.getByLabelText(/email/i), "test@example.com");

     await user.clear(screen.getByLabelText(/code/i));
     await user.type(screen.getByLabelText(/code/i), "ABC");

     await user.click(submitBtn);

     await waitFor(() => {
         expect(onSubmit).toHaveBeenCalledWith({
             email: "test@example.com",
             code: "ABC"
         });
     });
  });
});
