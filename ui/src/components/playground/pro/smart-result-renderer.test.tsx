/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from "@testing-library/react";
import { SmartResultRenderer } from "./smart-result-renderer";
import { describe, it, expect } from "vitest";

describe("SmartResultRenderer", () => {
  it("renders text content as table/json if applicable", () => {
    const result = {
      content: [
        { type: "text", text: JSON.stringify([{ id: 1, name: "test" }]) }
      ]
    };
    render(<SmartResultRenderer result={result} />);
    // Should show table headers or content
    expect(screen.getByText("test")).toBeInTheDocument();
    // Should show "Table" button if smart view is active
    expect(screen.getByText("Table")).toBeInTheDocument();
  });

  it("renders image content using img tag", () => {
    const result = {
      content: [
        { type: "image", data: "base64data", mimeType: "image/png" }
      ]
    };
    render(<SmartResultRenderer result={result} />);
    const img = screen.getByRole("img");
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute("src", "data:image/png;base64,base64data");
  });

  it("renders mixed content (text and image)", () => {
    const result = {
      content: [
        { type: "text", text: "Here is an image:" },
        { type: "image", data: "base64data", mimeType: "image/png" }
      ]
    };
    render(<SmartResultRenderer result={result} />);
    expect(screen.getByText("Here is an image:")).toBeInTheDocument();
    const img = screen.getByRole("img");
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute("src", "data:image/png;base64,base64data");
  });

  it("renders command output with nested JSON image content", () => {
    // This simulates what a command line tool might return if it outputs MCP JSON
    const nestedContent = JSON.stringify([
      { type: "image", data: "base64data", mimeType: "image/png" }
    ]);
    const result = {
      command: "test",
      stdout: nestedContent
    };
    render(<SmartResultRenderer result={result} />);
    const img = screen.getByRole("img");
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute("src", "data:image/png;base64,base64data");
  });
});
