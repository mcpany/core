import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { ToolPresets } from "./tool-presets";
import { toast } from "sonner";
import React from "react";
import { vi, describe, beforeEach, test, expect } from "vitest";

// Mock sonner toast
vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
  },
}));

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock ScrollArea to avoid complexities with Radix ScrollArea in JSDOM
vi.mock("@/components/ui/scroll-area", () => ({
  ScrollArea: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
}));

describe("ToolPresets", () => {
  const toolName = "test-tool";
  const onSelect = vi.fn();
  const currentData = { foo: "bar" };

  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    vi.spyOn(window, 'confirm').mockImplementation(() => true);
  });

  test("renders empty state", () => {
    render(<ToolPresets toolName={toolName} currentData={currentData} onSelect={onSelect} />);

    // Open popover
    fireEvent.click(screen.getByRole("button", { name: /Manage Presets/i }));

    expect(screen.getByText("Presets")).toBeInTheDocument();
    expect(screen.getByText("No saved presets for this tool.")).toBeInTheDocument();
  });

  test("saves a new preset", async () => {
    render(<ToolPresets toolName={toolName} currentData={currentData} onSelect={onSelect} />);

    // Open popover
    fireEvent.click(screen.getByRole("button", { name: /Manage Presets/i }));

    // Click Add button (Title: Create New Preset)
    fireEvent.click(screen.getByTitle("Create New Preset"));

    // Enter name
    const input = screen.getByPlaceholderText("Preset Name");
    fireEvent.change(input, { target: { value: "My Preset" } });

    // Save via Enter
    fireEvent.keyDown(input, { key: "Enter" });

    await waitFor(() => {
        expect(localStorage.getItem(`mcpany-presets-${toolName}`)).toContain("My Preset");
    });

    expect(toast.success).toHaveBeenCalledWith("Preset saved");
    expect(screen.getByText("My Preset")).toBeInTheDocument();
  });

  test("loads a preset", () => {
    const presets = [{ name: "Existing", data: { a: 1 } }];
    localStorage.setItem(`mcpany-presets-${toolName}`, JSON.stringify(presets));

    render(<ToolPresets toolName={toolName} currentData={currentData} onSelect={onSelect} />);

    // Open popover
    fireEvent.click(screen.getByRole("button", { name: /Manage Presets/i }));

    // Click preset text
    fireEvent.click(screen.getByText("Existing"));

    expect(onSelect).toHaveBeenCalledWith({ a: 1 });
    expect(toast.success).toHaveBeenCalledWith("Loaded preset: Existing");
  });

  test("deletes a preset", () => {
      const presets = [{ name: "To Delete", data: { a: 1 } }];
      localStorage.setItem(`mcpany-presets-${toolName}`, JSON.stringify(presets));

      render(<ToolPresets toolName={toolName} currentData={currentData} onSelect={onSelect} />);

      // Open popover
      fireEvent.click(screen.getByRole("button", { name: /Manage Presets/i }));

      // Find delete button by title
      const deleteBtn = screen.getByTitle("Delete Preset");
      fireEvent.click(deleteBtn);

      expect(localStorage.getItem(`mcpany-presets-${toolName}`)).toBe("[]");
      expect(screen.queryByText("To Delete")).not.toBeInTheDocument();
  });
});
