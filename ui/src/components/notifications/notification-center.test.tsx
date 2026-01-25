import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { NotificationCenter } from "./notification-center";
import { apiClient } from "@/lib/client";
import { describe, it, expect, vi, beforeEach, type Mock } from 'vitest';
import React from 'react';
import "@testing-library/jest-dom";

// Mock apiClient
vi.mock("@/lib/client", () => ({
  apiClient: {
    listAlerts: vi.fn(),
    updateAlertStatus: vi.fn(),
  },
}));

// Mock ResizeObserver for Radix UI
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock ScrollArea to avoid layout issues in test
vi.mock("@/components/ui/scroll-area", () => ({
  ScrollArea: ({ children }: { children: React.ReactNode }) => <div>{children}</div>
}));

describe("NotificationCenter", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the bell icon", () => {
    (apiClient.listAlerts as Mock).mockResolvedValue([]);
    render(<NotificationCenter />);
    expect(screen.getByLabelText("Notifications")).toBeInTheDocument();
  });

  it("shows badge when there are active alerts", async () => {
    (apiClient.listAlerts as Mock).mockResolvedValue([
      { id: "1", status: "active", title: "Test Alert", timestamp: new Date().toISOString(), severity: "critical", service: "test" }
    ]);

    render(<NotificationCenter />);

    // Wait for the badge to appear
    await waitFor(() => {
       const button = screen.getByLabelText("Notifications");
       // Check for the animated span using querySelector
       // Note: JSDOM might not parse tailwind classes perfectly but structure should be there
       const badge = button.querySelector('span.animate-ping');
       expect(badge).toBeInTheDocument();
    });
  });

  it("opens popover and lists alerts", async () => {
    const alert = { id: "1", status: "active", title: "Test Alert", message: "Something wrong", timestamp: new Date().toISOString(), severity: "critical", service: "test" };
    (apiClient.listAlerts as Mock).mockResolvedValue([alert]);

    render(<NotificationCenter />);

    const button = screen.getByLabelText("Notifications");
    fireEvent.click(button);

    await waitFor(() => {
      expect(screen.getByText("Test Alert")).toBeInTheDocument();
      expect(screen.getByText("Something wrong")).toBeInTheDocument();
    });
  });

  it("marks alert as read", async () => {
    const alert = { id: "1", status: "active", title: "Test Alert", message: "Something wrong", timestamp: new Date().toISOString(), severity: "critical", service: "test" };
    (apiClient.listAlerts as Mock).mockResolvedValue([alert]);
    (apiClient.updateAlertStatus as Mock).mockResolvedValue({});

    render(<NotificationCenter />);

    const button = screen.getByLabelText("Notifications");
    fireEvent.click(button);

    await waitFor(() => screen.getByText("Test Alert"));

    const checkBtn = screen.getByTitle("Mark as read");
    fireEvent.click(checkBtn);

    expect(apiClient.updateAlertStatus).toHaveBeenCalledWith("1", "acknowledged");
  });
});
