import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { FileInput } from "./file-input";
import React from "react";
import '@testing-library/jest-dom';
import { vi, describe, it, expect, beforeEach } from 'vitest';

// Mock URL.createObjectURL
global.URL.createObjectURL = vi.fn(() => "blob:test-url");
global.URL.revokeObjectURL = vi.fn();

// Mock FileReader
class MockFileReader {
  onload: ((e: ProgressEvent<FileReader>) => void) | null = null;
  onerror: (() => void) | null = null;
  result: string | null = null;

  readAsDataURL(file: File) {
    this.result = "data:image/png;base64,TESTBASE64";
    // Simulate async
    setTimeout(() => {
        if (this.onload) {
            this.onload({ target: { result: this.result } } as unknown as ProgressEvent<FileReader>);
        }
    }, 10);
  }
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
(global as any).FileReader = MockFileReader;

describe("FileInput", () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    it("renders correctly", () => {
        render(<FileInput onChange={vi.fn()} />);
        expect(screen.getByText("Click or drag file to upload")).toBeInTheDocument();
    });

    it("handles drag over and leave events", () => {
        render(<FileInput onChange={vi.fn()} />);
        const dropZoneText = screen.getByText("Click or drag file to upload");
        // Access grandparent or parent to find the div with classes
        const dropZone = dropZoneText.closest(".border-dashed");

        if (!dropZone) throw new Error("Drop zone not found");

        fireEvent.dragOver(dropZone);
        expect(dropZone).toHaveClass("border-primary");

        fireEvent.dragLeave(dropZone);
        // It might not immediately remove class if react update is pending, but usually sync in tests
        expect(dropZone).not.toHaveClass("border-primary");
    });

    it("handles file drop (image)", async () => {
        const handleChange = vi.fn();
        render(<FileInput onChange={handleChange} />);
        const dropZoneText = screen.getByText("Click or drag file to upload");
        const dropZone = dropZoneText.closest(".border-dashed");

        if (!dropZone) throw new Error("Drop zone not found");

        const file = new File(["dummy content"], "test.png", { type: "image/png" });

        // Create dataTransfer
        const dataTransfer = {
            files: [file],
            items: [{ kind: 'file', type: file.type, getAsFile: () => file }],
            types: ['Files']
        };

        fireEvent.drop(dropZone, { dataTransfer });

        await waitFor(() => {
            expect(handleChange).toHaveBeenCalledWith("TESTBASE64");
        });

        expect(screen.getByAltText("Preview")).toBeInTheDocument();
        expect(screen.getByText("test.png")).toBeInTheDocument();
    });

    it("handles file drop (non-image)", async () => {
         const handleChange = vi.fn();
        render(<FileInput onChange={handleChange} />);
        const dropZoneText = screen.getByText("Click or drag file to upload");
        const dropZone = dropZoneText.closest(".border-dashed");

        if (!dropZone) throw new Error("Drop zone not found");

        const file = new File(["dummy content"], "test.txt", { type: "text/plain" });

        const dataTransfer = {
            files: [file],
        };

        fireEvent.drop(dropZone, { dataTransfer });

        await waitFor(() => {
            expect(handleChange).toHaveBeenCalledWith("TESTBASE64");
        });

        expect(screen.queryByAltText("Preview")).not.toBeInTheDocument();
        expect(screen.getAllByText("test.txt")[0]).toBeInTheDocument();
    });

    it("triggers file select on keydown (Enter)", () => {
        render(<FileInput onChange={vi.fn()} />);
        const dropZoneText = screen.getByText("Click or drag file to upload");
        const dropZone = dropZoneText.closest(".border-dashed");

        if (!dropZone) throw new Error("Drop zone not found");

        const clickSpy = vi.spyOn(HTMLInputElement.prototype, 'click');

        fireEvent.keyDown(dropZone, { key: 'Enter', code: 'Enter' });

        expect(clickSpy).toHaveBeenCalled();
        clickSpy.mockRestore();
    });
});
