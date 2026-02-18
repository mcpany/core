/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, waitFor } from "@testing-library/react";
import { describe, it, expect, beforeEach, vi } from "vitest";
import { useServiceHealthHistory, ServiceHealth } from "./use-service-health-history";
import { ServiceHealthProvider } from "@/contexts/service-health-context";
import React from "react";

describe("useServiceHealthHistory", () => {
    beforeEach(() => {
        vi.restoreAllMocks();
    });

    const mockServices: ServiceHealth[] = [
        { id: "svc-1", name: "Service 1", status: "healthy", latency: "10ms", uptime: "99%" },
        { id: "svc-2", name: "Service 2", status: "degraded", latency: "100ms", uptime: "95%" }
    ];

    const mockHistory = {
        "svc-1": [{ timestamp: 1234567890, status: "healthy" } as any],
        "svc-2": [{ timestamp: 1234567890, status: "degraded" } as any]
    };

    it("should fetch initial health data and update history from server", async () => {
        global.fetch = vi.fn().mockImplementation((url) => {
            if (url.toString().includes("/api/v1/dashboard/health")) {
                return Promise.resolve({
                    ok: true,
                    json: async () => ({
                        services: mockServices,
                        history: mockHistory
                    })
                });
            }
            if (url.toString().includes("/api/v1/topology")) {
                return Promise.resolve({
                    ok: true,
                    json: async () => ({}),
                    text: async () => "{}",
                    headers: { get: () => null }
                });
            }
            return Promise.reject(new Error(`Unknown URL: ${url}`));
        });

        const wrapper = ({ children }: { children: React.ReactNode }) => (
            <ServiceHealthProvider>{children}</ServiceHealthProvider>
        );

        const { result } = renderHook(() => useServiceHealthHistory(), { wrapper });

        // Initial state
        expect(result.current.isLoading).toBe(false);

        // Wait for provider to fetch and update context
        await waitFor(() => {
            expect(result.current.services).toEqual(mockServices);
        });

        expect(Object.keys(result.current.history)).toHaveLength(2);
        expect(result.current.history["svc-1"]).toHaveLength(1);
        expect(result.current.history["svc-1"][0].status).toBe("healthy");
    });
});
