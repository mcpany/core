/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "../../tests/test-utils";
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
    configurationSchema: "",
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
    configurationSchema: "",
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
      <TooltipProvider>
        <ServiceHealthProvider>
          {component}
        </ServiceHealthProvider>
      </TooltipProvider>
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

  it("toggles view mode between table and grid", () => {
    // Override localStorage for this test
    Storage.prototype.getItem = jest.fn(() => null);
    Storage.prototype.setItem = jest.fn();

    const { container } = renderWithProvider(<ServiceList services={mockServices} />);

    // Initially should render Table mode (which uses <table> elements)
    expect(container.querySelector('table')).toBeInTheDocument();

    // Find the toggle buttons (LayoutGrid and List icons inside buttons)
    // The grid toggle button has a specific class or icon
    const buttons = screen.getAllByRole('button');
    // Assuming the second button in the header is the Grid view toggle
    // It's the one after the List button in the new MR
    // We can find by looking for the button that sets view mode to grid
    // Easiest is to simulate clicking the grid button
    const gridButton = buttons.find(b => b.innerHTML.includes('lucide-layout-grid'));

    if (gridButton) {
        fireEvent.click(gridButton);
    }

    // After clicking grid, table should disappear
    expect(container.querySelector('table')).not.toBeInTheDocument();

    // We should still see the services
    expect(screen.getByText("Service 1")).toBeInTheDocument();
    expect(screen.getByText("Service 2")).toBeInTheDocument();

    // Local storage should be updated
    expect(localStorage.setItem).toHaveBeenCalledWith('service_list_view_mode', 'grid');
  });
});
