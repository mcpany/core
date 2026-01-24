/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ToolInspector } from "./tool-inspector";
import { ToolDefinition, apiClient } from "@/lib/client";
import { vi, describe, it, expect, beforeEach } from "vitest";

// Mock apiClient
vi.mock("@/lib/client", async () => {
  const actual = await vi.importActual("@/lib/client");
  return {
    ...actual,
    apiClient: {
      executeTool: vi.fn(),
    },
  };
});

// Mock recharts to avoid complex SVG rendering issues in test
vi.mock("recharts", () => {
  return {
    ResponsiveContainer: ({ children }: any) => <div className="recharts-responsive-container">{children}</div>,
    AreaChart: () => <div>AreaChart</div>,
    Area: () => <div>Area</div>,
    XAxis: () => <div>XAxis</div>,
    YAxis: () => <div>YAxis</div>,
    CartesianGrid: () => <div>CartesianGrid</div>,
    Tooltip: () => <div>Tooltip</div>,
  };
});

// Mock SchemaViewer to avoid rendering issues
vi.mock("./schema-viewer", () => ({
    SchemaViewer: () => <div>SchemaViewer</div>
}));

describe("ToolInspector", () => {
  const mockTool: ToolDefinition = {
    name: "test-tool",
    description: "A test tool",
    serviceId: "test-service",
    inputSchema: { type: "object", properties: { arg: { type: "string" } } },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it("renders correctly", () => {
    render(<ToolInspector tool={mockTool} open={true} onOpenChange={() => {}} />);
    expect(screen.getByText("test-tool")).toBeInTheDocument();
    expect(screen.getByText("test-service")).toBeInTheDocument();
  });

  it("executes tool and persists history", async () => {
    const user = userEvent.setup();
    (apiClient.executeTool as any).mockResolvedValue({ result: "success" });

    render(<ToolInspector tool={mockTool} open={true} onOpenChange={() => {}} />);

    // Click execute
    const executeBtn = screen.getByText("Execute");
    await user.click(executeBtn);

    await waitFor(() => {
      expect(apiClient.executeTool).toHaveBeenCalled();
    });

    // Check if result is displayed
    expect(screen.getByText(/"result": "success"/)).toBeInTheDocument();

    // Check localStorage
    const stored = localStorage.getItem("tool-history-test-service-test-tool");
    expect(stored).not.toBeNull();
    const history = JSON.parse(stored!);
    expect(history).toHaveLength(1);
    expect(history[0].status).toBe("success");
  });

  it("clears history", async () => {
    const user = userEvent.setup();
    // Seed localStorage
    const history = [{ time: "10:00:00 AM", latency: 100, status: "success" }];
    localStorage.setItem("tool-history-test-service-test-tool", JSON.stringify(history));

    render(<ToolInspector tool={mockTool} open={true} onOpenChange={() => {}} />);

    // Go to Metrics tab
    const metricsTab = screen.getByText("Performance & Analytics");
    await user.click(metricsTab);

    // Check if history is displayed (via Total Calls stat)
    // We expect "1" to be present in the document (Total Calls)
    await waitFor(() => expect(screen.getByText("1")).toBeInTheDocument());

    // Click Clear History
    const clearBtn = screen.getByText("Clear History");
    await user.click(clearBtn);

    // Verify localStorage is empty
    const stored = localStorage.getItem("tool-history-test-service-test-tool");
    expect(JSON.parse(stored!)).toEqual([]);

    // Verify stats updated - Total Calls should be 0
    // "Total Calls" label is there, and "0" should be the value
    // Since "0" might appear multiple times (avg latency, error count), we use getAllByText
    expect(screen.getAllByText("0").length).toBeGreaterThan(0);
  });
});
