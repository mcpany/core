/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from "@testing-library/react";
import { SystemDiagnosticsWidget } from "./system-diagnostics-widget";
import { vi, describe, it, expect, beforeEach } from "vitest";
import { apiClient } from "@/lib/client";
import { analyzeTrace } from "@/lib/diagnostics";

// Mock deps
vi.mock("@/lib/client", () => ({
  apiClient: {
    listTraces: vi.fn()
  }
}));

vi.mock("@/lib/diagnostics", () => ({
  analyzeTrace: vi.fn()
}));

// Mock Link
vi.mock("next/link", () => ({
  default: ({ children }: { children: React.ReactNode }) => <div>{children}</div>
}));

describe("SystemDiagnosticsWidget", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders loading state", () => {
    (apiClient.listTraces as any).mockReturnValue(new Promise(() => {})); // Hang
    render(<SystemDiagnosticsWidget />);
    expect(screen.getByText("Analyzing system traces...")).toBeInTheDocument();
  });

  it("renders empty state", async () => {
    (apiClient.listTraces as any).mockResolvedValue([]);
    render(<SystemDiagnosticsWidget />);
    await waitFor(() => {
      expect(screen.getByText("All Systems Operational")).toBeInTheDocument();
    });
  });

  it("renders aggregated diagnostics", async () => {
    const traces = [
      { id: "1", status: "error" },
      { id: "2", status: "error" },
      { id: "3", status: "success" },
    ];
    (apiClient.listTraces as any).mockResolvedValue(traces);

    (analyzeTrace as any).mockImplementation((trace: any) => {
        if (trace.id === "1") return [{ title: "Schema Error", type: "error", message: "Invalid schema" }];
        if (trace.id === "2") return [{ title: "Schema Error", type: "error", message: "Invalid schema" }];
        return [];
    });

    render(<SystemDiagnosticsWidget />);

    await waitFor(() => {
        expect(screen.getByText("Schema Error")).toBeInTheDocument();
        expect(screen.getByText("2 issues")).toBeInTheDocument();
    });
  });
});
