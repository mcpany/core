/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { ConnectionDiagnosticDialog } from "./connection-diagnostic";
import { UpstreamServiceConfig } from "@/lib/types";
import { vi } from "vitest";

const mockService: UpstreamServiceConfig = {
  id: "test-service",
  name: "Test Service",
  version: "1.0.0",
  disable: false,
  priority: 0,
  loadBalancingStrategy: 0,
  sanitizedName: "test-service",
  callPolicies: [],
  preCallHooks: [],
  postCallHooks: [],
  prompts: [],
  autoDiscoverTool: false,
  configError: "",
  tags: [],
  httpService: {
    address: "https://example.com",
    tools: [],
    resources: [],
    prompts: [],
    calls: {},
  }
};

const mockWebSocketService: UpstreamServiceConfig = {
    id: "ws-service",
    name: "WebSocket Service",
    version: "1.0.0",
    disable: false,
    priority: 0,
    loadBalancingStrategy: 0,
    sanitizedName: "ws-service",
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    prompts: [],
    autoDiscoverTool: false,
    configError: "",
    tags: [],
    websocketService: {
      address: "ws://example.com",
      tools: [],
      resources: [],
      prompts: [],
      calls: {},
    }
  };

describe("ConnectionDiagnosticDialog", () => {
  beforeEach(() => {
    // Mock global fetch
    global.fetch = vi.fn(() =>
        Promise.resolve({
            ok: true,
            json: () => Promise.resolve([
                { id: "test-service", name: "Test Service", status: "healthy", message: "" },
                { id: "ws-service", name: "WebSocket Service", status: "healthy", message: "" }
            ]),
        })
    ) as any;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("renders the trigger button", () => {
    render(<ConnectionDiagnosticDialog service={mockService} />);
    expect(screen.getByText("Troubleshoot")).toBeInTheDocument();
  });

  it("opens the dialog and starts diagnostics", async () => {
    render(<ConnectionDiagnosticDialog service={mockService} />);

    const trigger = screen.getByText("Troubleshoot");
    fireEvent.click(trigger);

    expect(screen.getByText("Connection Diagnostics")).toBeInTheDocument();

    const startButton = screen.getByText("Start Diagnostics");
    fireEvent.click(startButton);

    expect(screen.getByText("Running...")).toBeInTheDocument();

    // Check if steps appear
    await waitFor(() => {
        expect(screen.getByText("Client-Side Configuration Check")).toBeInTheDocument();
    });

    // Wait for completion
    await waitFor(() => {
        expect(screen.getByText("Rerun Diagnostics")).toBeInTheDocument();
    }, { timeout: 5000 });

    expect(screen.getByText("Configuration valid")).toBeInTheDocument();

    // Check if backend health check was successful
    expect(global.fetch).toHaveBeenCalledWith("/api/dashboard/health", expect.any(Object));
    expect(screen.getByText("Connected")).toBeInTheDocument();
  });

  it("detects WebSocket service and adds browser check step", async () => {
    // We try to mock WebSocket just to prevent errors, but we won't assert on it heavily
    // since JSDOM mocking is flaky.
    const MockWebSocket = vi.fn().mockImplementation(() => {
        return {
            close: vi.fn(),
            onopen: null,
            onerror: null,
        };
    });
    vi.stubGlobal('WebSocket', MockWebSocket);
    if (typeof window !== 'undefined') {
        try {
            Object.defineProperty(window, 'WebSocket', {
                value: MockWebSocket,
                writable: true,
            });
        } catch (e) {
            // Ignore if we can't redefine
        }
    }

    render(<ConnectionDiagnosticDialog service={mockWebSocketService} />);

    const trigger = screen.getByText("Troubleshoot");
    fireEvent.click(trigger);

    const startButton = screen.getByText("Start Diagnostics");
    fireEvent.click(startButton);

    // Wait for the simulated UI delay
    await waitFor(() => {
        expect(screen.getByText("Client-Side Configuration Check")).toBeInTheDocument();
    });

    // Verify that the Browser Connectivity Check step is present
    await waitFor(() => {
        expect(screen.getByText("Browser Connectivity Check")).toBeInTheDocument();
    });
  });
});
