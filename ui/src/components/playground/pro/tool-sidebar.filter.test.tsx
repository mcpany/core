/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { ToolSidebar } from "./tool-sidebar";
import { ToolDefinition } from "@/lib/client";

// Mock ToolDefinition with tags
interface ExtendedToolDefinition extends ToolDefinition {
    tags?: string[];
}

const mockTools: ExtendedToolDefinition[] = [
    {
        name: "tool1",
        description: "Tool 1",
        serviceId: "service-a",
        tags: ["tag1", "tag2"],
        inputSchema: {},
    },
    {
        name: "tool2",
        description: "Tool 2",
        serviceId: "service-b",
        tags: ["tag2"],
        inputSchema: {},
    },
    {
        name: "tool3",
        description: "Tool 3",
        serviceId: "service-a",
        tags: ["tag3"],
        inputSchema: {},
    },
];

describe("ToolSidebar Filtering", () => {
    it("renders filter badges", () => {
        render(<ToolSidebar tools={mockTools} onSelectTool={() => {}} />);

        expect(screen.getByText("service-a")).toBeInTheDocument();
        expect(screen.getByText("service-b")).toBeInTheDocument();
        expect(screen.getByText("#tag1")).toBeInTheDocument();
        expect(screen.getByText("#tag2")).toBeInTheDocument();
        expect(screen.getByText("#tag3")).toBeInTheDocument();
    });

    it("filters by service", () => {
        render(<ToolSidebar tools={mockTools} onSelectTool={() => {}} />);

        const serviceBadge = screen.getByText("service-a");
        fireEvent.click(serviceBadge);

        expect(screen.getByText("tool1")).toBeInTheDocument();
        expect(screen.getByText("tool3")).toBeInTheDocument();
        expect(screen.queryByText("tool2")).not.toBeInTheDocument();
    });

    it("filters by tag", () => {
        render(<ToolSidebar tools={mockTools} onSelectTool={() => {}} />);

        const tagBadge = screen.getByText("#tag2");
        fireEvent.click(tagBadge);

        expect(screen.getByText("tool1")).toBeInTheDocument();
        expect(screen.getByText("tool2")).toBeInTheDocument();
        expect(screen.queryByText("tool3")).not.toBeInTheDocument();
    });

    it("clears filter", () => {
        render(<ToolSidebar tools={mockTools} onSelectTool={() => {}} />);

        const tagBadge = screen.getByText("#tag2");
        fireEvent.click(tagBadge);

        expect(screen.queryByText("tool3")).not.toBeInTheDocument();

        const allBadge = screen.getByText("All");
        fireEvent.click(allBadge);

        expect(screen.getByText("tool1")).toBeInTheDocument();
        expect(screen.getByText("tool2")).toBeInTheDocument();
        expect(screen.getByText("tool3")).toBeInTheDocument();
    });
});
