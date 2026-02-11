/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import WebhooksPage from "./page";
import { vi, describe, it, expect, beforeEach } from "vitest";
import { AuditConfig_StorageType } from "@proto/config/v1/config";

// Mock apiClient
vi.mock("@/lib/client", () => ({
  apiClient: {
    getGlobalSettings: vi.fn(),
    saveGlobalSettings: vi.fn(),
  },
}));

// Mock Sonner Toast
vi.mock("@/hooks/use-toast", () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

import { apiClient } from "@/lib/client";

describe("WebhooksPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the page and loads settings", async () => {
    (apiClient.getGlobalSettings as any).mockResolvedValue({
      alerts: { enabled: true, webhookUrl: "https://alerts.com" },
      audit: { enabled: false, webhookUrl: "", storageType: 0 },
    });

    render(<WebhooksPage />);

    // Wait for loading to finish
    await waitFor(() => {
        expect(screen.getByText("System Webhooks")).toBeInTheDocument();
    });

    const alertInput = screen.getByDisplayValue("https://alerts.com");
    expect(alertInput).toBeInTheDocument();
  });

  it("updates audit settings and saves", async () => {
    (apiClient.getGlobalSettings as any).mockResolvedValue({
      alerts: { enabled: false, webhookUrl: "" },
      audit: { enabled: false, webhookUrl: "", storageType: 0 },
    });

    render(<WebhooksPage />);

    await waitFor(() => {
      expect(screen.getByText("Audit Logging")).toBeInTheDocument();
    });

    // Enable Audit
    const auditSwitch = screen.getByTestId("audit-switch");
    fireEvent.click(auditSwitch);

    // Set URL
    const auditInput = screen.getByPlaceholderText("https://audit-logs.example.com/...");
    fireEvent.change(auditInput, { target: { value: "https://audit.com" } });

    // Save
    const saveButton = screen.getByText("Save Changes");
    fireEvent.click(saveButton);

    await waitFor(() => {
      expect(apiClient.saveGlobalSettings).toHaveBeenCalledWith(expect.objectContaining({
        audit: expect.objectContaining({
          enabled: true,
          webhookUrl: "https://audit.com",
          storageType: AuditConfig_StorageType.STORAGE_TYPE_WEBHOOK,
        })
      }));
    });
  });
});
