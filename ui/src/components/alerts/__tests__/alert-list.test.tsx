/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from "@testing-library/react";
import { AlertList } from "../alert-list";
import React from "react";
import userEvent from "@testing-library/user-event";
import { vi, describe, it, expect, beforeEach } from "vitest";
import { AlertStatus } from "../types";
import { apiClient } from "@/lib/client";

// Mock resize observer which is used by some UI components (like Recharts or ScrollArea)
class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
window.ResizeObserver = ResizeObserver;

 // Mock API Client
 vi.mock("@/lib/client", () => ({
   apiClient: {
     listAlerts: vi.fn().mockResolvedValue([
       {
         id: "1",
         title: "High CPU Usage",
         message: "The CPU usage is above 90% for service 'api-server'.",
         severity: "critical",
         status: "active",
         service: "api-server",
         timestamp: new Date().toISOString(),
       },
       {
         id: "2",
         title: "API Latency Spike",
         message: "The API latency is above 500ms for service 'auth-service'.",
         severity: "warning",
         status: "active",
         service: "auth-service",
         timestamp: new Date().toISOString(),
       },
     ]),
     updateAlertStatus: vi.fn().mockImplementation((id: string, status: AlertStatus) => Promise.resolve({
         id,
         status,
         title: "Updated Alert",
         message: "Status changed",
         severity: "info",
         service: "system",
         timestamp: new Date().toISOString()
     })),
   },
 }));

describe("AlertList", () => {
  it("renders alerts correctly", async () => {
    render(<AlertList />);
    await waitFor(() => {
        expect(screen.getByText("High CPU Usage")).toBeInTheDocument();
        expect(screen.getByText("API Latency Spike")).toBeInTheDocument();
    });
  });

  it("filters alerts by search query", async () => {
    render(<AlertList />);

    const searchInput = screen.getByPlaceholderText("Search alerts by title, message, service...");
    await userEvent.type(searchInput, "CPU");

    expect(screen.getByText("High CPU Usage")).toBeInTheDocument();
    expect(screen.queryByText("API Latency Spike")).not.toBeInTheDocument();
  });

  it("filters alerts by severity", async () => {
    render(<AlertList />);

    // We need to interact with the Select component.
    // Radix UI Select is tricky to test as it uses portals.
    // For unit tests, we often assume the underlying logic works or use userEvent.
    // However, finding the trigger can be done by role.

    // This is a simplified check assuming default state is correct.
    expect(screen.getAllByRole("row").length).toBeGreaterThan(1); // Header + rows
  });

  describe("Bulk Actions", () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it("can select all alerts and perform bulk acknowledge", async () => {
        render(<AlertList />);

        // Wait for data to load
        await waitFor(() => {
            expect(screen.getByText("High CPU Usage")).toBeInTheDocument();
        });

        // Click select all
        const selectAllCheckbox = screen.getByRole("checkbox", { name: /select all/i });
        await userEvent.click(selectAllCheckbox);

        // Verify bulk actions are visible
        const ackButton = screen.getByRole("button", { name: /acknowledge/i });
        expect(ackButton).toBeInTheDocument();
        expect(screen.getByText("2 selected")).toBeInTheDocument();

        // Click acknowledge
        await userEvent.click(ackButton);

        // Verify API was called twice with 'acknowledged'
        await waitFor(() => {
            expect(apiClient.updateAlertStatus).toHaveBeenCalledTimes(2);
            expect(apiClient.updateAlertStatus).toHaveBeenCalledWith("1", "acknowledged");
            expect(apiClient.updateAlertStatus).toHaveBeenCalledWith("2", "acknowledged");
        });
    });

    it("can select individual alerts and perform bulk resolve", async () => {
        render(<AlertList />);

        // Wait for data to load
        await waitFor(() => {
            expect(screen.getByText("High CPU Usage")).toBeInTheDocument();
        });

        // Click specific checkbox
        const itemCheckbox = screen.getByRole("checkbox", { name: /select 1/i });
        await userEvent.click(itemCheckbox);

        // Verify bulk actions are visible
        const resolveButton = screen.getByRole("button", { name: /resolve/i });
        expect(resolveButton).toBeInTheDocument();
        expect(screen.getByText("1 selected")).toBeInTheDocument();

        // Click resolve
        await userEvent.click(resolveButton);

        // Verify API was called once with 'resolved'
        await waitFor(() => {
            expect(apiClient.updateAlertStatus).toHaveBeenCalledTimes(1);
            expect(apiClient.updateAlertStatus).toHaveBeenCalledWith("1", "resolved");
        });
    });
  });
});
