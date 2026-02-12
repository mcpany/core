/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import WebhooksPage from "./page";
import { apiClient } from "@/lib/client";
import { vi, describe, it, expect, beforeEach } from "vitest";

// Mock apiClient
vi.mock("@/lib/client", () => ({
  apiClient: {
    getGlobalSettings: vi.fn(),
    saveGlobalSettings: vi.fn(),
  },
}));

// Mock toast
vi.mock("@/hooks/use-toast", () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

describe("WebhooksPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("loads and displays settings correctly", async () => {
    (apiClient.getGlobalSettings as any).mockResolvedValue({
      alerts: { enabled: true, webhook_url: "https://alerts.com" },
      audit: { enabled: true, storage_type: 4, webhook_url: "https://audit.com" },
    });

    render(<WebhooksPage />);

    // Wait for loading to finish
    await waitFor(() => expect(screen.queryByText("Loading configuration...")).not.toBeInTheDocument());

    expect(screen.getByLabelText(/Enable Alerts Webhook/i)).toBeChecked();
    expect(screen.getByDisplayValue("https://alerts.com")).toBeInTheDocument();

    expect(screen.getByLabelText(/Enable Audit Stream/i)).toBeChecked();
    expect(screen.getByDisplayValue("https://audit.com")).toBeInTheDocument();
  });

  it("handles save correctly", async () => {
    const initialSettings = {
        alerts: { enabled: false, webhook_url: "" },
        audit: { enabled: false, storage_type: 0, webhook_url: "" },
        other: "data"
    };
    (apiClient.getGlobalSettings as any).mockResolvedValue(initialSettings);

    render(<WebhooksPage />);
    await waitFor(() => expect(screen.queryByText("Loading configuration...")).not.toBeInTheDocument());

    // Enable Alerts and set URL
    const alertsSwitch = screen.getByLabelText(/Enable Alerts Webhook/i);
    fireEvent.click(alertsSwitch);

    const alertsInput = screen.getByPlaceholderText("https://api.example.com/webhooks/alerts");
    fireEvent.change(alertsInput, { target: { value: "https://my-alerts.com" } });

    // Save
    const saveButton = screen.getByText("Save Changes");
    fireEvent.click(saveButton);

    await waitFor(() => {
        expect(apiClient.saveGlobalSettings).toHaveBeenCalledWith(expect.objectContaining({
            alerts: expect.objectContaining({
                enabled: true
            })
        }));
    });
  });
});
