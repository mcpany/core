/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render } from "@testing-library/react";
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

    // Verify WebSocket connection
    // We expect it to construct the URL based on window.location
    // In test environment, window.location might be http://localhost:3000
    // So we check for the suffix
    expect(global.WebSocket).toHaveBeenCalledWith(expect.stringContaining("/api/v1/ws/logs"));
  });
});
