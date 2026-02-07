/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen } from "@testing-library/react";
import { ServiceHealthWidget } from "./service-health-widget";
import { vi, describe, it, expect, beforeEach } from "vitest";
import { useServiceHealthHistory, ServiceHealth, ServiceHistory } from "@/hooks/use-service-health-history";
import { TooltipProvider } from "@/components/ui/tooltip";

// Mock the hook
vi.mock("@/hooks/use-service-health-history", () => ({
  useServiceHealthHistory: vi.fn(),
}));

describe("ServiceHealthWidget", () => {
  const mockUseServiceHealthHistory = useServiceHealthHistory as unknown as ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders loading state", () => {
    mockUseServiceHealthHistory.mockReturnValue({
      services: [],
      history: {},
      isLoading: true,
    });

    render(<ServiceHealthWidget />);
    expect(screen.getByText("Checking system status...")).toBeInTheDocument();
  });

  it("renders empty state", () => {
    mockUseServiceHealthHistory.mockReturnValue({
      services: [],
      history: {},
      isLoading: false,
    });

    render(<ServiceHealthWidget />);
    expect(screen.getByText("No services connected.")).toBeInTheDocument();
  });

  it("renders services and timeline", () => {
    const mockServices: ServiceHealth[] = [
      {
        id: "svc1",
        name: "Test Service 1",
        status: "healthy",
        latency: "10ms",
        uptime: "1h",
      },
    ];
    const mockHistory: ServiceHistory = {
      svc1: [
        { timestamp: 1000, status: "healthy" },
        { timestamp: 2000, status: "healthy" },
      ],
    };

    mockUseServiceHealthHistory.mockReturnValue({
      services: mockServices,
      history: mockHistory,
      isLoading: false,
    });

    render(
      <TooltipProvider>
        <ServiceHealthWidget />
      </TooltipProvider>
    );
    expect(screen.getByText("Test Service 1")).toBeInTheDocument();
    expect(screen.getByText("Live health checks for 1 connected services.")).toBeInTheDocument();
  });

  it("sorts services by status", () => {
     const mockServices: ServiceHealth[] = [
      { id: "svc1", name: "Healthy Svc", status: "healthy", latency: "10ms", uptime: "1h" },
      { id: "svc2", name: "Unhealthy Svc", status: "unhealthy", latency: "0ms", uptime: "0h" },
    ];

    mockUseServiceHealthHistory.mockReturnValue({
      services: mockServices,
      history: {},
      isLoading: false,
    });

    render(<ServiceHealthWidget />);
    // querying by text content might return multiple elements (like the Badge and the Name)
    // We want the name elements specifically or the row containers.
    // The name is in a p tag with font-medium.
    // Let's rely on the order of appearance in the document.
    const healthy = screen.getByText("Healthy Svc");
    const unhealthy = screen.getByText("Unhealthy Svc");

    // comparePosition: 2 means 'preceding', 4 means 'following'.
    // If unhealthy comes before healthy, then unhealthy.compareDocumentPosition(healthy) should be 4 (healthy follows unhealthy)
    expect(unhealthy.compareDocumentPosition(healthy)).toBe(4);
  });
});
