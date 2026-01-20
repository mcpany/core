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
});
