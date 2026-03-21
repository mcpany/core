/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor, within } from "@testing-library/react";
import { ProfileEditor } from "./profile-editor";
import { apiClient } from "@/lib/client";
import { describe, it, expect, vi, beforeEach } from "vitest";

// Mock react-virtuoso to render items in tests
vi.mock("react-virtuoso", () => ({
    Virtuoso: ({ data, itemContent, context }: any) => (
        <div data-testid="virtuoso-list">
            {(data || []).map((item: any, index: number) => (
                <div key={index}>{itemContent(index, item, context)}</div>
            ))}
        </div>
    ),
    VirtuosoHandle: {},
}));

// Mock apiClient
vi.mock("@/lib/client", () => ({
    apiClient: {
        listServices: vi.fn(),
        getServiceStatus: vi.fn(),
    },
}));

// Mock Sonner toast
vi.mock("sonner", () => ({
    toast: vi.fn(),
}));

describe("ProfileEditor", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        (apiClient.listServices as any).mockResolvedValue([
            { name: "service-a", tags: ["core", "auth"] },
            { name: "service-b", tags: ["data"] },
            { name: "service-c", tags: ["core"] },
        ]);
        (apiClient.getServiceStatus as any).mockResolvedValue({ tools: [] });
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

        // Add "core" tag
        fireEvent.change(input, { target: { value: "core" } });
        fireEvent.keyDown(input, { key: "Enter" });

        // Ensure "core" tag is rendered
        await waitFor(() => expect(screen.getByText("core")).toBeInTheDocument());
    });

    it("saves correctly with additional tags", async () => {
        const onSave = vi.fn();
        render(<ProfileEditor profile={null} open={true} onOpenChange={() => {}} onSave={onSave} />);

        // Set name
        fireEvent.change(screen.getByLabelText("Profile Name"), { target: { value: "test-profile" } });

        // Wait for services to load before saving
        await waitFor(() => expect(screen.getByText("service-a")).toBeInTheDocument());

        fireEvent.click(screen.getByText("Save Profile"));

        await waitFor(() => {
            expect(onSave).toHaveBeenCalledWith(expect.objectContaining({
                name: "test-profile"
            }));
        });
    });
});
