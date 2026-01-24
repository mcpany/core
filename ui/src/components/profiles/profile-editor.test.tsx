/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { ProfileEditor } from "./profile-editor";
import { vi, describe, it, expect } from "vitest";
import { UpstreamServiceConfig } from "@/lib/client";

// Mock ResizeObserver for ScrollArea
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

describe("ProfileEditor", () => {
    const mockServices: UpstreamServiceConfig[] = [
        { name: "weather-service", version: "1.0", httpService: { address: "http://localhost:8080" } } as any,
        { name: "calculator-service", version: "2.0", mcpService: { } } as any
    ];

    it("renders create mode correctly", () => {
        const onSave = vi.fn();
        const onCancel = vi.fn();

        render(
            <ProfileEditor
                services={mockServices}
                onSave={onSave}
                onCancel={onCancel}
            />
        );

        expect(screen.getByLabelText("Profile Name")).toBeInTheDocument();
        expect(screen.getByLabelText("Profile Name")).toHaveValue("");
        expect(screen.getByLabelText("Tags (Comma separated)")).toHaveValue("dev");
        expect(screen.getByText("weather-service")).toBeInTheDocument();
        expect(screen.getByText("calculator-service")).toBeInTheDocument();

        // Save button should be disabled initially (no name)
        const saveBtn = screen.getByText("Save Profile");
        expect(saveBtn).toBeDisabled();
    });

    it("renders edit mode correctly", () => {
        const profile = {
            name: "test-profile",
            selector: { tags: ["prod", "secure"] },
            serviceConfig: {
                "weather-service": { enabled: true }
            }
        };

        render(
            <ProfileEditor
                profile={profile}
                services={mockServices}
                onSave={vi.fn()}
                onCancel={vi.fn()}
            />
        );

        expect(screen.getByLabelText("Profile Name")).toHaveValue("test-profile");
        expect(screen.getByLabelText("Profile Name")).toBeDisabled(); // Name is immutable
        expect(screen.getByLabelText("Tags (Comma separated)")).toHaveValue("prod, secure");

        // weather-service should be checked
        // Note: Shadcn Checkbox uses aria-checked
        const weatherCheckbox = screen.getByRole("checkbox", { name: "weather-service" });
        expect(weatherCheckbox).toHaveAttribute("aria-checked", "true");

        // calculator-service should be unchecked
        const calculatorCheckbox = screen.getByRole("checkbox", { name: "calculator-service" });
        expect(calculatorCheckbox).toHaveAttribute("aria-checked", "false");
    });

    it("submits correct data on save", () => {
        const onSave = vi.fn();

        render(
            <ProfileEditor
                services={mockServices}
                onSave={onSave}
                onCancel={vi.fn()}
            />
        );

        // Enter name
        fireEvent.change(screen.getByLabelText("Profile Name"), { target: { value: "new-profile" } });

        // Enter tags
        fireEvent.change(screen.getByLabelText("Tags (Comma separated)"), { target: { value: "demo, test" } });

        // Check weather-service
        const weatherCheckbox = screen.getByRole("checkbox", { name: "weather-service" });
        fireEvent.click(weatherCheckbox);

        // Click Save
        fireEvent.click(screen.getByText("Save Profile"));

        expect(onSave).toHaveBeenCalledWith({
            name: "new-profile",
            selector: {
                tags: ["demo", "test"]
            },
            serviceConfig: {
                "weather-service": { enabled: true }
            }
        });
    });

    it("searches and filters services", () => {
        render(
            <ProfileEditor
                services={mockServices}
                onSave={vi.fn()}
                onCancel={vi.fn()}
            />
        );

        // Search for "calc"
        fireEvent.change(screen.getByPlaceholderText("Search services..."), { target: { value: "calc" } });

        expect(screen.getByText("calculator-service")).toBeInTheDocument();
        expect(screen.queryByText("weather-service")).not.toBeInTheDocument();
    });
});
