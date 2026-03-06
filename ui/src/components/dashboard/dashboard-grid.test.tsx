/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor, act } from "@testing-library/react";
import { DashboardGrid } from "./dashboard-grid";
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";

// Mock next/navigation to fix invariant router error
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
  })
}));

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
  LazyRecentActivityWidget: () => <div data-testid="widget-recent-activity">Recent Activity Widget</div>,
  LazyAuditLogWidget: () => <div data-testid="widget-audit-log">Audit Log Widget</div>
}));
vi.mock("@/components/dashboard/tool-failure-rate-widget", () => ({
  ToolFailureRateWidget: () => <div data-testid="widget-failure-rate">Tool Failure Rate Widget</div>
}));
vi.mock("@/components/dashboard/network-graph-widget", () => ({
  NetworkGraphWidget: () => <div data-testid="widget-network-graph">Network Graph Widget</div>
}));

// Mock Drag and Drop
vi.mock("@hello-pangea/dnd", () => ({
  DragDropContext: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  Droppable: ({ children }: { children: (args: { droppableProps: object; innerRef: null; placeholder: null }) => React.ReactNode }) => children({ droppableProps: {}, innerRef: null, placeholder: null }),
  Draggable: ({ children }: { children: (args: { draggableProps: object; dragHandleProps: object; innerRef: null }, snapshot: { isDragging: boolean }) => React.ReactNode }) => children({ draggableProps: {}, dragHandleProps: {}, innerRef: null }, { isDragging: false }),
}));

vi.mock('@/components/dashboard/add-widget-sheet', () => ({
    AddWidgetSheet: ({ onAdd }: { onAdd: (type: string) => void }) => (
      <button onClick={() => onAdd('metrics')} data-testid="add-widget">
        Add Widget
      </button>
    ),
  }));

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe("DashboardGrid", () => {
    beforeEach(() => {
        vi.useFakeTimers();
        localStorage.clear();
      });

      afterEach(() => {
        vi.useRealTimers();
        vi.restoreAllMocks();
      });

  it("renders all default widgets initially", async () => {
    render(<DashboardGrid />);

    // We must resolve the microtasks to let the loadLayout async finish.
    await act(async () => {
        await vi.runAllTimersAsync();
    });
    expect(screen.getByTestId("widget-metrics")).toBeInTheDocument();
    expect(screen.getByTestId("widget-recent-activity")).toBeInTheDocument();
    expect(screen.getByTestId("widget-uptime")).toBeInTheDocument();
  });

  it("loads layout from localStorage", async () => {
    // Note: The DashboardGrid expects instanceId, but handles legacy format where id=type
    const savedLayout = [
        { instanceId: "1", type: "metrics", title: "Metrics Overview", size: "full", hidden: true }, // Hidden
        { instanceId: "2", type: "recent-activity", title: "Recent Activity", size: "half", hidden: false }
    ];
    localStorage.setItem("dashboard-layout", JSON.stringify(savedLayout));

    render(<DashboardGrid />);
    await act(async () => {
        await vi.runAllTimersAsync();
    });

    // Metrics should be hidden (not rendered in the grid list)
    expect(screen.queryByTestId("widget-metrics")).not.toBeInTheDocument();
    expect(screen.getByTestId("widget-recent-activity")).toBeInTheDocument();
  });

  it("migrates old layout schema", async () => {
    // Old schema: type="wide" (mapped to full), missing hidden
    const oldLayout = [
        { id: "metrics", title: "Metrics Overview", type: "wide" }
    ];
    localStorage.setItem("dashboard-layout", JSON.stringify(oldLayout));

    render(<DashboardGrid />);
    await act(async () => {
        await vi.runAllTimersAsync();
    });

    expect(screen.getByTestId("widget-metrics")).toBeInTheDocument();

    // Check local storage directly since we now trigger an immediate save in the load phase
    const updated = JSON.parse(localStorage.getItem("dashboard-layout") || "[]");
    expect(updated.length).toBeGreaterThan(0);
    expect(updated[0].instanceId).toBeDefined();
    expect(updated[0].type).toBe("metrics");
  });

  it("opens customization menu", async () => {
    render(<DashboardGrid />);
    await act(async () => {
        await vi.runAllTimersAsync();
    });

    const customizeBtn = screen.getByText("Layout");
    fireEvent.click(customizeBtn);

    expect(screen.getByText("Visible Widgets")).toBeInTheDocument();
    expect(screen.getByText("Metrics Overview")).toBeInTheDocument();
  });

  it("toggles widget visibility via customization menu", async () => {
    render(<DashboardGrid />);
    await act(async () => {
        await vi.runAllTimersAsync();
    });

    // Initially visible
    expect(screen.getByTestId("widget-metrics")).toBeInTheDocument();

    // Open menu
    fireEvent.click(screen.getByText("Layout"));

    // Toggle off
    const label = screen.getByText("Metrics Overview");
    await act(async () => {
        fireEvent.click(label);
    });

    // Should be hidden
    expect(screen.queryByTestId("widget-metrics")).not.toBeInTheDocument();

    // Toggle on
    await act(async () => {
        fireEvent.click(label);
    });

    expect(screen.getByTestId("widget-metrics")).toBeInTheDocument();
  });

  it("debounces localStorage writes", async () => {
    render(<DashboardGrid />);

    await act(async () => {
      await vi.runAllTimersAsync();
    });

    // Clear any previous writes from initialization/migration
    vi.spyOn(Storage.prototype, 'setItem');
    vi.mocked(localStorage.setItem).mockClear();

    const addButton = screen.getByTestId('add-widget');

    // Trigger update 1
    act(() => {
      addButton.click();
    });

    // localStorage should NOT be called yet because of 1000ms debounce
    expect(localStorage.setItem).not.toHaveBeenCalled();

    // Trigger update 2 immediately
    act(() => {
      addButton.click();
    });

    expect(localStorage.setItem).not.toHaveBeenCalled();

    // Fast forward time past the 1000ms debounce
    await act(async () => {
      await vi.advanceTimersByTimeAsync(1500);
    });

    // Now it should be called ONCE (for the latest state)
    expect(localStorage.setItem).toHaveBeenCalledTimes(1);
  });
});
