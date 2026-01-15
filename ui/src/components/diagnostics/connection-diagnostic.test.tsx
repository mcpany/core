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
  httpService: {
    address: "https://example.com",
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
                { id: "test-service", name: "Test Service", status: "healthy", message: "" }
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
});
