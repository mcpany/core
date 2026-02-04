/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, act } from "@testing-library/react";
import { LogStream } from "./log-stream";
import { vi } from "vitest";

// Mock next/dynamic to handle Virtuoso and JsonViewer
vi.mock("next/dynamic", () => ({
  default: () => {
    const MockDynamic = (props: any) => {
        // Mock Virtuoso behavior
        if (props.data && props.itemContent) {
             return (
                <div data-testid="virtuoso-mock">
                  {props.data.map((item: any, index: number) => props.itemContent(index, item))}
                </div>
              );
        }
        // Mock JsonViewer behavior
        if (props.data) {
            return <pre>{JSON.stringify(props.data, null, 2)}</pre>
        }
        return null;
    }
    return MockDynamic;
  }
}));

// Mock next/navigation
vi.mock("next/navigation", () => ({
    useSearchParams: () => ({
        get: (key: string) => {
            if (key === "source") return null;
            return null;
        }
    })
}));

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
        vi.advanceTimersByTime(500);
    });

    expect(screen.getByText("First Log")).toBeInTheDocument();

    // Find the Pause button.
    const pauseButton = screen.getAllByText(/Pause/i)[0];
    fireEvent.click(pauseButton);

    // Verify button text changes to Resume
    expect(screen.getAllByText(/Resume/i)[0]).toBeInTheDocument();

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
        vi.advanceTimersByTime(500);
    });

    // Should NOT be in document
    expect(screen.queryByText("Second Log")).not.toBeInTheDocument();

    // Click Resume
    const resumeButton = screen.getAllByText(/Resume/i)[0];
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
        vi.advanceTimersByTime(500);
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
        vi.advanceTimersByTime(500);
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

  it("detects and renders JSON logs", async () => {
    vi.useFakeTimers();
    render(<LogStream />);

    act(() => {
      if (mockWebSocket.onopen) mockWebSocket.onopen();
    });

    const jsonMessage = { foo: "bar", nested: { val: 123 } };
    const jsonLog = {
      id: "json-1",
      timestamp: new Date().toISOString(),
      level: "INFO",
      message: JSON.stringify(jsonMessage),
      source: "json-test"
    };

    act(() => {
      if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(jsonLog) });
    });

    act(() => {
        vi.advanceTimersByTime(500);
    });

    // Check if the expand button exists
    const expandButton = screen.getByLabelText("Expand JSON");
    expect(expandButton).toBeInTheDocument();

    // The raw message string should be visible
    expect(screen.getByText(JSON.stringify(jsonMessage))).toBeInTheDocument();

    // Click expand
    fireEvent.click(expandButton);

    // Verify state changed to expanded (button label changes)
    const collapseButton = screen.getByLabelText("Collapse JSON");
    expect(collapseButton).toBeInTheDocument();

    vi.useRealTimers();
  });

  it("highlights search terms in log messages", async () => {
    vi.useFakeTimers();
    render(<LogStream />);

    act(() => {
      if (mockWebSocket.onopen) mockWebSocket.onopen();
    });

    const log = {
      id: "search-1",
      timestamp: new Date().toISOString(),
      level: "INFO",
      message: "An error occurred in the system",
      source: "backend"
    };

    act(() => {
      if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(log) });
    });

    act(() => {
        vi.advanceTimersByTime(500);
    });

    // Enter search term
    const searchInput = screen.getByPlaceholderText("Search logs...");
    fireEvent.change(searchInput, { target: { value: "error" } });

    // Advance time to allow for any deferred updates (useDeferredValue)
    act(() => {
        vi.advanceTimersByTime(500);
    });

    // We expect the word "error" to be wrapped in a <mark> tag
    const highlighted = screen.getByText("error");
    expect(highlighted.tagName).toBe("MARK");
    expect(highlighted).toHaveClass("bg-yellow-500/40");

    vi.useRealTimers();
  });

  it("gracefully handles invalid JSON that passes heuristic", async () => {
    vi.useFakeTimers();
    render(<LogStream />);

    act(() => {
      if (mockWebSocket.onopen) mockWebSocket.onopen();
    });

    const invalidJsonMessage = "{ this is not valid json }";
    const log = {
      id: "invalid-json-1",
      timestamp: new Date().toISOString(),
      level: "INFO",
      message: invalidJsonMessage,
      source: "test"
    };

    act(() => {
      if (mockWebSocket.onmessage) mockWebSocket.onmessage({ data: JSON.stringify(log) });
    });

    act(() => {
        vi.advanceTimersByTime(500);
    });

    // Check if the expand button exists (heuristic should pass)
    const expandButton = screen.getByLabelText("Expand JSON");
    expect(expandButton).toBeInTheDocument();

    // Click expand
    fireEvent.click(expandButton);

    // Should show "Invalid JSON" message
    expect(screen.getByText("Invalid JSON")).toBeInTheDocument();

    vi.useRealTimers();
  });
});
