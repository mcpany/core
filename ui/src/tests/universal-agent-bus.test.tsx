import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import UniversalAgentBusPage from "../app/universal-agent-bus/page";

// Mock matchMedia to fix "matchMedia not present" error
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

describe("UniversalAgentBusPage", () => {
  it("renders the page title and description", () => {
    render(<UniversalAgentBusPage />);

    expect(screen.getByText("Universal Agent Bus")).toBeInTheDocument();
    expect(
      screen.getByText(/Manage and map subagents dynamically/i)
    ).toBeInTheDocument();
  });

  it("renders all expected feature cards", () => {
    render(<UniversalAgentBusPage />);

    expect(screen.getByText("Recursive Context Dashboard")).toBeInTheDocument();
    expect(screen.getByText("Multi-Agent Session Timeline")).toBeInTheDocument();
    expect(screen.getByText("Unified Discovery Manager")).toBeInTheDocument();
    expect(screen.getByText("Lazy-MCP Tool Search Dashboard")).toBeInTheDocument();
    expect(screen.getByText("Agent Chain Tracer (A2A)")).toBeInTheDocument();
  });
});
