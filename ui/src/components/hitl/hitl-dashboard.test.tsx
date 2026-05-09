/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { HitlDashboard } from "./hitl-dashboard";

describe("HitlDashboard", () => {
    beforeEach(() => {
        global.fetch = vi.fn((url, options) => {
            if (url === "/api/v1/hitl/approvals" && (!options || options.method === "GET")) {
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve([
                        { id: "1", tool: "database.drop_table", intent: "Pending verification for sensitive tool", status: "pending", requireMfa: true },
                        { id: "2", tool: "aws.terminate_instance", intent: "Pending verification for sensitive tool", status: "pending", requireMfa: false }
                    ])
                });
            }
            if (url.startsWith("/api/v1/hitl/approvals/") && options?.method === "POST") {
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({})
                });
            }
            return Promise.reject("Not mocked");
        }) as any;
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

        expect(global.fetch).toHaveBeenCalledWith("/api/v1/hitl/approvals/2", expect.objectContaining({
            method: "POST",
            body: JSON.stringify({ action: "approved", mfaCode: "" })
        }));
    });

    it("renders pending approvals and handles actions with MFA", async () => {
        render(<HitlDashboard />);
        expect(await screen.findByText("database.drop_table")).toBeInTheDocument();

        // Get the first approve button which is for database.drop_table (requireMfa: true)
        const approveBtns = screen.getAllByText("Approve");
        fireEvent.click(approveBtns[0]);

        // Should open MFA dialog
        expect(screen.getByText("Multi-Factor Authentication Required")).toBeInTheDocument();

        // Enter MFA code
        const mfaInput = screen.getByPlaceholderText("MFA Code");
        fireEvent.change(mfaInput, { target: { value: "123456" } });

        // Submit
        const verifyBtn = screen.getByText("Verify & Approve");
        fireEvent.click(verifyBtn);

        expect(global.fetch).toHaveBeenCalledWith("/api/v1/hitl/approvals/1", expect.objectContaining({
            method: "POST",
            body: JSON.stringify({ action: "approved", mfaCode: "123456" })
        }));
        expect(screen.queryByText("Multi-Factor Authentication Required")).not.toBeInTheDocument();
    });
});
