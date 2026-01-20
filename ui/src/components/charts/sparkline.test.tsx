/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */


import { render } from "@testing-library/react";
import { Sparkline } from "./sparkline";
import React from "react";

describe("Sparkline", () => {
    it("renders correctly with data", () => {
        const { container } = render(<Sparkline data={[10, 20, 30]} width={100} height={50} />);
        const svg = container.querySelector("svg");
        expect(svg).toBeInTheDocument();
        // Check if paths are created
        const paths = container.querySelectorAll("path");
        expect(paths.length).toBe(2); // One for fill, one for stroke
    });

    it("renders fallback when no data", () => {
        const { container } = render(<Sparkline data={[]} />);
        const div = container.querySelector("div");
        expect(div).toBeInTheDocument();
        expect(div).toHaveClass("bg-muted/20");
    });

    it("handles single data point", () => {
        const { container } = render(<Sparkline data={[10]} />);
        const svg = container.querySelector("svg");
        expect(svg).toBeInTheDocument();
    });
});
