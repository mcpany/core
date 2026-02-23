/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { SmartTemplateEditor } from "../smart-template-editor";
import React from "react";
import { describe, it, expect, vi } from 'vitest';

// Mock TooltipProvider since it might be used by Badge or other components
vi.mock("@/components/ui/tooltip", () => ({
  Tooltip: ({ children }: any) => <div>{children}</div>,
  TooltipContent: ({ children }: any) => <div>{children}</div>,
  TooltipTrigger: ({ children }: any) => <div>{children}</div>,
  TooltipProvider: ({ children }: any) => <div>{children}</div>,
}));

// Mock Toast
vi.mock("@/hooks/use-toast", () => ({
  useToast: () => ({ toast: vi.fn() }),
}));

describe("SmartTemplateEditor", () => {
  it("renders correctly with initial values", async () => {
    render(
      <SmartTemplateEditor
        value="Hello {{ name }}"
        onChange={vi.fn()}
        initialTestData='{"name": "World"}'
      />
    );

    // Check if template is rendered in textarea
    // We expect two textareas: one for template, one for test data
    const textareas = screen.getAllByRole("textbox");
    expect(textareas[0]).toHaveValue("Hello {{ name }}");
    expect(textareas[1]).toHaveValue('{"name": "World"}');

    // Check preview
    await waitFor(() => {
        expect(screen.getByText("Hello World")).toBeInTheDocument();
    });
  });

  it("updates preview when test data changes", async () => {
    render(
      <SmartTemplateEditor
        value="Hello {{ name }}"
        onChange={vi.fn()}
        initialTestData='{"name": "World"}'
      />
    );

    await waitFor(() => {
        expect(screen.getByText("Hello World")).toBeInTheDocument();
    });

    const testDataInput = screen.getAllByRole("textbox")[1]; // Second textarea is test data
    fireEvent.change(testDataInput, { target: { value: '{"name": "Jest"}' } });

    await waitFor(() => {
        expect(screen.getByText("Hello Jest")).toBeInTheDocument();
    });
  });

  it("shows error on invalid JSON", async () => {
    render(
      <SmartTemplateEditor
        value="Hello {{ name }}"
        onChange={vi.fn()}
        initialTestData='{"name": "World"}'
      />
    );

    const testDataInput = screen.getAllByRole("textbox")[1];
    fireEvent.change(testDataInput, { target: { value: '{invalid' } });

    await waitFor(() => {
        expect(screen.getByText(/Invalid JSON Test Data/)).toBeInTheDocument();
    });
  });

  it("shows variables as badges", () => {
    render(
      <SmartTemplateEditor
        value=""
        onChange={vi.fn()}
        variables={["userId", "apiKey"]}
      />
    );

    expect(screen.getByText("userId")).toBeInTheDocument();
    expect(screen.getByText("apiKey")).toBeInTheDocument();
  });
});
