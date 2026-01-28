/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { SequenceDiagram } from "./sequence-diagram";
import { Trace } from "@/types/trace";
import { describe, it, expect } from "vitest";

const mockTrace: Trace = {
  id: "trace-123",
  timestamp: new Date().toISOString(),
  totalDuration: 100,
  status: "success",
  trigger: "user",
  rootSpan: {
    id: "span-1",
    name: "test-tool",
    type: "tool",
    startTime: 1000,
    endTime: 1100,
    status: "success",
    input: { arg: "value" },
    output: { result: "ok" },
    children: [],
  },
};

describe("SequenceDiagram", () => {
  it("renders participants correctly", () => {
    render(<SequenceDiagram trace={mockTrace} />);
    expect(screen.getByText("Client")).toBeInTheDocument();
    expect(screen.getByText("MCP Core")).toBeInTheDocument();
    expect(screen.getByText("Tool")).toBeInTheDocument();
  });

  it("renders interaction labels", () => {
    render(<SequenceDiagram trace={mockTrace} />);
    expect(screen.getByText("Execute Request")).toBeInTheDocument();
    expect(screen.getByText("Call test-tool")).toBeInTheDocument();
    expect(screen.getByText("Execution Result")).toBeInTheDocument();
    expect(screen.getByText("Response")).toBeInTheDocument();
  });

  it("opens dialog on interaction click", () => {
    render(<SequenceDiagram trace={mockTrace} />);
    const requestLabel = screen.getByText("Execute Request");

    // Find the clickable group (parent g element usually, or the label itself acts as trigger)
    // The onClick is on the group <g>.
    // We can click the text which bubbles up.
    fireEvent.click(requestLabel);

    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Client requests tool execution")).toBeInTheDocument();
    // Check payload
    expect(screen.getByText(/value/)).toBeInTheDocument();
  });
});
