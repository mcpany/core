/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { HitlDashboard } from "./hitl-dashboard";
import { apiClient } from "@/lib/client";

vi.mock("@/lib/client", () => ({
    apiClient: {
        getHitlApprovals: vi.fn(),
        actionHitlApproval: vi.fn(),
    }
}));

describe("HitlDashboard", () => {
    beforeEach(() => {
        vi.mocked(apiClient.getHitlApprovals).mockResolvedValue([
            { id: "1", tool: "database.drop_table", intent: "Pending verification for sensitive tool", status: "pending", requireMfa: true },
            { id: "2", tool: "aws.terminate_instance", intent: "Pending verification for sensitive tool", status: "pending", requireMfa: false }
        ]);
        vi.mocked(apiClient.actionHitlApproval).mockResolvedValue();
    });

    afterEach(() => {
        vi.restoreAllMocks();
    });

    it("renders pending approvals and handles actions without MFA", async () => {
        render(<HitlDashboard />);
        expect(await screen.findByText("aws.terminate_instance")).toBeInTheDocument();

        // Get the second approve button which is for aws.terminate_instance (requireMfa: false)
        const approveBtns = screen.getAllByText("Approve");
        fireEvent.click(approveBtns[1]);

        expect(apiClient.actionHitlApproval).toHaveBeenCalledWith("2", "approved", undefined);
    });

    it("renders pending approvals and handles actions with MFA", async () => {
        render(<HitlDashboard />);
        expect(await screen.findByText("database.drop_table")).toBeInTheDocument();

        // Get the first approve button which is for database.drop_table (requireMfa: true)
        const approveBtns = screen.getAllByText("Approve");
        fireEvent.click(approveBtns[0]);

        // Should open MFA dialog
        expect(screen.getByText("Multi-Factor Authentication")).toBeInTheDocument();

        // Enter MFA code
        const mfaInput = screen.getByPlaceholderText("Enter 6-digit code");
        fireEvent.change(mfaInput, { target: { value: "123456" } });

        // Submit
        const verifyBtn = screen.getByText("Verify & Approve");
        fireEvent.click(verifyBtn);

        expect(apiClient.actionHitlApproval).toHaveBeenCalledWith("1", "approved", "123456");
        expect(screen.queryByText("Multi-Factor Authentication")).not.toBeInTheDocument();
    });
});
