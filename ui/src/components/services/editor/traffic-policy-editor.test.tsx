/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { render, screen, fireEvent } from "@testing-library/react";
import { TrafficPolicyEditor } from "./traffic-policy-editor";
import { UpstreamServiceConfig } from "@/lib/client";
import { vi } from 'vitest';

// Mock the Service object
const mockService: UpstreamServiceConfig = {
    id: "test-service",
    name: "test-service",
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
    // Relevant fields
    rateLimit: {
        isEnabled: false,
        requestsPerSecond: 10,
        burst: 20 as any
    },
    connectionPool: {
        maxConnections: 100,
        maxIdleConnections: 10,
        idleTimeout: "30s" as any
    },
    resilience: {
        timeout: "1m" as any,
        circuitBreaker: {
            failureRateThreshold: 0.5,
            consecutiveFailures: 5,
            openDuration: "10s" as any,
            halfOpenRequests: 1
        },
        retryPolicy: {
            numberOfRetries: 3,
            baseBackoff: "100ms" as any,
            maxBackoff: "1s" as any,
            maxElapsedTime: "5s" as any
        }
    }
};

describe("TrafficPolicyEditor", () => {
    it("renders all sections", () => {
        const handleChange = vi.fn();
        render(<TrafficPolicyEditor service={mockService} onChange={handleChange} />);

        expect(screen.getByText("Rate Limiting")).toBeInTheDocument();
        expect(screen.getByText("Connection Pool")).toBeInTheDocument();
        expect(screen.getByText("Resilience")).toBeInTheDocument();
    });

    it("handles Rate Limit updates", () => {
        const handleChange = vi.fn();
        render(<TrafficPolicyEditor service={mockService} onChange={handleChange} />);

        // Toggle Switch
        const switchBtn = screen.getByLabelText("Enable Rate Limiting");
        fireEvent.click(switchBtn);
        expect(handleChange).toHaveBeenCalledWith({
            rateLimit: expect.objectContaining({ isEnabled: true })
        });

        // Update RPS
        // We need to re-render or assume component is stateless regarding input value if passed via props?
        // It is controlled by props.
        const input = screen.getByLabelText("Requests Per Second");
        fireEvent.change(input, { target: { value: "50" } });
        expect(handleChange).toHaveBeenCalledWith({
            rateLimit: expect.objectContaining({ requestsPerSecond: 50 })
        });
    });

    it("handles Connection Pool updates", () => {
        const handleChange = vi.fn();
        render(<TrafficPolicyEditor service={mockService} onChange={handleChange} />);

        const input = screen.getByLabelText("Max Connections");
        fireEvent.change(input, { target: { value: "200" } });
        expect(handleChange).toHaveBeenCalledWith({
            connectionPool: expect.objectContaining({ maxConnections: 200 })
        });
    });

    it("handles Resilience updates (Circuit Breaker)", () => {
        const handleChange = vi.fn();
        render(<TrafficPolicyEditor service={mockService} onChange={handleChange} />);

        const input = screen.getByLabelText("Failure Threshold (%)");
        fireEvent.change(input, { target: { value: "0.8" } });
        expect(handleChange).toHaveBeenCalledWith({
            resilience: expect.objectContaining({
                circuitBreaker: expect.objectContaining({ failureRateThreshold: 0.8 })
            })
        });
    });

    it("handles Resilience updates (Timeout)", () => {
        const handleChange = vi.fn();
        render(<TrafficPolicyEditor service={mockService} onChange={handleChange} />);

        const input = screen.getByLabelText("Request Timeout");
        fireEvent.change(input, { target: { value: "45s" } });
        expect(handleChange).toHaveBeenCalledWith({
            resilience: expect.objectContaining({ timeout: "45s" })
        });
    });
});
