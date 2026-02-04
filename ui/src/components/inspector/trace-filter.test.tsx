/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { TraceFilter } from "./trace-filter";
import { vi, describe, it, expect } from "vitest";

// Mock Radix UI Select to simplify testing if needed, but integration test is better.
// However, Radix Select renders into a Portal, which might not be available in JSDOM default setup without configuration.
// Let's assume standard testing library setup handles it or we use pointer events.
// For simplicity in this environment, I will verify the Input mostly, and check if Select Trigger renders.

describe("TraceFilter", () => {
  it("renders with initial values", () => {
    const setSearchQuery = vi.fn();
    const setStatusFilter = vi.fn();

    render(
      <TraceFilter
        searchQuery="initial-search"
        setSearchQuery={setSearchQuery}
        statusFilter="all"
        setStatusFilter={setStatusFilter}
      />
    );

    const input = screen.getByPlaceholderText("Filter by name or ID...");
    expect(input).toBeInTheDocument();
    expect(input).toHaveValue("initial-search");

    // Check if Select Trigger is present (it has role 'combobox')
    expect(screen.getByRole("combobox")).toBeInTheDocument();
  });

  it("calls setSearchQuery on input change", () => {
    const setSearchQuery = vi.fn();
    const setStatusFilter = vi.fn();

    render(
      <TraceFilter
        searchQuery=""
        setSearchQuery={setSearchQuery}
        statusFilter="all"
        setStatusFilter={setStatusFilter}
      />
    );

    const input = screen.getByPlaceholderText("Filter by name or ID...");
    fireEvent.change(input, { target: { value: "test-query" } });

    expect(setSearchQuery).toHaveBeenCalledWith("test-query");
  });
});
