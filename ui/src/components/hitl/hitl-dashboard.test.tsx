/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { HitlDashboard } from "./hitl-dashboard";

describe("HitlDashboard", () => {
    it("renders pending approvals and handles actions", () => {
        render(<HitlDashboard />);
        expect(screen.getByText("database.drop_table")).toBeInTheDocument();

        const approveBtn = screen.getByText("Approve");
        fireEvent.click(approveBtn);

        expect(screen.getByText("Status: approved")).toBeInTheDocument();
    });
});
