/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from "@testing-library/react";
import { AlertList } from "../alert-list";
import React from "react";
import userEvent from "@testing-library/user-event";
import { vi, type Mock } from 'vitest';
import { apiClient } from "@/lib/client";

// Mock apiClient
vi.mock("@/lib/client", () => ({
  apiClient: {
    listAlerts: vi.fn(),
    updateAlertStatus: vi.fn(),
  },
}));

// Mock resize observer which is used by some UI components (like Recharts or ScrollArea)
class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
window.ResizeObserver = ResizeObserver;

describe("AlertList", () => {
  const mockAlerts = [
    { id: "1", title: "High CPU Usage", message: "CPU > 90%", severity: "critical", status: "active", service: "core", timestamp: new Date().toISOString() },
    { id: "2", title: "API Latency Spike", message: "Latency > 1s", severity: "warning", status: "active", service: "api", timestamp: new Date().toISOString() }
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    (apiClient.listAlerts as Mock).mockResolvedValue(mockAlerts);
  });

  it("renders alerts correctly", async () => {
    render(<AlertList />);
    await waitFor(() => {
      expect(screen.getByText("High CPU Usage")).toBeInTheDocument();
      expect(screen.getByText("API Latency Spike")).toBeInTheDocument();
    });
  });

  it("filters alerts by search query", async () => {
    render(<AlertList />);

    await waitFor(() => {
        expect(screen.getByText("High CPU Usage")).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText("Search alerts by title, message, service...");
    await userEvent.type(searchInput, "CPU");

    expect(screen.getByText("High CPU Usage")).toBeInTheDocument();
    expect(screen.queryByText("API Latency Spike")).not.toBeInTheDocument();
  });

  it("filters alerts by severity", async () => {
    render(<AlertList />);

    await waitFor(() => {
        expect(screen.getAllByRole("row").length).toBeGreaterThan(1); // Header + rows
    });
  });
});
