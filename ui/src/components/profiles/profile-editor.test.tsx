/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor, within } from "@testing-library/react";
import { ProfileEditor } from "./profile-editor";
import { apiClient } from "@/lib/client";
import { describe, it, expect, vi, beforeEach } from "vitest";

// Mock apiClient
vi.mock("@/lib/client", () => ({
    apiClient: {
        listServices: vi.fn(),
    },
}));

// Mock Sonner toast
vi.mock("sonner", () => ({
    toast: {
        error: vi.fn(),
        success: vi.fn(),
    },
}));

// Mock react-virtuoso to render all items without virtualization
vi.mock("react-virtuoso", () => ({
    Virtuoso: ({ data, itemContent, context }: any) => (
        <div>
            {data.map((item: any, index: number) => (
                <div key={index}>
                    {itemContent(index, item, context)}
                </div>
            ))}
        </div>
    ),
}));

describe("ProfileEditor", () => {
    const mockServices = [
        { name: "service-a", tags: ["finance"], version: "1.0.0", httpService: {} },
        { name: "service-b", tags: ["hr"], version: "1.0.0", grpcService: {} },
        { name: "service-c", tags: [], version: "1.0.0", commandLineService: {} },
    ];

    beforeEach(() => {
        vi.clearAllMocks();
        (apiClient.listServices as any).mockResolvedValue(mockServices);
    });

    it("renders correctly for new profile", async () => {
        render(<ProfileEditor profile={null} open={true} onOpenChange={() => {}} onSave={async () => {}} />);

        expect(screen.getByText("Create New Profile")).toBeInTheDocument();
        await waitFor(() => expect(apiClient.listServices).toHaveBeenCalled());
        expect(screen.getByText("service-a")).toBeInTheDocument();
    });

    it("allows adding and removing tags and updates service selection", async () => {
        render(<ProfileEditor profile={null} open={true} onOpenChange={() => {}} onSave={async () => {}} />);

        await waitFor(() => expect(screen.getByText("service-a")).toBeInTheDocument());

        const input = screen.getByPlaceholderText("Add tag (e.g. finance, hr)");

        // Add tag "finance"
        fireEvent.change(input, { target: { value: "finance" } });
        fireEvent.keyDown(input, { key: "Enter" });

        // There should be multiple "finance" texts now (one in service list, one in tag list)
        const finances = screen.getAllByText("finance");
        expect(finances.length).toBeGreaterThanOrEqual(2);

        // Check implicit selection
        // service-a has tag "finance", so it should be auto-selected (disabled checkbox)
        const labelA = screen.getByText("service-a");
        const rowA = labelA.closest("div.flex.items-start");
        // Use querySelector for role checkbox as within() might fail if structure is complex or virtualized (though we mocked it)
        const checkboxA = within(rowA as HTMLElement).getByRole("checkbox");

        expect(checkboxA).toBeDisabled();
        expect(checkboxA).toBeChecked();

        // service-b should not be checked
        const labelB = screen.getByText("service-b");
        const rowB = labelB.closest("div.flex.items-start");
        const checkboxB = within(rowB as HTMLElement).getByRole("checkbox");
        expect(checkboxB).not.toBeChecked();

        // Remove tag "finance"
        // Find the "finance" text node that has a sibling button (the X button) in its parent
        const financeTagNode = screen.getAllByText("finance").find(node => {
            return node.parentElement?.querySelector("button");
        });

        expect(financeTagNode).toBeDefined();
        const removeButton = within(financeTagNode!.parentElement!).getByRole("button");
        fireEvent.click(removeButton);

        // Re-query checkbox A to ensure fresh state
        const labelA2 = screen.getByText("service-a");
        const rowA2 = labelA2.closest("div.flex.items-start");
        const checkboxA2 = within(rowA2 as HTMLElement).getByRole("checkbox");
        expect(checkboxA2).not.toBeChecked();
    });

    it("saves correctly with additional tags", async () => {
        const onSave = vi.fn();
        render(<ProfileEditor profile={null} open={true} onOpenChange={() => {}} onSave={onSave} />);

        // Set name
        fireEvent.change(screen.getByLabelText("Profile Name"), { target: { value: "test-profile" } });

        // Add tag
        const input = screen.getByPlaceholderText("Add tag (e.g. finance, hr)");
        fireEvent.change(input, { target: { value: "finance" } });
        fireEvent.keyDown(input, { key: "Enter" });

        // Save
        fireEvent.click(screen.getByText("Save Profile"));

        await waitFor(() => expect(onSave).toHaveBeenCalledWith({
            name: "test-profile",
            selector: {
                tags: ["dev", "finance"] // Default type is dev
            },
            serviceConfig: {},
            secrets: {}
        }));
    });
});
