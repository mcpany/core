/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, waitFor } from "@testing-library/react";
import { ServiceConfigDiff } from "./service-config-diff";
import { UpstreamServiceConfig } from "@/lib/types";
import { vi, describe, it, expect } from "vitest";

// Mock Monaco Editor
vi.mock("@monaco-editor/react", () => ({
  DiffEditor: ({ original, modified }: { original: string, modified: string }) => (
    <div data-testid="diff-editor">
      <div data-testid="original">{original}</div>
      <div data-testid="modified">{modified}</div>
    </div>
  ),
}));

// Mock useTheme
vi.mock("next-themes", () => ({
  useTheme: () => ({ theme: "light", systemTheme: "light" }),
}));

const mockServiceOriginal: UpstreamServiceConfig = {
    id: "s1",
    name: "Service 1",
    version: "1.0",
    disable: false,
    priority: 0,
    loadBalancingStrategy: 0,
    tags: ["prod"],
    sanitizedName: "service-1",
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    prompts: [],
    autoDiscoverTool: false,
    configError: "",
    readOnly: false,
    httpService: {
        address: "http://localhost:8080",
        tools: [],
        calls: {},
        resources: [],
        prompts: [],
        healthCheck: undefined,
        tlsConfig: undefined
    }
};

const mockServiceModified: UpstreamServiceConfig = {
    ...mockServiceOriginal,
    name: "Service 1 Updated",
    httpService: {
        ...mockServiceOriginal.httpService!,
        address: "http://localhost:9090"
    }
};

describe("ServiceConfigDiff", () => {
  it("renders the diff editor with YAML content", async () => {
    render(<ServiceConfigDiff original={mockServiceOriginal} modified={mockServiceModified} />);

    // Expect loading state first
    expect(screen.getByText("Loading Editor...")).toBeInTheDocument();

    // Wait for dynamic import to resolve
    await waitFor(() => {
        expect(screen.getByTestId("diff-editor")).toBeInTheDocument();
    });

    // Check if content contains YAML formatted strings
    const originalContent = screen.getByTestId("original").textContent;
    const modifiedContent = screen.getByTestId("modified").textContent;

    expect(originalContent).toContain('name: Service 1');
    expect(originalContent).toContain('address: http://localhost:8080');

    expect(modifiedContent).toContain('name: Service 1 Updated');
    expect(modifiedContent).toContain('address: http://localhost:9090');
  });
});
