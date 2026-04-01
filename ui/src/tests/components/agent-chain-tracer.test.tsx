/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { AgentChainTracer } from "@/components/dashboard/agent-chain-tracer";
import * as useTracesHook from "@/hooks/use-traces";
import { Trace } from "@/types/trace";

// Mock the trace hook
vi.mock("@/hooks/use-traces", () => ({
  useTraces: vi.fn(),
}));

describe("AgentChainTracer Component", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it("renders with no traces", () => {
    vi.mocked(useTracesHook.useTraces).mockReturnValue({
      traces: [],
      loading: false,
      isConnected: true,
      isPaused: false,
      setIsPaused: vi.fn(),
      clearTraces: vi.fn(),
      refresh: vi.fn(),
    });

    render(<AgentChainTracer />);
    expect(screen.getByText("Agent Chain Tracer (A2A)")).toBeInTheDocument();
  });

  it("renders mapped traces properly based on status", () => {
    const mockTraces: Trace[] = [
      {
        id: "trace-1234",
        timestamp: "2026-04-01T10:00:00Z",
        totalDuration: 150,
        status: "success",
        trigger: "user",
        rootSpan: {
          id: "span-1234",
          name: "action-1",
          type: "tool",
          startTime: 100,
          endTime: 250,
          status: "success",
          serviceName: "agent-alpha",
        },
      },
      {
        id: "trace-error",
        timestamp: "2026-04-01T10:01:00Z",
        totalDuration: 50,
        status: "error",
        trigger: "system",
        rootSpan: {
          id: "span-error",
          name: "action-2",
          type: "tool",
          startTime: 100,
          endTime: 150,
          status: "error",
          serviceName: "agent-beta",
        },
      },
    ];

    vi.mocked(useTracesHook.useTraces).mockReturnValue({
      traces: mockTraces,
      loading: false,
      isConnected: true,
      isPaused: false,
      setIsPaused: vi.fn(),
      clearTraces: vi.fn(),
      refresh: vi.fn(),
    });

    render(<AgentChainTracer />);

    // Validates Agent Names rendered
    expect(screen.getByText("agent-alpha")).toBeInTheDocument();
    expect(screen.getByText("agent-beta")).toBeInTheDocument();

    // Validate Action Names
    expect(screen.getByText("action-1")).toBeInTheDocument();
    expect(screen.getByText("action-2")).toBeInTheDocument();
  });
});
