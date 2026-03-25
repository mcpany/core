/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { ConnectClientButton } from "./connect-client-button";
import React from "react";
import { vi } from "vitest";

// Mock ResizeObserver which is used by some UI components but not in JSDOM
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock clipboard
Object.assign(navigator, {
  clipboard: {
    writeText: vi.fn(),
  },
});

// Mock matchMedia
window.matchMedia =
  window.matchMedia ||
  function () {
    return {
      matches: false,
      addListener: function () {},
      removeListener: function () {},
    };
  };

describe("ConnectClientButton", () => {
  it("renders the connect button", () => {
    render(<ConnectClientButton />);
    // "Connect" text might be hidden on mobile but we are testing in JSDOM default (usually desktop size or generic)
    // The button has "Connect" text.
    const button = screen.getByText("Connect");
    expect(button).toBeInTheDocument();
  });

  it("opens dialog when clicked", () => {
    render(<ConnectClientButton />);
    const button = screen.getByText("Connect");
    fireEvent.click(button);
    expect(screen.getByText("Connect to MCP Any")).toBeInTheDocument();
    // Default tab is Claude
    expect(
      screen.getByText("Claude Desktop Configuration"),
    ).toBeInTheDocument();
  });

  it("allows API key input", async () => {
    render(<ConnectClientButton />);
    fireEvent.click(screen.getByText("Connect"));
    const input = screen.getByPlaceholderText("Optional (if configured)");
    fireEvent.change(input, { target: { value: "my-secret-key" } });
    expect(input).toHaveValue("my-secret-key");

    // Switch to Cursor tab where the URL is rendered as a read-only input
    // Radix UI Tabs uses onMouseDown to switch tabs, not onClick
    const cursorTab = screen.getByText("Cursor");
    fireEvent.mouseDown(cursorTab, { button: 0 });

    // Check if the URL contains the api key (wait for Cursor tab content to mount)
    await waitFor(() => {
      const urlInput = screen.getByDisplayValue(/api_key=my-secret-key/);
      expect(urlInput).toBeInTheDocument();
    });
  });
});
