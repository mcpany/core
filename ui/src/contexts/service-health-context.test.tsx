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

    it("fetches topology and updates history", async () => {
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

        (global.fetch as any).mockResolvedValue({
            ok: true,
            json: async () => mockTopology,
            text: async () => JSON.stringify(mockTopology)
        });

        const wrapper = ({ children }: { children: React.ReactNode }) => (
            <ServiceHealthProvider>{children}</ServiceHealthProvider>
        );

        const { result } = renderHook(() => useServiceHealth(), { wrapper });

        // Wait for fetch
        await waitFor(() => {
             expect(global.fetch).toHaveBeenCalledWith('/api/v1/topology');
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

        (global.fetch as any).mockResolvedValue({
            ok: true,
            json: async () => mockTopology,
            text: async () => JSON.stringify(mockTopology)
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
        expect(global.fetch).toHaveBeenCalledWith('/api/v1/topology');

        // Capture render count after initial fetch setup
        const rendersAfterInit = renderCount;

        // Mock next fetch with SAME topology (same content, new object)
        // This simulates a poll where metrics might be processed but topology structure is identical
        const mockTopology2 = JSON.parse(JSON.stringify(mockTopology));
        (global.fetch as any).mockResolvedValue({
            ok: true,
            json: async () => mockTopology2,
            text: async () => JSON.stringify(mockTopology2)
        });

        // Advance timer to trigger poll
        await act(async () => {
            await vi.advanceTimersByTimeAsync(5000);
        });

        // Verify fetch was called again
        expect(global.fetch).toHaveBeenCalledTimes(2);

        // Verify render count DID NOT increase
        expect(renderCount).toBe(rendersAfterInit);

        vi.useRealTimers();
    }, 10000);
});
