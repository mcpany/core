/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { vi, describe, it, expect } from "vitest";
import { StructuredDataViewer } from "./structured-data-viewer";

vi.mock("@/hooks/use-toast", () => ({
  useToast: vi.fn(() => ({
    toast: vi.fn(),
  })),
}));

describe("StructuredDataViewer", () => {
  it("renders object correctly", () => {
    const data = { name: "Test User", age: 30 };
    render(<StructuredDataViewer data={data} />);
    expect(screen.getByText('"Test User"')).toBeInTheDocument();
    expect(screen.getByText("30")).toBeInTheDocument();
  });

  it("handles copy button click", () => {
    Object.assign(navigator, {
      clipboard: {
        writeText: vi.fn().mockImplementation(() => Promise.resolve()),
      },
    });

    const data = { key: "value" };
    const { container } = render(<StructuredDataViewer data={data} />);
    const btn = container.querySelector("button");
    fireEvent.click(btn!);
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      JSON.stringify(data, null, 2),
    );
  });
});
