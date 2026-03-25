/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { FileInput } from "./file-input";
import { vi } from "vitest";

describe("FileInput", () => {
    it("renders correctly", () => {
        const onChange = vi.fn();
        render(<FileInput onChange={onChange} />);
        expect(screen.getByText("Select File")).toBeInTheDocument();
    });

    it("handles drag over and leave", () => {
        const onChange = vi.fn();
        const { container } = render(<FileInput onChange={onChange} />);
        const dropZone = container.firstChild as HTMLElement;

        fireEvent.dragOver(dropZone);
        expect(dropZone).toHaveClass("border-dashed");

        fireEvent.dragLeave(dropZone);
        expect(dropZone).not.toHaveClass("border-dashed");
    });

    it("handles file drop", async () => {
        // Use a wrapper to simulate controlled component behavior
        const Wrapper = () => {
             const [val, setVal] = React.useState<string | undefined>(undefined);
             return <FileInput value={val} onChange={setVal} />;
        };

        const { container } = render(<Wrapper />);
        const dropZone = container.firstChild as HTMLElement;

        const file = new File(["hello"], "hello.txt", { type: "text/plain" });

        fireEvent.drop(dropZone, {
            dataTransfer: {
                files: [file]
            }
        });

        // "hello" in base64 is "aGVsbG8="
        // We wait for the filename to appear, which implies onChange was called
        // and value was updated.
        await waitFor(() => {
            expect(screen.getByText("hello.txt")).toBeInTheDocument();
        });
    });

    it("rejects large files", async () => {
         const onChange = vi.fn();
         const { container } = render(<FileInput onChange={onChange} />);
         const dropZone = container.firstChild as HTMLElement;

         // Mock a large file since creating a real 6MB buffer in test might be slow/heavy
         const largeFile = {
             size: 6 * 1024 * 1024,
             name: "large.txt",
             type: "text/plain"
         } as unknown as File;

         fireEvent.drop(dropZone, {
             dataTransfer: {
                 files: [largeFile]
             }
         });

         await waitFor(() => {
             expect(screen.getByText(/File is too large/i)).toBeInTheDocument();
         });
         expect(onChange).not.toHaveBeenCalled();
    });
});
