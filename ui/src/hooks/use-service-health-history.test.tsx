/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, waitFor } from "@testing-library/react";
import { describe, it, expect, beforeEach, vi } from "vitest";
import { useServiceHealthHistory, ServiceHealth } from "./use-service-health-history";
import { ServiceHealthProvider } from "../contexts/service-health-context";
import React from "react";

// Mock global fetch
global.fetch = vi.fn();

describe("useServiceHealthHistory", () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    const mockServices: ServiceHealth[] = [
        { id: "svc-1", name: "Service 1", status: "healthy", latency: "10ms", uptime: "99%" },
        { id: "svc-2", name: "Service 2", status: "degraded", latency: "100ms", uptime: "95%" }
    ];

    const mockHistory = {
        "svc-1": [{ timestamp: 1234567890, status: "healthy" } as any],
        "svc-2": [{ timestamp: 1234567890, status: "degraded" } as any]
    };

    const mockTopology = {
        core: { id: "core" }
    };

    it("should fetch initial health data and update history from context", async () => {
        (global.fetch as any).mockImplementation((url: string) => {
            if (url.includes("/api/v1/topology")) {
                return Promise.resolve({
                    ok: true,
                    json: async () => mockTopology,
                    text: async () => JSON.stringify(mockTopology),
                    headers: new Headers()
                });
            }
            if (url.includes("/api/v1/dashboard/health")) {
                return Promise.resolve({
                    ok: true,
                    json: async () => ({
                        services: mockServices,
                        history: mockHistory
                    })
                });
            }
            return Promise.reject(new Error(`Unknown URL: ${url}`));
        });

        const wrapper = ({ children }: { children: React.ReactNode }) => (
            <ServiceHealthProvider>{children}</ServiceHealthProvider>
        );

        const { result } = renderHook(() => useServiceHealthHistory(), { wrapper });

        // Initial state might be loading
        // Wait for services to be populated
        await waitFor(() => {
            expect(result.current.services).toEqual(mockServices);
        });

        expect(result.current.isLoading).toBe(false);
        expect(Object.keys(result.current.history)).toHaveLength(2);
        expect(result.current.history["svc-1"]).toHaveLength(1);
        expect(result.current.history["svc-1"][0].status).toBe("healthy");
    });
});
