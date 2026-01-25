/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { TraceList } from "./trace-list";
import { vi, describe, it, expect } from "vitest";
import { Trace } from "@/app/api/traces/route";

const mockTraces: Trace[] = [
    {
        id: "t1",
        timestamp: Date.now(),
        rootSpan: { name: "tool.success", type: 'tool', startTime: 0, endTime: 10, status: 'success', input: {}, output: {}, children: [] },
        status: "success",
        totalDuration: 10,
        trigger: "user"
    },
    {
        id: "t2",
        timestamp: Date.now(),
        rootSpan: { name: "tool.error", type: 'tool', startTime: 0, endTime: 10, status: 'error', input: {}, output: {}, children: [], errorMessage: "failed" },
        status: "error",
        totalDuration: 10,
        trigger: "user"
    }
];

describe("TraceList", () => {
    it("renders traces", () => {
        render(<TraceList traces={mockTraces} selectedId={null} onSelect={() => {}} searchQuery="" onSearchChange={() => {}} statusFilter="all" onStatusFilterChange={() => {}} isLive={false} onToggleLive={() => {}} />);
        expect(screen.getByText("tool.success")).toBeInTheDocument();
        expect(screen.getByText("tool.error")).toBeInTheDocument();
    });

    it("filters by status", () => {
        const { rerender } = render(<TraceList traces={mockTraces} selectedId={null} onSelect={() => {}} searchQuery="" onSearchChange={() => {}} statusFilter="success" onStatusFilterChange={() => {}} isLive={false} onToggleLive={() => {}} />);
        expect(screen.getByText("tool.success")).toBeInTheDocument();
        expect(screen.queryByText("tool.error")).not.toBeInTheDocument();

        rerender(<TraceList traces={mockTraces} selectedId={null} onSelect={() => {}} searchQuery="" onSearchChange={() => {}} statusFilter="error" onStatusFilterChange={() => {}} isLive={false} onToggleLive={() => {}} />);
        expect(screen.queryByText("tool.success")).not.toBeInTheDocument();
        expect(screen.getByText("tool.error")).toBeInTheDocument();
    });
});
