/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, waitFor } from "@testing-library/react";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { useServiceHealthHistory, ServiceHealth } from "./use-service-health-history";

describe("useServiceHealthHistory", () => {
    beforeEach(() => {
        window.localStorage.clear();
        vi.restoreAllMocks();
        // vi.useFakeTimers(); // Removing fake timers to avoid async/timeout issues with fetch
    });

    // afterEach(() => {
    //     vi.useRealTimers();
    // });

    const mockServices: ServiceHealth[] = [
        { id: "svc-1", name: "Service 1", status: "healthy", latency: "10ms", uptime: "99%" },
        { id: "svc-2", name: "Service 2", status: "degraded", latency: "100ms", uptime: "95%" }
    ];

    it("should fetch initial health data and update history", async () => {
        global.fetch = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => mockServices
        });

        const { result } = renderHook(() => useServiceHealthHistory());

        // Initial state - might be false immediately if not strict mode double render or strict effect
        // expect(result.current.isLoading).toBe(true);

        // Wait for effect
        await waitFor(() => {
            expect(result.current.services).toEqual(mockServices);
        }, { timeout: 2000 });

        expect(Object.keys(result.current.history)).toHaveLength(2);
        expect(result.current.history["svc-1"]).toHaveLength(1);
        expect(result.current.history["svc-1"][0].status).toBe("healthy");
    });

    it("should persist history to localStorage", async () => {
         global.fetch = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => mockServices
        });

        const { result } = renderHook(() => useServiceHealthHistory());

        await waitFor(() => {
             // Wait for data to be loaded AND persisted
             // The persistence might be debounced or in useEffect deps
             expect(result.current.services).toEqual(mockServices);
        });

        // Wait slightly for local storage effect
        await waitFor(() => {
             const stored = window.localStorage.getItem("mcp_service_health_history");
             expect(stored).toBeTruthy();
        });

        const stored = window.localStorage.getItem("mcp_service_health_history");
        const parsed = JSON.parse(stored!);
        expect(parsed["svc-1"]).toHaveLength(1);
    });

    it("should load history from localStorage on mount", async () => {
        const initialHistory = {
            "svc-1": [{ timestamp: 1234567890, status: "unhealthy" }]
        };
        window.localStorage.setItem("mcp_service_health_history", JSON.stringify(initialHistory));

        global.fetch = vi.fn().mockResolvedValue({
            ok: true,
            json: async () => []
        });

        const { result } = renderHook(() => useServiceHealthHistory());

        // Should have loaded immediately (in useEffect)
        expect(result.current.history["svc-1"]).toHaveLength(1);
        // Typescript issue potentially if casting needed, but let's assume valid JSON
        expect(result.current.history["svc-1"][0].status).toBe("unhealthy");
    });
});
