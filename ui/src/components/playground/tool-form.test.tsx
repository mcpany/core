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
  schema: {
    type: "object",
    properties: {
      username: { type: "string", description: "The username" },
      age: { type: "integer", description: "The age" },
      isActive: { type: "boolean", description: "Is active?" },
      role: { type: "string", enum: ["admin", "user"], description: "User role" }
    },
    required: ["username", "age"]
  }
};

describe("ToolForm", () => {
  it("renders form fields based on schema", () => {
    const onSubmit = vi.fn();
    const onCancel = vi.fn();

    render(<ToolForm tool={mockTool} onSubmit={onSubmit} onCancel={onCancel} />);

    expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/age/i)).toBeInTheDocument();
    // boolean is a switch, usually checked by role='switch' or similar, but label association works
    expect(screen.getByText(/isActive/i)).toBeInTheDocument();
    // select is tricky in shadcn/radix, but we can check for the trigger text or label
    const roleLabels = screen.getAllByText(/role/i);
    expect(roleLabels.length).toBeGreaterThan(0);
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
    const switchEl = screen.getByRole("switch");
    await user.click(switchEl);

    // Select role (Combobox/Select in shadcn usually opens a portal)
    // For simplicity in unit test with radis-ui/select mock might be needed or just skip complex interaction
    // Let's rely on text input for now.

    // Submit
    const submitBtn = screen.getByRole("button", { name: /run tool/i });
    await user.click(submitBtn);

    await waitFor(() => {
        expect(onSubmit).toHaveBeenCalledWith({
            username: "jdoe",
            age: 30, // processed as number
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
    const submitBtn = screen.getByRole("button", { name: /run tool/i });
    await user.click(submitBtn);

    await waitFor(() => {
       // Check for validation messages
       // The component renders "This field is required"
       const errors = screen.getAllByText(/this field is required/i);
       expect(errors.length).toBeGreaterThan(0);
       expect(onSubmit).not.toHaveBeenCalled();
    });
  });
});
