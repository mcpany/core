/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { ResilienceEditor } from "./resilience-editor";
import { describe, it, expect, vi } from "vitest";

describe("ResilienceEditor", () => {
    const mockOnChange = vi.fn();
    const defaultResilience = {
        timeout: "30s",
        circuitBreaker: {
            failureRateThreshold: 0.5,
            consecutiveFailures: 5,
            openDuration: "60s",
            halfOpenRequests: 1
        },
        retryPolicy: {
            numberOfRetries: 3,
            baseBackoff: "100ms",
            maxBackoff: "5s",
            maxElapsedTime: "10s"
        }
    };

    it("renders all fields with correct values", () => {
        render(<ResilienceEditor resilience={defaultResilience as any} onChange={mockOnChange} />);

        expect(screen.getByLabelText(/Global Timeout/i)).toHaveValue("30s");
        expect(screen.getByLabelText(/Consecutive Failures/i)).toHaveValue(5);
        expect(screen.getByLabelText(/Open Duration/i)).toHaveValue("60s");
        expect(screen.getByLabelText(/Number of Retries/i)).toHaveValue(3);
        expect(screen.getByLabelText(/Base Backoff/i)).toHaveValue("100ms");
    });

    it("calls onChange when fields are updated", () => {
        render(<ResilienceEditor resilience={defaultResilience as any} onChange={mockOnChange} />);

        const timeoutInput = screen.getByLabelText(/Global Timeout/i);
        fireEvent.change(timeoutInput, { target: { value: "60s" } });

        expect(mockOnChange).toHaveBeenCalledWith(expect.objectContaining({
            timeout: "60s"
        }));

        const retriesInput = screen.getByLabelText(/Number of Retries/i);
        fireEvent.change(retriesInput, { target: { value: "10" } });

        expect(mockOnChange).toHaveBeenCalledWith(expect.objectContaining({
            retryPolicy: expect.objectContaining({
                numberOfRetries: 10
            })
        }));
    });
});
