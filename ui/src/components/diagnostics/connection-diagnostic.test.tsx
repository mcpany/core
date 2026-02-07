/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { ConnectionDiagnosticDialog } from "./connection-diagnostic";
import { UpstreamServiceConfig } from "@/lib/types";
import { apiClient } from "@/lib/client";
import { vi } from "vitest";

vi.mock("@/lib/client", () => ({
    apiClient: {
        getService: vi.fn(),
    },
}));

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
    // Default mock for getService (Operational Verification)
    (apiClient.getService as any).mockResolvedValue({
        service: {
            toolCount: 5,
            lastError: ""
        }
    });

    // Default mock global fetch (Success case)
    global.fetch = vi.fn((url: string | Request, _init?: RequestInit) => {
        if (typeof url === 'string' && url.includes("/api/dashboard/health")) {
            return Promise.resolve({
                ok: true,
                json: () => Promise.resolve([
                    { id: "test-service", name: "Test Service", status: "healthy", message: "" },
                    { id: "ws-service", name: "WebSocket Service", status: "healthy", message: "" }
                ]),
            });
        }
        // Mock for Browser Connectivity Check (HTTP Service)
        if (typeof url === 'string' && (url.startsWith("http") || url.startsWith("https"))) {
             return Promise.resolve({
                ok: false, // opaque response in no-cors usually
                type: 'opaque',
                status: 0,
                json: () => Promise.reject("Opaque response"),
            });
        }
        return Promise.reject("Unknown URL");
    }) as unknown as typeof fetch;
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

    // Check Operational Verification
    expect(screen.getByText("Operational Verification")).toBeInTheDocument();
    await waitFor(() => {
         expect(screen.getByText("Fully Operational")).toBeInTheDocument();
    });
  });

  it("detects HTTP service and adds browser check step", async () => {
      render(<ConnectionDiagnosticDialog service={mockService} />);

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

      // Verify success log
      await waitFor(() => {
           expect(screen.getByText(/Successfully connected to HTTP server from browser/)).toBeInTheDocument();
      });
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
        } catch (_e) {
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

  it("displays context-aware suggestion for 404 error", async () => {
      // Mock failure response with 404 error
      global.fetch = vi.fn(() =>
        Promise.resolve({
            ok: true,
            json: () => Promise.resolve([
                { id: "test-service", name: "Test Service", status: "unhealthy", message: "404 Not Found" }
            ]),
        })
      ) as unknown as typeof fetch;

      render(<ConnectionDiagnosticDialog service={mockService} />);

      const trigger = screen.getByText("Troubleshoot");
      fireEvent.click(trigger);

      const startButton = screen.getByText("Start Diagnostics");
      fireEvent.click(startButton);

      // Wait for completion
      await waitFor(() => {
          expect(screen.getByText("Rerun Diagnostics")).toBeInTheDocument();
      }, { timeout: 5000 });

      // Check for suggestion card elements
      expect(screen.getByText("Not Found (404)")).toBeInTheDocument();
      expect(screen.getByText("The requested path or resource does not exist on the upstream server.")).toBeInTheDocument();
      // Use getAllByText because it appears in logs and the card
      expect(screen.getAllByText(/Check the URL path in your configuration/).length).toBeGreaterThan(0);
  });

  it("displays context-aware suggestion for Connection Refused", async () => {
    // Mock failure response with connection refused
    global.fetch = vi.fn(() =>
      Promise.resolve({
          ok: true,
          json: () => Promise.resolve([
              { id: "test-service", name: "Test Service", status: "unhealthy", message: "dial tcp 127.0.0.1:8080: connect: connection refused" }
          ]),
      })
    ) as unknown as typeof fetch;

    render(<ConnectionDiagnosticDialog service={mockService} />);

    const trigger = screen.getByText("Troubleshoot");
    fireEvent.click(trigger);

    const startButton = screen.getByText("Start Diagnostics");
    fireEvent.click(startButton);

    // Wait for completion
    await waitFor(() => {
        expect(screen.getByText("Rerun Diagnostics")).toBeInTheDocument();
    }, { timeout: 5000 });

    // Check for suggestion card elements
    expect(screen.getByText("Connection Refused")).toBeInTheDocument();
    // Use getAllByText because it appears in logs and the card
    expect(screen.getAllByText(/Check if the upstream service is running/).length).toBeGreaterThan(0);
});

  it("warns when no tools are discovered", async () => {
        (apiClient.getService as any).mockResolvedValue({
            service: {
                toolCount: 0,
                lastError: ""
            }
        });

        render(<ConnectionDiagnosticDialog service={mockService} />);

        const trigger = screen.getByText("Troubleshoot");
        fireEvent.click(trigger);

        const startButton = screen.getByText("Start Diagnostics");
        fireEvent.click(startButton);

        await waitFor(() => {
            expect(screen.getByText("No Tools (Warning)")).toBeInTheDocument();
        }, { timeout: 5000 });

        expect(screen.getByText("No Tools Discovered")).toBeInTheDocument();
  });

  it("detects and analyzes ZodError in operational check", async () => {
        (apiClient.getService as any).mockResolvedValue({
            service: {
                toolCount: 0,
                lastError: "ZodError: Invalid input"
            }
        });

        render(<ConnectionDiagnosticDialog service={mockService} />);

        const trigger = screen.getByText("Troubleshoot");
        fireEvent.click(trigger);

        const startButton = screen.getByText("Start Diagnostics");
        fireEvent.click(startButton);

        await waitFor(() => {
            expect(screen.getByText("Operational Error")).toBeInTheDocument();
        }, { timeout: 5000 });

        expect(screen.getByText("Schema Validation Error")).toBeInTheDocument();
        expect(screen.getByText("The upstream server returned data that does not match the expected schema.")).toBeInTheDocument();
  });

  it("warns about localhost/Docker usage when connection fails", async () => {
    // Mock fetch to throw error for localhost
    const originalFetch = global.fetch;
    global.fetch = vi.fn((url: string | Request, init?: RequestInit) => {
        if (typeof url === 'string' && url.includes("localhost")) {
            return Promise.reject(new TypeError("Failed to fetch"));
        }
        return originalFetch(url, init);
    }) as unknown as typeof fetch;

    // Mock localhost service
    const localhostService = {
        ...mockService,
        httpService: {
            address: "http://localhost:3000",
            tools: [],
            resources: [],
            prompts: [],
            calls: {},
        }
    };

    render(<ConnectionDiagnosticDialog service={localhostService} />);

    const trigger = screen.getByText("Troubleshoot");
    fireEvent.click(trigger);

    const startButton = screen.getByText("Start Diagnostics");
    fireEvent.click(startButton);

    // Wait for failure
    await waitFor(() => {
        expect(screen.getByText("Not Accessible")).toBeInTheDocument();
    });

    // Check for the warning message in logs
    // We use getAllByText or check for partial text
    await waitFor(() => {
        expect(screen.getByText((content) => content.includes("WARNING: You are using 'localhost'"))).toBeInTheDocument();
    });
    expect(screen.getByText((content) => content.includes("host.docker.internal"))).toBeInTheDocument();

    // Restore fetch
    global.fetch = originalFetch;
  });
});
