/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from "@testing-library/react";
import { PipelineVisualizer } from "./pipeline-visualizer";
import { describe, it, expect, vi, beforeEach } from "vitest";

// Mock fetch
global.fetch = vi.fn();

// Mock dnd module
vi.mock("@hello-pangea/dnd", () => ({
    DragDropContext: ({ children }: any) => <div>{children}</div>,
    Droppable: ({ children }: any) => children({ droppableProps: {}, innerRef: null }),
    Draggable: ({ children }: any) => children({ draggableProps: {}, dragHandleProps: {}, innerRef: null }),
}));

describe("PipelineVisualizer", () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it("renders middlewares in order", async () => {
        (global.fetch as any).mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                middlewares: [
                    { name: "auth", priority: 10 },
                    { name: "logging", priority: 20 }
                ]
            })
        });

        render(<PipelineVisualizer />);

        await waitFor(() => {
            expect(screen.getByText("auth")).toBeInTheDocument();
            expect(screen.getByText("logging")).toBeInTheDocument();
        });

        // Check priority display
        expect(screen.getByText("Priority: 10")).toBeInTheDocument();
        expect(screen.getByText("Priority: 20")).toBeInTheDocument();
    });

    it("shows toggle state correctly", async () => {
         (global.fetch as any).mockResolvedValueOnce({
            ok: true,
            json: async () => ({
                middlewares: [
                    { name: "auth", priority: 10, disabled: false },
                    { name: "logging", priority: 20, disabled: true }
                ]
            })
        });

        render(<PipelineVisualizer />);

        await waitFor(() => {
            expect(screen.getByText("auth")).toBeInTheDocument();
        });

        expect(screen.getByText("On")).toBeInTheDocument();
        expect(screen.getByText("Off")).toBeInTheDocument();
    });
});
