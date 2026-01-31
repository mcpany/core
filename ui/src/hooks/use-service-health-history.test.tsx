/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, waitFor } from "@testing-library/react";
import { describe, it, expect, beforeEach, vi } from "vitest";
import { useServiceHealthHistory, ServiceHealth } from "./use-service-health-history";

describe("useServiceHealthHistory", () => {
    beforeEach(() => {
        vi.restoreAllMocks();
    });

    const mockServices: ServiceHealth[] = [
        { id: "svc-1", name: "Service 1", status: "healthy", latency: "10ms", uptime: "99%" },
        { id: "svc-2", name: "Service 2", status: "degraded", latency: "100ms", uptime: "95%" }
    ];

    const mockHistory = {
        "svc-1": [{ timestamp: 1234567890, status: "healthy" }],
        "svc-2": [{ timestamp: 1234567890, status: "degraded" }]
    };

    it("should fetch initial health data and update history", async () => {
        global.fetch = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => ({
                services: mockServices,
                history: mockHistory
            })
        });

        const { result } = renderHook(() => useServiceHealthHistory());

        // Initial state
        expect(result.current.isLoading).toBe(true);

        // Wait for effect
        await waitFor(() => {
            expect(result.current.isLoading).toBe(false);
        });

        // Verify URL
        expect(global.fetch).toHaveBeenCalledWith("/api/v1/dashboard/health");

        expect(result.current.services).toEqual(mockServices);
        expect(result.current.history).toEqual(mockHistory);
    });
});
