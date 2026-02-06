/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render, screen, fireEvent } from "@testing-library/react";
import { BulkServiceEditor, BulkUpdates } from "./bulk-service-editor";
import { vi } from "vitest";

// Mock EnvVarEditor since it is complex and we just want to test BulkServiceEditor logic
vi.mock("@/components/services/env-var-editor", () => ({
    EnvVarEditor: ({ onChange }: { onChange: (env: any) => void }) => (
        <div data-testid="env-var-editor">
            <button onClick={() => onChange({ "NEW_VAR": { plainText: "new-value", validationRegex: "" } })}>
                Set Env
            </button>
        </div>
    )
}));

describe("BulkServiceEditor", () => {
    it("renders correctly with selected count", () => {
        render(
            <BulkServiceEditor
                selectedCount={5}
                onApply={() => {}}
                onCancel={() => {}}
            />
        );
        expect(screen.getByText(/You are editing/)).toBeInTheDocument();
        expect(screen.getByText("5")).toBeInTheDocument();
    });

    it("allows entering tags", () => {
        const onApply = vi.fn();
        render(
            <BulkServiceEditor
                selectedCount={1}
                onApply={onApply}
                onCancel={() => {}}
            />
        );

        const input = screen.getByPlaceholderText("production, web, internal");
        fireEvent.change(input, { target: { value: "tag1, tag2" } });

        const applyBtn = screen.getByText("Apply Changes");
        fireEvent.click(applyBtn);

        expect(onApply).toHaveBeenCalledWith({
            tags: ["tag1", "tag2"]
        });
    });

    it("calls onCancel", () => {
        const onCancel = vi.fn();
        render(
            <BulkServiceEditor
                selectedCount={1}
                onApply={() => {}}
                onCancel={onCancel}
            />
        );

        const cancelBtn = screen.getByText("Cancel");
        fireEvent.click(cancelBtn);

        expect(onCancel).toHaveBeenCalled();
    });
});
