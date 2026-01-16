/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { ServiceList } from "./service-list";
import { UpstreamServiceConfig } from "@/lib/client";

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
  it("renders services", () => {
    render(<ServiceList services={mockServices} />);
    expect(screen.getByText("Service 1")).toBeInTheDocument();
    expect(screen.getByText("Service 2")).toBeInTheDocument();
  });

  it("filters services by tag", () => {
    render(<ServiceList services={mockServices} />);

    const input = screen.getByPlaceholderText("Filter by tag...");
    fireEvent.change(input, { target: { value: "prod" } });

    expect(screen.getByText("Service 1")).toBeInTheDocument();
    expect(screen.queryByText("Service 2")).not.toBeInTheDocument();
  });

  it("filters services by partial tag match", () => {
    render(<ServiceList services={mockServices} />);

    const input = screen.getByPlaceholderText("Filter by tag...");
    fireEvent.change(input, { target: { value: "ext" } });

    expect(screen.queryByText("Service 1")).not.toBeInTheDocument();
    expect(screen.getByText("Service 2")).toBeInTheDocument();
  });

  it("shows no results when no match", () => {
    render(<ServiceList services={mockServices} />);

    const input = screen.getByPlaceholderText("Filter by tag...");
    fireEvent.change(input, { target: { value: "missing" } });

    expect(screen.queryByText("Service 1")).not.toBeInTheDocument();
    expect(screen.queryByText("Service 2")).not.toBeInTheDocument();
    expect(screen.getByText("No services match the tag filter.")).toBeInTheDocument();
  });
});
