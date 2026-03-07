/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { ToolDetail } from "./tool-detail";
import { apiClient } from "@/lib/client";

// Mock the apiClient
vi.mock("@/lib/client", () => ({
  apiClient: {
    getService: vi.fn(),
    getServiceStatus: vi.fn(),
    getToolUsage: vi.fn(),
  },
}));

// Mock the components
vi.mock("./service-property-card", () => ({
  ServicePropertyCard: () => <div data-testid="service-property-card" />,
}));

vi.mock("./schema-visualizer", () => ({
  SchemaVisualizer: () => <div data-testid="schema-visualizer" />,
}));

// Mock lucide-react
vi.mock("lucide-react", () => ({
  Wrench: () => <div data-testid="wrench-icon" />,
  AlertTriangle: () => <div data-testid="alert-icon" />,
  TrendingUp: () => <div data-testid="trending-icon" />,
  Braces: () => <div data-testid="braces-icon" />,
}));

// Mock useToast
vi.mock("@/hooks/use-toast", () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

describe("ToolDetail", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders correctly with full data including usage metrics", async () => {
    // Mock getService
    vi.mocked(apiClient.getService).mockResolvedValue({
      service: {
        id: "service1",
        name: "TestService",
        httpService: {
          tools: [
            {
              name: "TestTool",
              description: "A test tool",
              inputSchema: {},
            },
          ],
        },
      },
    });

    // Mock getServiceStatus
    vi.mocked(apiClient.getServiceStatus).mockResolvedValue({
      metrics: {
        "tool_usage:TestTool": 42,
      },
    });

    // Mock getToolUsage
    vi.mocked(apiClient.getToolUsage).mockResolvedValue([
      {
        name: "TestTool",
        serviceId: "service1",
        successRate: 0.95,
        avgLatencyMs: 150,
        errorCount: 2,
      },
    ]);

    render(<ToolDetail serviceId="service1" toolName="TestTool" />);

    // Wait for the async details to load and assert everything inside waitFor to avoid race conditions
    await waitFor(() => {
      expect(screen.getByText("TestTool")).toBeInTheDocument();

      // Assert total calls from getServiceStatus
      expect(screen.getByText("Total Calls")).toBeInTheDocument();
      expect(screen.getByText("42")).toBeInTheDocument();

      // Assert new metrics from getToolUsage
      expect(screen.getByText("Success Rate")).toBeInTheDocument();
      expect(screen.getByText("95.0%")).toBeInTheDocument();

      expect(screen.getByText("Avg Latency")).toBeInTheDocument();
      expect(screen.getByText("150 ms")).toBeInTheDocument();

      expect(screen.getByText("Error Count")).toBeInTheDocument();
      expect(screen.getByText("2")).toBeInTheDocument();
    }, { timeout: 5000 });
  });
});
