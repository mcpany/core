/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { renderHook, act } from "@testing-library/react";
import { useRealTimeTopology } from "./use-real-time-topology";
import { apiClient } from "@/lib/client";
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";

// Mock apiClient
vi.mock("@/lib/client", () => ({
    apiClient: {
        getTopology: vi.fn(),
    },
}));

// Mock dagre to avoid complex graph logic
vi.mock("dagre", () => {
    const Graph = vi.fn();
    Graph.prototype.setGraph = vi.fn();
    Graph.prototype.setDefaultEdgeLabel = vi.fn();
    Graph.prototype.setNode = vi.fn();
    Graph.prototype.setEdge = vi.fn();
    Graph.prototype.node = vi.fn(() => ({ x: 0, y: 0 }));
    return {
        graphlib: { Graph },
        layout: vi.fn(),
    };
});

describe("useRealTimeTopology", () => {
    beforeEach(() => {
        vi.useFakeTimers();
        // Reset mocks
        vi.mocked(apiClient.getTopology).mockReset();
        vi.mocked(apiClient.getTopology).mockResolvedValue({
            clients: [],
            core: { id: "core", label: "Core", children: [] },
        });

        // Mock document.hidden
        Object.defineProperty(document, "hidden", {
            configurable: true,
            get: () => false,
        });
    });

    afterEach(() => {
        vi.useRealTimers();
        vi.restoreAllMocks();
    });

    it("fetches topology on mount", async () => {
        const { result } = renderHook(() => useRealTimeTopology());

        // Wait for initial fetch
        await act(async () => {
            await Promise.resolve();
        });

        expect(apiClient.getTopology).toHaveBeenCalledTimes(1);
    });

    it("polls when isLive is true", async () => {
        const { result } = renderHook(() => useRealTimeTopology());

        // Enable live mode
        act(() => {
            result.current.setIsLive(true);
        });

        // Wait for polling interval
        await act(async () => {
             vi.advanceTimersByTime(1100);
        });

        // Expect calls: 1 initial + 1 from poll
        expect(apiClient.getTopology).toHaveBeenCalledTimes(2);

        await act(async () => {
             vi.advanceTimersByTime(1000);
        });

        expect(apiClient.getTopology).toHaveBeenCalledTimes(3);
    });

    it("stops polling when isLive is false", async () => {
        const { result } = renderHook(() => useRealTimeTopology());

        act(() => {
            result.current.setIsLive(true);
        });

        await act(async () => {
             vi.advanceTimersByTime(1100);
        });
        expect(apiClient.getTopology).toHaveBeenCalledTimes(2);

        act(() => {
            result.current.setIsLive(false);
        });

        await act(async () => {
             vi.advanceTimersByTime(2000);
        });

        // Should not have increased
        expect(apiClient.getTopology).toHaveBeenCalledTimes(2);
    });

    it("pauses polling when document is hidden", async () => {
        const { result } = renderHook(() => useRealTimeTopology());

        act(() => {
            result.current.setIsLive(true);
        });

        await act(async () => {
             vi.advanceTimersByTime(1100);
        });
        expect(apiClient.getTopology).toHaveBeenCalledTimes(2);

        // Mock hidden = true
        Object.defineProperty(document, "hidden", {
            configurable: true,
            get: () => true,
        });

        // Advance time
        await act(async () => {
             vi.advanceTimersByTime(2000);
        });

        // Should not have increased
        expect(apiClient.getTopology).toHaveBeenCalledTimes(2);
    });
});
