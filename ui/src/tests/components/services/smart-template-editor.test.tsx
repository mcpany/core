/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from 'react';
import { render, screen, waitFor } from "@testing-library/react";
import { SmartTemplateEditor } from "@/components/services/editor/smart-template-editor";
import { describe, it, expect } from "vitest";

// Mock resize observer if not global (it is global in setup.ts)

describe("SmartTemplateEditor", () => {
    it("renders correctly with default props", () => {
        render(<SmartTemplateEditor value="" onChange={() => {}} />);
        expect(screen.getByText("Template")).toBeInTheDocument();
        expect(screen.getByText("Test Data (JSON)")).toBeInTheDocument();
        expect(screen.getByText("Live Preview")).toBeInTheDocument();
    });

    it("renders template with initial test data", async () => {
        const template = "Hello {{ name }}!";
        const testData = JSON.stringify({ name: "World" });

        render(
            <SmartTemplateEditor
                value={template}
                onChange={() => {}}
                testData={testData}
            />
        );

        // Nunjucks rendering is synchronous but state update might be async or immediate
        // Using waitFor to be safe
        await waitFor(() => {
            // Check for the preview content. It's inside a pre tag.
            // Using a loose match since it might be wrapped
            expect(screen.getByText((content) => content.includes("Hello World!"))).toBeInTheDocument();
        });
    });

    it("displays error for invalid JSON test data", async () => {
        const template = "Hello {{ name }}!";
        // Invalid JSON
        const testData = "{ name: 'World' ";

        render(
            <SmartTemplateEditor
                value={template}
                onChange={() => {}}
                testData={testData}
            />
        );

        await waitFor(() => {
            expect(screen.getByText("Invalid JSON in Test Data")).toBeInTheDocument();
        });
    });

    it("displays error for template syntax error", async () => {
        const template = "Hello {% if %}"; // Invalid Jinja/Nunjucks
        const testData = "{}";

        render(
            <SmartTemplateEditor
                value={template}
                onChange={() => {}}
                testData={testData}
            />
        );

        await waitFor(() => {
            expect(screen.getByText((content) => content.includes("Template Error"))).toBeInTheDocument();
        });
    });

    it("renders variables as badges", () => {
        const variables = ["userId", "query"];
        render(
            <SmartTemplateEditor
                value=""
                onChange={() => {}}
                variables={variables}
            />
        );

        expect(screen.getByText("userId")).toBeInTheDocument();
        expect(screen.getByText("query")).toBeInTheDocument();
        expect(screen.getByText("2 variables available")).toBeInTheDocument();
    });
});
