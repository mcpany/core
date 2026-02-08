/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ToolForm } from "@/components/playground/tool-form";
import { ToolDefinition } from "@/lib/client";
import { vi } from "vitest";

// Mock useToast
vi.mock("@/hooks/use-toast", () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

// Mock ToolDefinition
const mockTool: ToolDefinition = {
  name: "test_tool",
  description: "A test tool",
  inputSchema: {
    type: "object",
    properties: {
      foo: { type: "string" }
    },
    required: ["foo"]
  }
};

describe("ToolForm", () => {
  const handleSubmit = vi.fn();
  const handleCancel = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders form mode by default", () => {
    render(<ToolForm tool={mockTool} onSubmit={handleSubmit} onCancel={handleCancel} />);
    expect(screen.getByRole("tab", { name: "Form" })).toHaveAttribute("data-state", "active");
    expect(screen.getByRole("button", { name: "Build Command" })).toBeInTheDocument();
  });

  it("switches to Schema tab and displays schema", async () => {
    const user = userEvent.setup();
    render(<ToolForm tool={mockTool} onSubmit={handleSubmit} onCancel={handleCancel} />);

    const schemaTab = screen.getByRole("tab", { name: "Schema" });
    await user.click(schemaTab);

    expect(schemaTab).toHaveAttribute("data-state", "active");
    // Verify schema content is displayed (basic check)
    // We check for key strings because syntax highlighting might split the JSON structure
    // We use getAllByText because the token might appear multiple times (e.g. properties and required)
    expect(screen.getAllByText(/foo/).length).toBeGreaterThan(0);
  });

  it("includes a copy button in schema view", async () => {
    // Mock clipboard
    const writeText = vi.fn().mockResolvedValue(undefined);
    // If navigator.clipboard exists, spy on it, otherwise define it
    if (navigator.clipboard) {
        vi.spyOn(navigator.clipboard, 'writeText').mockImplementation(writeText);
    } else {
        Object.defineProperty(navigator, 'clipboard', {
            value: { writeText },
            writable: true
        });
    }

    const user = userEvent.setup();
    render(<ToolForm tool={mockTool} onSubmit={handleSubmit} onCancel={handleCancel} />);

    // Switch to Schema tab
    await user.click(screen.getByRole("tab", { name: "Schema" }));

    // Find copy button (by title)
    // Note: JsonTree renders a copy button for each object node, so there might be multiple.
    // We want the root one.
    const copyBtns = screen.getAllByTitle("Copy JSON");
    expect(copyBtns.length).toBeGreaterThan(0);
    const copyBtn = copyBtns[0];

    // Use fireEvent to bypass opacity/visibility checks
    fireEvent.click(copyBtn);
    expect(writeText).toHaveBeenCalledWith(JSON.stringify(mockTool.inputSchema, null, 2));
  });
});
