/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from "@testing-library/react";
import { SchemaViewer, Schema } from "@/components/tools/schema-viewer";
import React from 'react';

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe("SchemaViewer", () => {
  const sampleSchema: Schema = {
    type: "object",
    description: "A person object",
    required: ["name"],
    properties: {
      name: {
        type: "string",
        description: "The name of the person"
      },
      age: {
        type: "integer",
        description: "Age in years"
      },
      tags: {
        type: "array",
        items: {
          type: "string"
        }
      }
    }
  };

  it("renders the root object", () => {
    render(<SchemaViewer schema={sampleSchema} name="root" />);
    expect(screen.getByText("root")).toBeInTheDocument();
    // The text content is lowercase "object", usually styled uppercase via CSS
    expect(screen.getByText("object")).toBeInTheDocument();
  });

  it("renders properties", () => {
    render(<SchemaViewer schema={sampleSchema} />);
    expect(screen.getByText("name")).toBeInTheDocument();
    // "string" appears multiple times
    expect(screen.getAllByText("string").length).toBeGreaterThan(0);
    expect(screen.getByText("age")).toBeInTheDocument();
    expect(screen.getByText("integer")).toBeInTheDocument();
  });

  it("indicates required fields", () => {
    render(<SchemaViewer schema={sampleSchema} />);
    // The asterisk is in a span with title="Required"
    const requiredMarks = screen.getAllByTitle("Required");
    expect(requiredMarks.length).toBeGreaterThan(0);
  });

  it("renders array items", () => {
      render(<SchemaViewer schema={sampleSchema} />);
      expect(screen.getByText("tags")).toBeInTheDocument();
      expect(screen.getByText("array")).toBeInTheDocument();
      // "Items:" label
      expect(screen.getByText("Items:")).toBeInTheDocument();
  });
});
