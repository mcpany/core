/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from "../../tests/test-utils";
import { RecentActivityWidget } from "./recent-activity-widget";
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";

import { apiClient } from "@/lib/client";

// Mock Trace data
const mockTraces = [
  {
    id: "trace-1",
    timestamp: new Date().toISOString(),
    totalDuration: 150,
    status: "success",
    rootSpan: {
      name: "POST /get_weather",
    },
  },
  {
    id: "trace-2",
    timestamp: new Date(Date.now() - 60000).toISOString(), // 1 min ago
    totalDuration: 1200,
    status: "error",
    rootSpan: {
      name: "GET /list_users",
    },
  },
];

vi.mock("@/lib/client", () => ({
  apiClient: {
    getTraces: vi.fn(),
  },
}));

describe("RecentActivityWidget", () => {
  beforeEach(() => {
    vi.mocked(apiClient.getTraces).mockReset();
  });

  it("renders loading state initially", async () => {
    // Return a promise that never resolves immediately to test loading state
    vi.mocked(apiClient.getTraces).mockReturnValue(new Promise(() => {}));

    render(<RecentActivityWidget />);
    expect(screen.getByText(/Loading timeline/i)).toBeInTheDocument();
  });

  it("renders traces when fetch succeeds", async () => {
    vi.mocked(apiClient.getTraces).mockResolvedValue(mockTraces as any);

    render(<RecentActivityWidget />);

    await waitFor(() => {
        expect(screen.getByText("get_weather")).toBeInTheDocument();
        expect(screen.getByText("list_users")).toBeInTheDocument();
    });

    // Check for success/error indicators (indirectly via text content or class presence if we query by role, but simple text check is good for now)
    expect(screen.getByText("Failed")).toBeInTheDocument(); // Trace 2 has failed badge
    expect(screen.getByText("150ms")).toBeInTheDocument();
    expect(screen.getByText("1200ms")).toBeInTheDocument();
  });

  it("renders empty state when no traces", async () => {
    vi.mocked(apiClient.getTraces).mockResolvedValue([]);

    render(<RecentActivityWidget />);

    await waitFor(() => {
        expect(screen.getByText("No recent activity recorded.")).toBeInTheDocument();
    });
  });

  it("renders error state when fetch fails", async () => {
    vi.mocked(apiClient.getTraces).mockRejectedValue(new Error("Failed"));

    render(<RecentActivityWidget />);

    await waitFor(() => {
        expect(screen.getByText("Failed to load activity.")).toBeInTheDocument();
    });
  });
});
