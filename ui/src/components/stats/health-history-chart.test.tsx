/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor, act } from "@testing-library/react";
import { HealthHistoryChart } from "./health-history-chart";
import { vi, describe, it, expect, beforeEach } from "vitest";
import { apiClient } from "@/lib/client";

// Mock the apiClient
vi.mock("@/lib/client", () => ({
  apiClient: {
    getDashboardTraffic: vi.fn(),
  },
}));

// Mock Recharts responsive container to render immediately
vi.mock("recharts", async () => {
    const OriginalModule = await vi.importActual("recharts");
    return {
        ...OriginalModule,
        ResponsiveContainer: ({ children }: { children: any }) => (
            <div style={{ width: 800, height: 800 }}>{children}</div>
        ),
    };
});


describe("HealthHistoryChart", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders correctly with healthy data", async () => {
    const mockData = [
      { time: "10:00", requests: 100, errors: 0 },
      { time: "10:01", requests: 50, errors: 0 },
    ];
    (apiClient.getDashboardTraffic as any).mockResolvedValue(mockData);

    render(<HealthHistoryChart />);

    await waitFor(() => {
      // The overall success rate calculation should be present
      expect(screen.getByText(/100.0% Success Rate/i)).toBeInTheDocument();
    });

    // Check if title is present
    expect(screen.getByText("Traffic & Health (Last Hour)")).toBeInTheDocument();
    expect(screen.getByText("Availability based on request success rate.")).toBeInTheDocument();
  });

  it("renders correctly with degraded data", async () => {
    const mockData = [
      { time: "10:00", requests: 100, errors: 15 }, // 85% availability -> Critical
    ];
    (apiClient.getDashboardTraffic as any).mockResolvedValue(mockData);

    render(<HealthHistoryChart />);

    await waitFor(() => {
      // 85%
      expect(screen.getByText(/85.0% Success Rate/i)).toBeInTheDocument();
    });
  });

  it("renders correctly with empty data", async () => {
    (apiClient.getDashboardTraffic as any).mockResolvedValue([]);

    render(<HealthHistoryChart />);

    await waitFor(() => {
      expect(screen.getByText("No Data")).toBeInTheDocument();
    });
  });

  it("polls for data updates", async () => {
    vi.useFakeTimers();
    (apiClient.getDashboardTraffic as any).mockResolvedValue([]);

    render(<HealthHistoryChart />);

    expect(apiClient.getDashboardTraffic).toHaveBeenCalledTimes(1);

    // Fast forward 30 seconds
    await act(async () => {
        vi.advanceTimersByTime(30000);
    });

    expect(apiClient.getDashboardTraffic).toHaveBeenCalledTimes(2);

    vi.useRealTimers();
  });
});
