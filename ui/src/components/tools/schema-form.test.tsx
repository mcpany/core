/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { SchemaForm } from "./schema-form";
import userEvent from "@testing-library/user-event";
import { vi } from "vitest";

describe("SchemaForm", () => {
  it("renders FileInput for contentEncoding='base64'", () => {
    const schema = {
      type: "object",
      properties: {
        image: {
          type: "string",
          contentEncoding: "base64",
          contentMediaType: "image/png",
          description: "An image file"
        }
      }
    };
    const onChange = vi.fn();

    render(<SchemaForm schema={schema} value={{}} onChange={onChange} />);

    // Check for label
    expect(screen.getByText("image")).toBeInTheDocument();
    expect(screen.getByText("An image file")).toBeInTheDocument();

    // Check for FileInput button
    expect(screen.getByText("Select File")).toBeInTheDocument();
  });

  it("calls onChange with base64 string when file is selected", async () => {
    const schema = {
      type: "object",
      properties: {
        file: {
          type: "string",
          contentEncoding: "base64"
        }
      }
    };
    const onChange = vi.fn();

    const file = new File(["hello"], "hello.txt", { type: "text/plain" });

    // Find input by generic means
    // render returns container.
    const { container } = render(<SchemaForm schema={schema} value={{}} onChange={onChange} />);
    const fileInput = container.querySelector("input[type='file']");

    if (!fileInput) throw new Error("File input not found");

    await userEvent.upload(fileInput, file);

    // FileReader is async. FileInput handles it via onload.
    // We need to wait for onChange to be called.

    // Note: in JSDOM, FileReader might need to be mocked if it doesn't work as expected,
    // but usually it works for basic text.
    // "hello" in base64 is "aGVsbG8="

    await waitFor(() => {
        expect(onChange).toHaveBeenCalledWith({ file: "aGVsbG8=" });
    });
  });
});
