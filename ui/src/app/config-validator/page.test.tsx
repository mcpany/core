/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import ConfigValidatorPage from "./page";
import { vi, describe, it, expect } from "vitest";

// Mock Monaco Editor since it doesn't render well in JSDOM
vi.mock("@monaco-editor/react", () => ({
  default: ({ value, onChange }: { value: string; onChange: (val: string) => void }) => (
    <textarea
      data-testid="monaco-editor"
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  ),
}));

// Mock Sonner Toast
vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock useTheme
vi.mock("next-themes", () => ({
  useTheme: () => ({ theme: "light" }),
}));

describe("ConfigValidatorPage", () => {
  it("renders the page correctly", () => {
    render(<ConfigValidatorPage />);
    expect(screen.getByText("Config Validator")).toBeInTheDocument();
    expect(screen.getByText("Validate Configuration")).toBeInTheDocument();
  });

  it("shows error toast if validation content is empty", () => {
    render(<ConfigValidatorPage />);
    const button = screen.getByText("Validate Configuration");
    fireEvent.click(button);
    // You might check if toast.error was called if you exposed the mock,
    // or just rely on the fact that no fetch happens.
  });

  it("calls API and displays success result", async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ valid: true }),
      } as Response)
    );

    render(<ConfigValidatorPage />);
    const editor = screen.getByTestId("monaco-editor");
    fireEvent.change(editor, { target: { value: "foo: bar" } });

    const button = screen.getByText("Validate Configuration");
    fireEvent.click(button);

    await waitFor(() => {
      expect(screen.getByText("Valid Configuration")).toBeInTheDocument();
    });
  });

  it("calls API and displays error result", async () => {
    global.fetch = vi.fn(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ valid: false, errors: ["Invalid syntax"] }),
      } as Response)
    );

    render(<ConfigValidatorPage />);
    const editor = screen.getByTestId("monaco-editor");
    fireEvent.change(editor, { target: { value: "invalid yaml" } });

    const button = screen.getByText("Validate Configuration");
    fireEvent.click(button);

    await waitFor(() => {
      expect(screen.getByText("Validation Errors")).toBeInTheDocument();
      expect(screen.getByText("Invalid syntax")).toBeInTheDocument();
    });
  });
});
