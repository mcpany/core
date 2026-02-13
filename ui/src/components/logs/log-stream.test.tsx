/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, act, waitFor } from "@testing-library/react";
import { LogStream } from "./log-stream";
import { vi } from "vitest";

// Mock react-virtuoso
vi.mock("react-virtuoso", () => ({
  Virtuoso: ({ data, itemContent }: any) => (
    <div data-testid="virtuoso-mock">
      {data.map((item: any, index: number) => itemContent(index, item))}
    </div>
  ),
  VirtuosoHandle: {},
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

    await waitFor(() => expect(screen.getByText("First Log")).toBeInTheDocument());

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

    // Wait a bit to ensure it DOESN'T appear (waitFor would fail if we expect NOT, so we check queryByText)
    // We can't wait for "not to happen" easily.
    // We can wait for 200ms then check.
    await new Promise(r => setTimeout(r, 200));

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

    await waitFor(() => expect(screen.getByText("Third Log")).toBeInTheDocument());
  });

  it("filters logs by source", async () => {
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

    // Verify both logs are present initially
    await waitFor(() => {
        expect(screen.getByText("Log from Service A")).toBeInTheDocument();
        expect(screen.getByText("Log from Service B")).toBeInTheDocument();
    });

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
    await waitFor(() => {
        expect(screen.getByText("Log from Service A")).toBeInTheDocument();
        expect(screen.queryByText("Log from Service B")).not.toBeInTheDocument();
    });
  });

  it("detects and renders JSON logs", async () => {
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

    // Check if the expand button exists
    const expandButton = await waitFor(() => screen.getByLabelText("Expand JSON"));
    expect(expandButton).toBeInTheDocument();

    // The raw message string should be visible
    expect(screen.getByText(JSON.stringify(jsonMessage))).toBeInTheDocument();

    // Click expand
    fireEvent.click(expandButton);

    // Verify state changed to expanded (button label changes)
    await waitFor(() => {
        const collapseButton = screen.getByLabelText("Collapse JSON");
        expect(collapseButton).toBeInTheDocument();
    });
  });

  it("highlights search terms in log messages", async () => {
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

    await waitFor(() => expect(screen.getByText("An error occurred in the system")).toBeInTheDocument());

    // Enter search term
    const searchInput = screen.getByPlaceholderText("Search logs...");
    fireEvent.change(searchInput, { target: { value: "error" } });

    // We expect the word "error" to be wrapped in a <mark> tag
    await waitFor(() => {
        const highlighted = screen.getByText("error");
        expect(highlighted.tagName).toBe("MARK");
        expect(highlighted).toHaveClass("bg-yellow-500/40");
    });
  });

  it("gracefully handles invalid JSON that passes heuristic", async () => {
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

    // Check if the expand button exists (heuristic should pass)
    const expandButton = await waitFor(() => screen.getByLabelText("Expand JSON"));
    expect(expandButton).toBeInTheDocument();

    // Click expand
    fireEvent.click(expandButton);

    // Should show "Invalid JSON" message
    await waitFor(() => expect(screen.getByText("Invalid JSON")).toBeInTheDocument());
  });
});
