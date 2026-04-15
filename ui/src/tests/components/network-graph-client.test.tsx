/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { NetworkGraphClient } from "@/components/network/network-graph-client";
import { MemoryRouter } from "react-router-dom";

// Mock matchMedia for JSDOM
Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation(query => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(), // deprecated
        removeListener: vi.fn(), // deprecated
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
    })),
});

// Mock xyflow since it's hard to test in JSDOM
vi.mock("@xyflow/react", async () => {
    const actual = await vi.importActual("@xyflow/react") as any;
    return {
        ...actual,
        ReactFlow: ({ children }: any) => <div data-testid="react-flow-mock">{children}</div>,
        Background: () => <div data-testid="react-flow-background" />,
        Controls: () => <div data-testid="react-flow-controls" />,
        Panel: ({ children }: any) => <div data-testid="react-flow-panel">{children}</div>,
        useNodesState: () => [[], vi.fn()],
        useEdgesState: () => [[], vi.fn()],
        addEdge: vi.fn(),
        MarkerType: { ArrowClosed: "arrowclosed" },
        Position: { Top: "top", Bottom: "bottom", Left: "left", Right: "right" },
        ReactFlowProvider: ({ children }: any) => <div data-testid="react-flow-provider-mock">{children}</div>,
    };
});

// Mock hooks
vi.mock("@/hooks/use-network-topology", () => ({
    useNetworkTopology: () => ({
        nodes: [],
        edges: [],
        onNodesChange: vi.fn(),
        onEdgesChange: vi.fn(),
        onConnect: vi.fn(),
    }),
}));

describe("NetworkGraphClient Component", () => {
    it("renders the NetworkGraphClient and ReactFlow provider", () => {
        render(
            <MemoryRouter>
                <NetworkGraphClient />
            </MemoryRouter>
        );
        expect(screen.getByTestId("react-flow-provider-mock")).toBeInTheDocument();
        expect(screen.getByTestId("react-flow-mock")).toBeInTheDocument();
    });

    it("renders the legend inside the panel", () => {
        render(
            <MemoryRouter>
                <NetworkGraphClient />
            </MemoryRouter>
        );
        // Panel should contain legend items like "Core", "Service", "Client"
        expect(screen.getByText("Core")).toBeInTheDocument();
        expect(screen.getByText("Service")).toBeInTheDocument();
        expect(screen.getByText("Client")).toBeInTheDocument();
    });
});
