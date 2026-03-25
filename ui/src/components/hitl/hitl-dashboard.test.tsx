/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { HitlDashboard } from "./hitl-dashboard";
import React from "react";

describe("HitlDashboard", () => {
    it("renders pending approvals and handles actions", () => {
        render(<HitlDashboard />);
        expect(screen.getByText("database.drop_table")).toBeInTheDocument();
        expect(screen.getByText("system.restart")).toBeInTheDocument();

        // Approve the action that does not require MFA
        const approveBtns = screen.getAllByText("Approve");
        fireEvent.click(approveBtns[1]);

        expect(screen.getByText("Status: approved")).toBeInTheDocument();
    });

    it("prompts for MFA when approving a high-risk action", async () => {
        render(<HitlDashboard />);
        const approveBtns = screen.getAllByText("Approve");

        // Approve the action that requires MFA
        fireEvent.click(approveBtns[0]);

        // Dialog should be visible
        expect(screen.getByText("MFA Required")).toBeInTheDocument();

        const mfaInput = screen.getByPlaceholderText("123456");
        fireEvent.change(mfaInput, { target: { value: "123456" } });

        const verifyBtn = screen.getByText("Verify & Approve");
        fireEvent.click(verifyBtn);

        await waitFor(() => {
            // First item should now show approved status
            const statuses = screen.getAllByText("Status: approved");
            expect(statuses.length).toBeGreaterThan(0);
        });
    });
});
