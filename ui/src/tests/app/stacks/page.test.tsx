/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import StacksPage from "@/app/stacks/page";
import { marketplaceService } from "@/lib/marketplace-service";

// Mock dependencies
vi.mock("@/lib/marketplace-service", () => ({
  marketplaceService: {
    fetchLocalCollections: vi.fn(),
    saveLocalCollection: vi.fn(),
    deleteLocalCollection: vi.fn(),
  },
}));

vi.mock("@/hooks/use-toast", () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}));

// Mock Link since next/link can be tricky in tests
vi.mock("next/link", () => ({
  default: ({ children, href }: { children: React.ReactNode; href: string }) => (
    <a href={href}>{children}</a>
  ),
}));

describe("StacksPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders empty state when no stacks exist", () => {
    (marketplaceService.fetchLocalCollections as any).mockReturnValue([]);
    render(<StacksPage />);
    expect(screen.getByText("No Stacks Found")).toBeInTheDocument();
    expect(screen.getByText("Create a new stack to get started.")).toBeInTheDocument();
  });

  it("renders list of stacks", () => {
    (marketplaceService.fetchLocalCollections as any).mockReturnValue([
      { name: "Stack1", description: "Desc1", services: [] },
      { name: "Stack2", description: "Desc2", services: [] },
    ]);
    render(<StacksPage />);
    expect(screen.getByText("Stack1")).toBeInTheDocument();
    expect(screen.getByText("Stack2")).toBeInTheDocument();
  });

  it("opens create dialog and creates a stack", async () => {
    (marketplaceService.fetchLocalCollections as any).mockReturnValue([]);
    render(<StacksPage />);

    // There might be two buttons (header and empty state), click the first one
    const createBtn = screen.getAllByText("Create Stack", { selector: 'button' })[0];
    fireEvent.click(createBtn);

    const nameInput = screen.getByLabelText("Name");
    fireEvent.change(nameInput, { target: { value: "NewStack" } });

    const descInput = screen.getByLabelText("Description");
    fireEvent.change(descInput, { target: { value: "New Desc" } });

    const submitBtn = screen.getByText("Create", { selector: 'button' });
    fireEvent.click(submitBtn);

    expect(marketplaceService.saveLocalCollection).toHaveBeenCalledWith(expect.objectContaining({
        name: "NewStack",
        description: "New Desc"
    }));
  });
});
