/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { SystemStatusBanner } from "./system-status-banner";
import { apiClient } from "@/lib/client";
import { vi, describe, it, expect, afterEach } from "vitest";

// Mock apiClient
vi.mock("@/lib/client", () => ({
  apiClient: {
    getDoctorStatus: vi.fn(),
  },
}));

// Mock ResizeObserver for Dialog
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe("SystemStatusBanner", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("does not render when system is healthy", async () => {
    (apiClient.getDoctorStatus as any).mockResolvedValue({
      status: "healthy",
      checks: {
        database: { status: "ok" },
      },
    });

    const { container } = render(<SystemStatusBanner />);

    // Wait for effect to resolve
    await waitFor(() => {
        expect(apiClient.getDoctorStatus).toHaveBeenCalled();
    });

    expect(container).toBeEmptyDOMElement();
  });

  it("renders critical status when configuration fails", async () => {
    (apiClient.getDoctorStatus as any).mockResolvedValue({
      status: "unhealthy",
      checks: {
        configuration: {
          status: "error",
          message: "Failed to parse config",
          diff: "- old_value\n+ new_value",
        },
      },
    });

    render(<SystemStatusBanner />);

    await waitFor(() => {
      expect(screen.getByText(/System Status: Critical/i)).toBeInTheDocument();
    });

    // Check primary issue preview
    expect(screen.getByText(/Configuration: Failed to parse config/i)).toBeInTheDocument();

    // Open details
    fireEvent.click(screen.getByText(/View Details/i));

    await waitFor(() => {
      expect(screen.getByText("System Diagnostics")).toBeInTheDocument();
      expect(screen.getByText("Configuration Changes")).toBeInTheDocument();
      expect(screen.getByText("- old_value")).toBeInTheDocument();
      expect(screen.getByText("+ new_value")).toBeInTheDocument();
    });
  });

  it("renders connection error", async () => {
    (apiClient.getDoctorStatus as any).mockRejectedValue(new Error("Network Error"));

    render(<SystemStatusBanner />);

    await waitFor(() => {
      expect(screen.getByText(/System Status: Critical/i)).toBeInTheDocument();
      expect(screen.getByText(/Connection Error/i)).toBeInTheDocument();
    });
  });
});
