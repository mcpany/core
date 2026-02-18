/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import { SystemStatusIndicator } from "@/components/system-status-indicator";
import { apiClient } from "@/lib/client";
import { vi, describe, it, expect, beforeEach } from "vitest";

// Mock the apiClient
vi.mock("@/lib/client", () => ({
  apiClient: {
    getDoctorStatus: vi.fn(),
  },
}));

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe("SystemStatusIndicator", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders healthy status correctly", async () => {
    (apiClient.getDoctorStatus as any).mockResolvedValue({
      status: "healthy",
      timestamp: new Date().toISOString(),
      checks: {},
    });

    render(<SystemStatusIndicator />);

    // Initial loading state or resolved state
    await waitFor(() => {
      expect(screen.getByText("Healthy")).toBeInTheDocument();
    });
  });

  it("renders degraded status correctly", async () => {
    (apiClient.getDoctorStatus as any).mockResolvedValue({
      status: "degraded",
      timestamp: new Date().toISOString(),
      checks: {
        database: { status: "degraded", message: "Slow query" },
      },
    });

    render(<SystemStatusIndicator />);

    await waitFor(() => {
      expect(screen.getByText("degraded")).toBeInTheDocument();
    });
  });

  it("opens sheet on click", async () => {
    (apiClient.getDoctorStatus as any).mockResolvedValue({
      status: "healthy",
      timestamp: new Date().toISOString(),
      checks: {},
    });

    render(<SystemStatusIndicator />);

    await waitFor(() => {
        expect(screen.getByTitle("System Status")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTitle("System Status"));

    await waitFor(() => {
        expect(screen.getByText("Real-time diagnostics and environment health check.")).toBeInTheDocument();
    });
  });
});
