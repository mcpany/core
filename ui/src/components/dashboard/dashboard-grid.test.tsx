/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { DashboardGrid } from "./dashboard-grid";
import { vi, describe, it, expect, beforeEach } from "vitest";
import { DashboardDensityProvider } from "@/contexts/dashboard-density-context";

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
    render(
      <DashboardDensityProvider>
        <DashboardGrid />
      </DashboardDensityProvider>
    );
    expect(screen.getByTestId("widget-metrics")).toBeInTheDocument();
    expect(screen.getByTestId("widget-recent-activity")).toBeInTheDocument();
    expect(screen.getByTestId("widget-uptime")).toBeInTheDocument();
  });

  it("loads layout from localStorage", () => {
    // The component expects migrated "WidgetInstance" format mostly, but handles legacy.
    // Let's use the new format to match `WidgetInstance` interface in dashboard-grid.tsx
    // The key change is `id` -> `instanceId` and `type` is preserved.
    // But the migration logic handles `id` as `type` if `instanceId` is missing.
    const savedLayout = [
        { id: "metrics", title: "Metrics Overview", type: "metrics", hidden: true }, // Legacy-ish format being tested
        { id: "recent-activity", title: "Recent Activity", type: "recent-activity", hidden: false }
    ];
    localStorage.setItem("dashboard-layout", JSON.stringify(savedLayout));

    render(
      <DashboardDensityProvider>
        <DashboardGrid />
      </DashboardDensityProvider>
    );

    // Metrics should be hidden (not rendered in the grid list)
    expect(screen.queryByTestId("widget-metrics")).not.toBeInTheDocument();
    expect(screen.getByTestId("widget-recent-activity")).toBeInTheDocument();
  });

  it("migrates old layout schema", () => {
    // Old schema: type="wide" (mapped to full), missing hidden
    const oldLayout = [
        { id: "metrics", title: "Metrics Overview", type: "metrics" } // id used as type
    ];
    localStorage.setItem("dashboard-layout", JSON.stringify(oldLayout));

    render(
      <DashboardDensityProvider>
        <DashboardGrid />
      </DashboardDensityProvider>
    );

    expect(screen.getByTestId("widget-metrics")).toBeInTheDocument();
    // Verify it updated localStorage with new schema (we can't check localStorage directly easily as it updates on DragEnd usually or specific actions, but state should reflect it)
  });

  it("opens customization menu", async () => {
    render(
      <DashboardDensityProvider>
        <DashboardGrid />
      </DashboardDensityProvider>
    );

    const customizeBtn = screen.getByText("Layout");
    fireEvent.click(customizeBtn);

    expect(screen.getByText("Visible Widgets")).toBeInTheDocument();
    expect(screen.getByText("Metrics Overview")).toBeInTheDocument();
  });

  it("toggles widget visibility via customization menu", async () => {
    render(
      <DashboardDensityProvider>
        <DashboardGrid />
      </DashboardDensityProvider>
    );

    // Initially visible
    expect(screen.getByTestId("widget-metrics")).toBeInTheDocument();

    // Open menu
    fireEvent.click(screen.getByText("Layout"));

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
