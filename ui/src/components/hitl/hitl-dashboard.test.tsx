/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { HitlDashboard } from "./hitl-dashboard";
import { apiClient } from "@/lib/client";
import { vi, describe, it, expect, beforeEach, afterEach } from "vitest";

// Mock the apiClient
vi.mock("@/lib/client", () => {
    return {
        apiClient: {
            getHITLApprovals: vi.fn(),
            resolveHITLApproval: vi.fn()
        }
    };
});

// Mock the useToast hook
vi.mock("@/hooks/use-toast", () => {
    return {
        useToast: () => ({
            toast: vi.fn()
        })
    };
});

describe("HitlDashboard", () => {
    beforeEach(() => {
        vi.mocked(apiClient.getHITLApprovals).mockResolvedValue([
            { id: "1", tool: "database.drop_table", intent: "Pending verification for sensitive tool", status: "pending", requireMfa: true },
            { id: "2", tool: "aws.terminate_instance", intent: "Pending verification for sensitive tool", status: "pending", requireMfa: false }
        ]);
        vi.mocked(apiClient.resolveHITLApproval).mockResolvedValue(undefined);
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it("renders pending approvals and handles actions without MFA", async () => {
        render(<HitlDashboard />);
        expect(await screen.findByText("aws.terminate_instance")).toBeInTheDocument();

        // Get the second approve button which is for aws.terminate_instance (requireMfa: false)
        // Since there are 2 approve buttons (one for db, one for aws), we need to select the right one.
        // The first card is db.drop_table, the second is aws.terminate_instance
        const approveBtns = await screen.findAllByText("Approve");
        fireEvent.click(approveBtns[1]);

        expect(apiClient.resolveHITLApproval).toHaveBeenCalledWith("2", "approved", undefined);
    });

    it("renders pending approvals and handles actions with MFA", async () => {
        render(<HitlDashboard />);
        expect(await screen.findByText("database.drop_table")).toBeInTheDocument();

        // Get the first approve button which is for database.drop_table (requireMfa: true)
        const approveBtns = await screen.findAllByText("Approve");
        fireEvent.click(approveBtns[0]);

        // Should open MFA dialog
        expect(await screen.findByText("Multi-Factor Authentication Required")).toBeInTheDocument();

        // Enter MFA code
        const mfaInput = screen.getByPlaceholderText("MFA Code");
        fireEvent.change(mfaInput, { target: { value: "123456" } });

        // Submit
        const verifyBtn = screen.getByText("Verify & Approve");
        fireEvent.click(verifyBtn);

        expect(apiClient.resolveHITLApproval).toHaveBeenCalledWith("1", "approved", "123456");
        // Verify dialog closes or is not present
        expect(screen.queryByText("Multi-Factor Authentication Required")).not.toBeInTheDocument();
    });
});
