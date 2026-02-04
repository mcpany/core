/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { SmartToolSearch } from "@/components/tools/smart-tool-search";
import { ToolDefinition } from "@proto/config/v1/tool";
import { vi } from "vitest";

// Mock useRecentTools hook
const mockAddRecent = vi.fn();
const mockRecentTools: string[] = [];

vi.mock("@/hooks/use-recent-tools", () => ({
  useRecentTools: () => ({
    recentTools: mockRecentTools,
    addRecent: mockAddRecent,
  }),
}));

// Mock the Command components to simplify testing (avoiding cmdk internal logic complexity in unit tests)
// We just want to verify that SmartToolSearch passes the correct props to CommandItem
vi.mock("@/components/ui/command", () => {
  return {
    Command: ({ children, className }: any) => <div className={className}>{children}</div>,
    CommandInput: ({ placeholder, value, onValueChange, onFocus }: any) => (
      <input
        placeholder={placeholder}
        value={value}
        onChange={(e) => onValueChange(e.target.value)}
        onFocus={onFocus}
        data-testid="command-input"
      />
    ),
    CommandList: ({ children }: any) => <div data-testid="command-list">{children}</div>,
    CommandEmpty: ({ children }: any) => <div data-testid="command-empty">{children}</div>,
    CommandGroup: ({ children, heading }: any) => (
      <div data-testid={`command-group-${heading}`}>
        {heading}
        {children}
      </div>
    ),
    CommandItem: ({ children, value, onSelect, className }: any) => (
      <div
        data-testid="command-item"
        data-value={value}
        onClick={onSelect}
        className={className}
      >
        {children}
      </div>
    ),
    CommandSeparator: () => <hr />,
  };
});

describe("SmartToolSearch", () => {
  const mockTools: ToolDefinition[] = [
    {
      name: "tool-a",
      description: "Description for tool A",
      serviceId: "service-1",
      inputSchema: {},
    },
    {
      name: "tool-b",
      description: "Description for tool B",
      serviceId: "service-2",
      inputSchema: {},
    },
  ];

  const mockSetSearchQuery = vi.fn();
  const mockOnToolSelect = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders search input", () => {
    render(
      <SmartToolSearch
        tools={mockTools}
        searchQuery=""
        setSearchQuery={mockSetSearchQuery}
        onToolSelect={mockOnToolSelect}
      />
    );
    expect(screen.getByTestId("command-input")).toBeInTheDocument();
  });

  it("opens dropdown on focus", () => {
    render(
      <SmartToolSearch
        tools={mockTools}
        searchQuery=""
        setSearchQuery={mockSetSearchQuery}
        onToolSelect={mockOnToolSelect}
      />
    );

    // Initially list should not be visible
    expect(screen.queryByTestId("command-list")).not.toBeInTheDocument();

    // Focus input
    fireEvent.focus(screen.getByTestId("command-input"));

    // List should be visible
    expect(screen.getByTestId("command-list")).toBeInTheDocument();
  });

  it("updates search query on input", () => {
    render(
      <SmartToolSearch
        tools={mockTools}
        searchQuery=""
        setSearchQuery={mockSetSearchQuery}
        onToolSelect={mockOnToolSelect}
      />
    );

    fireEvent.change(screen.getByTestId("command-input"), { target: { value: "test" } });
    expect(mockSetSearchQuery).toHaveBeenCalledWith("test");
  });

  it("renders tool items with correct search values (including serviceId)", () => {
    render(
      <SmartToolSearch
        tools={mockTools}
        searchQuery=""
        setSearchQuery={mockSetSearchQuery}
        onToolSelect={mockOnToolSelect}
      />
    );

    fireEvent.focus(screen.getByTestId("command-input"));

    const items = screen.getAllByTestId("command-item");
    expect(items).toHaveLength(2);

    // Verify the data-value attribute contains name, description, and serviceId
    expect(items[0]).toHaveAttribute("data-value", "tool-a Description for tool A service-1");
    expect(items[1]).toHaveAttribute("data-value", "tool-b Description for tool B service-2");
  });

  it("calls onToolSelect and addRecent when item is selected", () => {
    render(
      <SmartToolSearch
        tools={mockTools}
        searchQuery=""
        setSearchQuery={mockSetSearchQuery}
        onToolSelect={mockOnToolSelect}
      />
    );

    fireEvent.focus(screen.getByTestId("command-input"));
    const items = screen.getAllByTestId("command-item");

    fireEvent.click(items[0]); // Select tool-a

    expect(mockAddRecent).toHaveBeenCalledWith("tool-a");
    expect(mockOnToolSelect).toHaveBeenCalledWith(mockTools[0]);
  });
});
