/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from "@testing-library/react";
import { RecentActivityWidget } from "./recent-activity-widget";
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";

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

// Mock Fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe("RecentActivityWidget", () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it("renders loading state initially", async () => {
    // Return a promise that never resolves immediately to test loading state
    mockFetch.mockReturnValue(new Promise(() => {}));

    render(<RecentActivityWidget />);
    expect(screen.getByText(/Loading activity/i)).toBeInTheDocument();
  });

  it("renders traces when fetch succeeds", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => mockTraces,
    });

    render(<RecentActivityWidget />);

    await waitFor(() => {
        expect(screen.getByText("get_weather")).toBeInTheDocument();
        expect(screen.getByText("list_users")).toBeInTheDocument();
    });

    expect(mockFetch).toHaveBeenCalledWith('/api/traces?limit=5');

    // Check for success/error indicators (indirectly via text content or class presence if we query by role, but simple text check is good for now)
    expect(screen.getByText("Failed")).toBeInTheDocument(); // Trace 2 has failed badge
    expect(screen.getByText("150ms")).toBeInTheDocument();
    expect(screen.getByText("1200ms")).toBeInTheDocument();
  });

  it("renders empty state when no traces", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => [],
    });

    render(<RecentActivityWidget />);

    await waitFor(() => {
        expect(screen.getByText("No recent activity recorded.")).toBeInTheDocument();
    });
  });

  it("renders error state when fetch fails", async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
    });

    render(<RecentActivityWidget />);

    await waitFor(() => {
        expect(screen.getByText("Failed to load activity.")).toBeInTheDocument();
    });
  });
});
