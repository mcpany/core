/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { ServiceList } from "./service-list";
import { UpstreamServiceConfig } from "@/lib/client";
import { ServiceHealthProvider } from "@/contexts/service-health-context";
import { TooltipProvider } from "@/components/ui/tooltip";

const mockServices: UpstreamServiceConfig[] = [
  {
    id: "s1",
    name: "Service 1",
    version: "1.0",
    disable: false,
    priority: 0,
    loadBalancingStrategy: 0,
    tags: ["prod", "db"],
    sanitizedName: "service-1",
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    prompts: [],
    autoDiscoverTool: false,
    configError: "",
    readOnly: false,
    httpService: {
        address: "http://localhost:8080",
        tools: [],
        calls: {},
        resources: [],
        prompts: []
    }
  },
  {
    id: "s2",
    name: "Service 2",
    version: "1.0",
    disable: false,
    priority: 0,
    loadBalancingStrategy: 0,
    tags: ["dev", "external"],
    sanitizedName: "service-2",
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    prompts: [],
    autoDiscoverTool: false,
    configError: "",
    readOnly: false,
    httpService: {
        address: "http://localhost:8081",
        tools: [],
        calls: {},
        resources: [],
        prompts: []
    }
  }
];

describe("ServiceList", () => {
  const renderWithProvider = (component: React.ReactNode) => {
    return render(
      <ServiceHealthProvider>
        <TooltipProvider>
            {component}
        </TooltipProvider>
      </ServiceHealthProvider>
    );
  };

  it("renders services", () => {
    renderWithProvider(<ServiceList services={mockServices} />);
    expect(screen.getByText("Service 1")).toBeInTheDocument();
    expect(screen.getByText("Service 2")).toBeInTheDocument();
  });

  it("filters services by tag", () => {
    renderWithProvider(<ServiceList services={mockServices} />);

    const input = screen.getByPlaceholderText("Filter by tag...");
    fireEvent.change(input, { target: { value: "prod" } });

    expect(screen.getByText("Service 1")).toBeInTheDocument();
    expect(screen.queryByText("Service 2")).not.toBeInTheDocument();
  });

  it("filters services by partial tag match", () => {
    renderWithProvider(<ServiceList services={mockServices} />);

    const input = screen.getByPlaceholderText("Filter by tag...");
    fireEvent.change(input, { target: { value: "ext" } });

    expect(screen.queryByText("Service 1")).not.toBeInTheDocument();
    expect(screen.getByText("Service 2")).toBeInTheDocument();
  });

  it("shows no results when no match", () => {
    renderWithProvider(<ServiceList services={mockServices} />);

    const input = screen.getByPlaceholderText("Filter by tag...");
    fireEvent.change(input, { target: { value: "missing" } });

    expect(screen.queryByText("Service 1")).not.toBeInTheDocument();
    expect(screen.queryByText("Service 2")).not.toBeInTheDocument();
    expect(screen.getByText("No services match the tag filter.")).toBeInTheDocument();
  });

  it("sorts services by name", () => {
    const unsortedServices = [
      { ...mockServices[1], name: "Service B" },
      { ...mockServices[0], name: "Service A" }
    ];
    renderWithProvider(<ServiceList services={unsortedServices} />);

    const nameHeader = screen.getByText("Name");
    fireEvent.click(nameHeader); // Sort Ascending

    const rows = screen.getAllByRole("row");
    // Row 0 is header. Row 1 is first service (Service A).
    expect(rows[1]).toHaveTextContent("Service A");
    expect(rows[2]).toHaveTextContent("Service B");

    fireEvent.click(nameHeader); // Sort Descending
    const rowsDesc = screen.getAllByRole("row");
    expect(rowsDesc[1]).toHaveTextContent("Service B");
    expect(rowsDesc[2]).toHaveTextContent("Service A");
  });

  it("toggles density", () => {
    renderWithProvider(<ServiceList services={mockServices} />);

    const toggleButton = screen.getByLabelText("Toggle density");

    // Initial state: default density (p-4)
    const initialCells = screen.getAllByRole("cell");
    expect(initialCells[0]).toHaveClass("p-4");

    // Toggle to compact
    fireEvent.click(toggleButton);
    const compactCells = screen.getAllByRole("cell");
    expect(compactCells[0]).toHaveClass("p-2");

    // Toggle back to default
    fireEvent.click(toggleButton);
    const defaultCells = screen.getAllByRole("cell");
    expect(defaultCells[0]).toHaveClass("p-4");
  });
});
