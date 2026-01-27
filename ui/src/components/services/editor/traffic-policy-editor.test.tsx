/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { TrafficPolicyEditor } from "./traffic-policy-editor";
import { UpstreamServiceConfig } from "@/lib/client";
import { vi } from "vitest";

// Mock Duration Utils if needed, but we can use real ones
// We need to mock the UI components if they are complex, but Card/Input are simple enough?
// Actually, Card/Input might use Context or other things.
// Let's assume they are fine in JSDOM.

describe("TrafficPolicyEditor", () => {
    const mockService: UpstreamServiceConfig = {
        name: "test-service",
        id: "123",
        sanitizedName: "test-service",
        version: "1.0.0",
        priority: 0,
        disable: false,
        autoDiscoverTool: false,
        configError: "",
        readOnly: false,
        lastError: "",
        toolCount: 0,
        loadBalancingStrategy: 0,
        callPolicies: [],
        preCallHooks: [],
        postCallHooks: [],
        prompts: [],
        tags: [],
        configurationSchema: "",
        rateLimit: {
            isEnabled: false,
            requestsPerSecond: 0,
            burst: 0,
            storage: 0,
            keyBy: 0,
            costMetric: 0,
            toolLimits: {}
        }
    };

    it("should render inputs", () => {
        render(<TrafficPolicyEditor service={mockService} onChange={() => {}} />);
        expect(screen.getByText("Rate Limiting")).toBeInTheDocument();
        expect(screen.getByLabelText("Enable Rate Limiting")).toBeInTheDocument();
    });

    it("should call onChange when rate limit is toggled", () => {
        const handleChange = vi.fn();
        render(<TrafficPolicyEditor service={mockService} onChange={handleChange} />);

        fireEvent.click(screen.getByLabelText("Enable Rate Limiting"));

        expect(handleChange).toHaveBeenCalledWith(expect.objectContaining({
            rateLimit: expect.objectContaining({
                isEnabled: true
            })
        }));
    });

    it("should call onChange when requests per second changes", () => {
        const handleChange = vi.fn();
        // Enable it first so input is not disabled (logic in component)
        const enabledService = {
            ...mockService,
            rateLimit: { ...mockService.rateLimit!, isEnabled: true }
        };
        render(<TrafficPolicyEditor service={enabledService} onChange={handleChange} />);

        fireEvent.change(screen.getByLabelText("Requests / Sec"), { target: { value: "10" } });

        expect(handleChange).toHaveBeenCalledWith(expect.objectContaining({
            rateLimit: expect.objectContaining({
                requestsPerSecond: 10
            })
        }));
    });
});
