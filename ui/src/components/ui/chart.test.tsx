/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { ChartStyle } from "./chart";

describe("ChartStyle Security", () => {
  it("should sanitize dangerous characters from color values", () => {
    // Semicolons and braces are stripped, preventing breaking out of the style attribute
    const dangerousColor = "red; } body { background: red";
    const config = {
      test: {
        color: dangerousColor,
      },
    };

    const { container } = render(<ChartStyle id="test-chart" config={config} />);
    const styleTag = container.querySelector("style");

    // The sanitizer removes ; } { :
    // "red; } body { background: red" -> "red  body  background red"

    // We expect the property definition to be intact and sanitized.
    // The value should be stripped of dangerous characters.
    expect(styleTag?.innerHTML).toContain("--color-test: red  body  background red;");
  });

  it("should block values containing 'url('", () => {
    // Even if sanitized, url() should be blocked to prevent external requests
    const dangerousColor = "url(http://evil.com/image.png)";
    const config = {
      test: {
        color: dangerousColor,
      },
    };

    const { container } = render(<ChartStyle id="test-chart" config={config} />);
    const styleTag = container.querySelector("style");

    // Should NOT contain the dangerous value, it returns null for that entry
    // So --color-test should not be generated
    expect(styleTag?.innerHTML).not.toContain("--color-test");
  });

  it("should block values containing 'expression('", () => {
    const dangerousColor = "expression(alert(1))";
    const config = {
      test: {
        color: dangerousColor,
      },
    };

    const { container } = render(<ChartStyle id="test-chart" config={config} />);
    const styleTag = container.querySelector("style");
    expect(styleTag?.innerHTML).not.toContain("--color-test");
  });

  it("should allow safe color values", () => {
    const safeColor = "hsl(var(--primary))";
    const config = {
      test: {
        color: safeColor,
      },
    };

    const { container } = render(<ChartStyle id="test-chart" config={config} />);
    const styleTag = container.querySelector("style");
    expect(styleTag?.innerHTML).toContain("--color-test: hsl(var(--primary));");
  });

  it("should block values containing 'javascript:'", () => {
    const dangerousColor = "javascript:alert(1)";
    const config = {
      test: {
        color: dangerousColor,
      },
    };

    const { container } = render(<ChartStyle id="test-chart" config={config} />);
    const styleTag = container.querySelector("style");
    expect(styleTag?.innerHTML).not.toContain("--color-test");
  });

  it("should block values containing 'data:'", () => {
    const dangerousColor = "data:text/html;base64,...";
    const config = {
      test: {
        color: dangerousColor,
      },
    };

    const { container } = render(<ChartStyle id="test-chart" config={config} />);
    const styleTag = container.querySelector("style");
    expect(styleTag?.innerHTML).not.toContain("--color-test");
  });
});
