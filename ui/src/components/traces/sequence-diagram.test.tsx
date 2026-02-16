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

const nestedTrace: Trace = {
    id: "trace-nested",
    timestamp: new Date().toISOString(),
    totalDuration: 200,
    status: "success",
    trigger: "user",
    rootSpan: {
      id: "span-root",
      name: "orchestrator",
      type: "tool",
      startTime: 1000,
      endTime: 1200,
      status: "success",
      input: { task: "do complex thing" },
      output: { result: "done" },
      children: [
          {
              id: "span-child-1",
              name: "sub-tool",
              type: "tool",
              startTime: 1050,
              endTime: 1150,
              status: "success",
              input: { sub: "task" },
              output: { sub: "done" },
              children: []
          },
          {
              id: "span-child-2",
              name: "weather-service",
              type: "service",
              serviceName: "wttr.in",
              startTime: 1160,
              endTime: 1190,
              status: "success",
              input: { city: "London" },
              output: { temp: 20 },
              children: []
          }
      ],
    },
  };

describe("SequenceDiagram", () => {
  it("renders participants correctly for simple trace", () => {
    render(<SequenceDiagram trace={mockTrace} />);
    expect(screen.getByText("Client")).toBeInTheDocument();
    expect(screen.getByText("MCP Core")).toBeInTheDocument();
    expect(screen.getByText("test-tool")).toBeInTheDocument();
  });

  it("renders interaction labels for simple trace", () => {
    render(<SequenceDiagram trace={mockTrace} />);
    expect(screen.getByText("Execute Request")).toBeInTheDocument();
    expect(screen.getByText("Call test-tool")).toBeInTheDocument();
    // "Result" is the label for return from tool
    expect(screen.getAllByText("Result")[0]).toBeInTheDocument();
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
    expect(screen.getByText("Client requests execution")).toBeInTheDocument();
    // Check payload
    expect(screen.getByText(/value/)).toBeInTheDocument();
  });

  it("renders nested interactions correctly", () => {
      render(<SequenceDiagram trace={nestedTrace} />);

      // Participants
      expect(screen.getByText("orchestrator")).toBeInTheDocument();
      expect(screen.getByText("sub-tool")).toBeInTheDocument();
      expect(screen.getByText("wttr.in")).toBeInTheDocument();

      // Interactions
      expect(screen.getByText("Call orchestrator")).toBeInTheDocument();
      expect(screen.getByText("Call sub-tool")).toBeInTheDocument();
      expect(screen.getByText("Access weather-service")).toBeInTheDocument();
  });

  it("calculates X coordinates correctly (Optimization Verification)", () => {
    const { container } = render(<SequenceDiagram trace={mockTrace} />);

    // Check lines for participants
    // Client (index 0): x = 100
    // MCP Core (index 1): x = 350
    // test-tool (index 2): x = 600

    // Select all vertical lines (lifelines)
    // They have y1=80 and y2 large.
    const lines = container.querySelectorAll("line");
    // We expect 3 vertical lines + interaction lines.
    // Vertical lines are the ones with strokeDasharray="6 6"
    const verticalLines = Array.from(lines).filter(l => l.getAttribute("stroke-dasharray") === "6 6");

    expect(verticalLines).toHaveLength(3);
    expect(verticalLines[0]).toHaveAttribute("x1", "100");
    expect(verticalLines[1]).toHaveAttribute("x1", "350");
    expect(verticalLines[2]).toHaveAttribute("x1", "600");
  });
});
