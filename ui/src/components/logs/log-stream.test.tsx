/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, act } from "@testing-library/react";
import { LogStream } from "./log-stream";
import { vi } from "vitest";

describe("LogStream", () => {
  let mockWebSocket: any;

  beforeEach(() => {
    // Mock WebSocket
    mockWebSocket = {
      close: vi.fn(),
      send: vi.fn(),
      onopen: null,
      onmessage: null,
      onclose: null,
      onerror: null,
    };

    global.WebSocket = vi.fn(function() {
        return mockWebSocket;
    }) as any;
  });

  // Mock the Select component to simplify testing logic
  vi.mock("@/components/ui/select", () => ({
    Select: ({ value, onValueChange, children }: any) => (
      <select
        data-testid="mock-select"
        value={value}
        onChange={(e) => onValueChange(e.target.value)}
      >
        {children}
      </select>
    ),
    SelectTrigger: ({ children }: any) => <div>{children}</div>,
    SelectValue: ({ placeholder }: any) => <div>{placeholder}</div>,
    SelectContent: ({ children }: any) => <>{children}</>,
    SelectItem: ({ value, children }: any) => (
      <option value={value}>{children}</option>
    ),
  }));

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("connects to the correct WebSocket URL", () => {
    render(<LogStream />);
    expect(global.WebSocket).toHaveBeenCalledWith(expect.stringContaining("/api/v1/ws/logs"));
  });

  it("stops processing logs when paused", async () => {
    vi.useFakeTimers();
    render(<LogStream />);

    // Trigger connection open
    act(() => {
      if (mockWebSocket.onopen) mockWebSocket.onopen();
    });

    // Send a log message
    const log1 = {
      id: "1",
      timestamp: new Date().toISOString(),
      level: "INFO",
      message: "First Log",
      source: "test"
    };

    act(() => {
      if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(log1) });
    });

    // Advance time to flush buffer (100ms interval)
    act(() => {
        vi.advanceTimersByTime(200);
    });

    expect(screen.getByText("First Log")).toBeInTheDocument();

    // Find the Pause button.
    const pauseButton = screen.getByText(/Pause/i, { selector: 'button' });
    fireEvent.click(pauseButton);

    // Verify button text changes to Resume
    expect(screen.getByText(/Resume/i)).toBeInTheDocument();

    // Send another log message while paused
    const log2 = {
      id: "2",
      timestamp: new Date().toISOString(),
      level: "INFO",
      message: "Second Log",
      source: "test"
    };

    act(() => {
      if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(log2) });
    });

    // Advance time
    act(() => {
        vi.advanceTimersByTime(200);
    });

    // Should NOT be in document
    expect(screen.queryByText("Second Log")).not.toBeInTheDocument();

    // Click Resume
    const resumeButton = screen.getByText(/Resume/i, { selector: 'button' });
    fireEvent.click(resumeButton);

    // Send third log
    const log3 = {
      id: "3",
      timestamp: new Date().toISOString(),
      level: "INFO",
      message: "Third Log",
      source: "test"
    };

    act(() => {
      if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(log3) });
    });

    // Advance time
    act(() => {
        vi.advanceTimersByTime(200);
    });

    expect(screen.getByText("Third Log")).toBeInTheDocument();

    vi.useRealTimers();
  });

  it("filters logs by source", async () => {
    vi.useFakeTimers();
    render(<LogStream />);

    // Trigger connection open
    act(() => {
      if (mockWebSocket.onopen) mockWebSocket.onopen();
    });

    // Send logs from different sources
    const logA = {
      id: "A",
      timestamp: new Date().toISOString(),
      level: "INFO",
      message: "Log from Service A",
      source: "service-a"
    };

    const logB = {
      id: "B",
      timestamp: new Date().toISOString(),
      level: "INFO",
      message: "Log from Service B",
      source: "service-b"
    };

    act(() => {
      if (mockWebSocket.onmessage) {
          mockWebSocket.onmessage({ data: JSON.stringify(logA) });
          mockWebSocket.onmessage({ data: JSON.stringify(logB) });
      }
    });

    // Advance time to flush buffer
    act(() => {
        vi.advanceTimersByTime(200);
    });

    // Verify both logs are present initially
    expect(screen.getByText("Log from Service A")).toBeInTheDocument();
    expect(screen.getByText("Log from Service B")).toBeInTheDocument();

    vi.useRealTimers();

    // Find the Selects. There are two (Source and Level).
    // We can find the one that contains the "service-a" option.
    const selects = screen.getAllByTestId("mock-select");
    const sourceSelect = selects.find(select =>
      select.innerHTML.includes("service-a")
    );

    if (!sourceSelect) {
        throw new Error("Could not find Source select");
    }

    // Change value
    fireEvent.change(sourceSelect, { target: { value: "service-a" } });

    // Verify filtering
    expect(screen.getByText("Log from Service A")).toBeInTheDocument();
    expect(screen.queryByText("Log from Service B")).not.toBeInTheDocument();
  });
});
