/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from "@testing-library/react";
import { ToolFailureRateWidget } from "./tool-failure-rate-widget";
import { vi, describe, it, expect } from "vitest";
import { apiClient } from "@/lib/client";

// Mock API Client
vi.mock("@/lib/client", () => ({
  apiClient: {
    getToolFailures: vi.fn()
  }
}));

// Mock Link if necessary, but usually standard next/link works partially in tests if we check the anchor
// Or we can mock it to just be an anchor
vi.mock("next/link", () => ({
  default: ({ children, href }: { children: React.ReactNode, href: string }) => <a href={href}>{children}</a>
}));

describe("ToolFailureRateWidget", () => {
    it("renders tool failure rates with links", async () => {
        const mockData = [
            { name: "test_tool", serviceId: "service_a", failureRate: 20.5, totalCalls: 100 },
            { name: "safe_tool", serviceId: "service_b", failureRate: 0, totalCalls: 50 }
        ];

        (apiClient.getToolFailures as any).mockResolvedValue(mockData);

        render(<ToolFailureRateWidget />);

        await waitFor(() => {
            expect(screen.getByText("test_tool")).toBeInTheDocument();
        });

        const link = screen.getByText("test_tool").closest("a");
        expect(link).toHaveAttribute("href", "/traces?tool=test_tool&status=error");

        expect(screen.getByText("20.5%")).toBeInTheDocument();
        expect(screen.getByText("safe_tool")).toBeInTheDocument();
    });

    it("handles empty state", async () => {
         (apiClient.getToolFailures as any).mockResolvedValue([]);
         render(<ToolFailureRateWidget />);
         await waitFor(() => {
             expect(screen.getByText("No tool call data available.")).toBeInTheDocument();
         });
    });
});
