/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { DashboardGrid } from "./dashboard-grid";
import { vi, describe, it, expect, beforeEach } from "vitest";

// Mock Child Widgets
vi.mock("@/components/dashboard/metrics-overview", () => ({
  MetricsOverview: () => <div data-testid="widget-metrics">Metrics Overview Widget</div>
}));
vi.mock("@/components/dashboard/service-health-widget", () => ({
  ServiceHealthWidget: () => <div data-testid="widget-service-health">Service Health Widget</div>
}));
vi.mock("@/components/dashboard/lazy-charts", () => ({
  LazyRequestVolumeChart: () => <div data-testid="widget-request-volume">Request Volume Widget</div>,
  LazyTopToolsWidget: () => <div data-testid="widget-top-tools">Top Tools Widget</div>,
  LazyHealthHistoryChart: () => <div data-testid="widget-uptime">System Uptime Widget</div>,
  LazyRecentActivityWidget: () => <div data-testid="widget-recent-activity">Recent Activity Widget</div>
}));
vi.mock("@/components/dashboard/tool-failure-rate-widget", () => ({
  ToolFailureRateWidget: () => <div data-testid="widget-failure-rate">Tool Failure Rate Widget</div>
}));

// Mock Drag and Drop
vi.mock("@hello-pangea/dnd", () => ({
  DragDropContext: ({ children }: any) => <div>{children}</div>,
  Droppable: ({ children }: any) => children({ droppableProps: {}, innerRef: null, placeholder: null }),
  Draggable: ({ children }: any) => children({ draggableProps: {}, dragHandleProps: {}, innerRef: null }, { isDragging: false }),
}));

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe("DashboardGrid", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("renders all default widgets initially", () => {
    render(<DashboardGrid />);
    expect(screen.getByTestId("widget-metrics")).toBeInTheDocument();
    expect(screen.getByTestId("widget-recent-activity")).toBeInTheDocument();
    expect(screen.getByTestId("widget-uptime")).toBeInTheDocument();
  });

  it("loads layout from localStorage", () => {
    const savedLayout = [
        { id: "metrics", title: "Metrics Overview", type: "full", hidden: true }, // Hidden
        { id: "recent-activity", title: "Recent Activity", type: "half", hidden: false }
    ];
    localStorage.setItem("dashboard-layout", JSON.stringify(savedLayout));

    render(<DashboardGrid />);

    // Metrics should be hidden (not rendered in the grid list)
    expect(screen.queryByTestId("widget-metrics")).not.toBeInTheDocument();
    expect(screen.getByTestId("widget-recent-activity")).toBeInTheDocument();
  });

  it("migrates old layout schema", () => {
    // Old schema: type="wide" (mapped to full), missing hidden
    const oldLayout = [
        { id: "metrics", title: "Metrics Overview", type: "wide" }
    ];
    localStorage.setItem("dashboard-layout", JSON.stringify(oldLayout));

    render(<DashboardGrid />);

    expect(screen.getByTestId("widget-metrics")).toBeInTheDocument();
    // Verify it updated localStorage with new schema (we can't check localStorage directly easily as it updates on DragEnd usually or specific actions, but state should reflect it)
  });

  it("opens customization menu", async () => {
    render(<DashboardGrid />);

    const customizeBtn = screen.getByText("Customize View");
    fireEvent.click(customizeBtn);

    expect(screen.getByText("Toggle Widgets")).toBeInTheDocument();
    expect(screen.getByText("Metrics Overview")).toBeInTheDocument();
  });

  it("toggles widget visibility via customization menu", async () => {
    render(<DashboardGrid />);

    // Initially visible
    expect(screen.getByTestId("widget-metrics")).toBeInTheDocument();

    // Open menu
    fireEvent.click(screen.getByText("Customize View"));

    // Toggle off
    // Note: In JSDOM, clicking the label usually triggers the checkbox
    const label = screen.getByText("Metrics Overview");
    fireEvent.click(label);

    // Should be hidden
    await waitFor(() => {
        expect(screen.queryByTestId("widget-metrics")).not.toBeInTheDocument();
    });

    // Toggle on
    fireEvent.click(label);

    await waitFor(() => {
        expect(screen.getByTestId("widget-metrics")).toBeInTheDocument();
    });
  });
});
