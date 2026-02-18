/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { renderHook, waitFor, render, act } from "@testing-library/react";
import { ServiceHealthProvider, useServiceHealth, useTopology } from "./service-health-context";
import React from "react";
import { describe, it, expect, beforeEach, vi } from "vitest";

// Mock global fetch
global.fetch = vi.fn();

describe("ServiceHealthContext", () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    const mockTopology = {
        core: {
            id: "core",
            type: "NODE_TYPE_CORE",
            children: [
                {
                    id: "service-1",
                    type: "NODE_TYPE_SERVICE",
                    status: "NODE_STATUS_ACTIVE",
                    metrics: { latencyMs: 100, errorRate: 0, qps: 5 }
                }
            ]
        }
    };

    const mockHealth = {
        services: [],
        history: {}
    };

    it("fetches topology and updates history", async () => {
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
                     json: async () => mockHealth,
                     headers: new Headers()
                 });
             }
             return Promise.reject(new Error(`Unknown URL: ${url}`));
        });

        const wrapper = ({ children }: { children: React.ReactNode }) => (
            <ServiceHealthProvider>{children}</ServiceHealthProvider>
        );

        const { result } = renderHook(() => useServiceHealth(), { wrapper });

        // Wait for fetch
        await waitFor(() => {
             expect(global.fetch).toHaveBeenCalledWith('/api/v1/topology', expect.anything());
        });

        // Check history
        await waitFor(() => {
            const history = result.current.getServiceHistory("service-1");
            expect(history.length).toBeGreaterThan(0);
            expect(history[0].latencyMs).toBe(100);
            expect(history[0].status).toBe("NODE_STATUS_ACTIVE");
        });
    });

    it("useTopology should not re-render when only metrics update", async () => {
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
                     json: async () => mockHealth,
                     headers: new Headers()
                 });
             }
             return Promise.reject(new Error(`Unknown URL: ${url}`));
        });

        vi.useFakeTimers();

        let renderCount = 0;
        const Consumer = () => {
            useTopology();
            renderCount++;
            return null;
        };

        render(
            <ServiceHealthProvider>
                <Consumer />
            </ServiceHealthProvider>
        );

        // Allow initial effect to run
        await act(async () => {
            await vi.advanceTimersByTimeAsync(100);
        });

        // Verify first fetch
        expect(global.fetch).toHaveBeenCalledWith('/api/v1/topology', expect.anything());

        // Capture render count after initial fetch setup
        const rendersAfterInit = renderCount;

        // Mock next fetch with SAME topology (same content, new object)
        // This simulates a poll where metrics might be processed but topology structure is identical
        const mockTopology2 = JSON.parse(JSON.stringify(mockTopology));
        (global.fetch as any).mockImplementation((url: string) => {
             if (url.includes("/api/v1/topology")) {
                 return Promise.resolve({
                     ok: true,
                     json: async () => mockTopology2,
                     text: async () => JSON.stringify(mockTopology2),
                     headers: new Headers()
                 });
             }
             if (url.includes("/api/v1/dashboard/health")) {
                 return Promise.resolve({
                     ok: true,
                     json: async () => mockHealth,
                     headers: new Headers()
                 });
             }
             return Promise.reject(new Error(`Unknown URL: ${url}`));
        });

        // Advance timer to trigger poll
        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });

        // Verify fetch was called again
        // Expect 2 batches of calls (initial + 1 poll) x 2 endpoints = 4 calls?
        // Wait, initial effect runs once.
        // It calls fetchTopology.
        // fetchTopology calls fetch(topology) then fetch(health).
        // So 2 calls initially.
        // Then setInterval fires.
        // It calls fetchTopology again.
        // So 2 calls again.
        // Total 4 calls.
        // BUT strict counts might be tricky if effect fires strictly or mocked timers behave differently.
        // Let's just check it was called more than 2 times.
        expect(global.fetch).toHaveBeenCalledTimes(4);

        // Verify render count DID NOT increase
        expect(renderCount).toBe(rendersAfterInit);

        vi.useRealTimers();
    }, 10000);
});
