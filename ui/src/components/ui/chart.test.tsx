/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { render } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { ChartStyle } from "./chart";

describe("ChartStyle Security", () => {
  it("should block CSS injections and unknown formats completely via strict allowlist", () => {
    const dangerousColor = "red; } body { background: red";
    const config = {
      test: {
        color: dangerousColor,
      },
    };

    const { container } = render(
      <ChartStyle id="test-chart" config={config} />,
    );
    const styleTag = container.querySelector("style");

    // It should fail closed and not generate the color test at all
    expect(styleTag?.innerHTML).not.toContain("--color-test");
  });

  it("should block values containing 'url('", () => {
    const dangerousColor = "url(http://evil.com/image.png)";
    const config = {
      test: {
        color: dangerousColor,
      },
    };

    const { container } = render(
      <ChartStyle id="test-chart" config={config} />,
    );
    const styleTag = container.querySelector("style");

    expect(styleTag?.innerHTML).not.toContain("--color-test");
  });

  it("should block values containing 'expression('", () => {
    const dangerousColor = "expression(alert(1))";
    const config = {
      test: {
        color: dangerousColor,
      },
    };

    const { container } = render(
      <ChartStyle id="test-chart" config={config} />,
    );
    const styleTag = container.querySelector("style");
    expect(styleTag?.innerHTML).not.toContain("--color-test");
  });

  it("should block @import CSS injections", () => {
    const dangerousColor =
      "</style><style>@import url('http://evil.com/malicious.css');</style>";
    const config = {
      test: {
        color: dangerousColor,
      },
    };

    const { container } = render(
      <ChartStyle id="test-chart" config={config} />,
    );
    const styleTag = container.querySelector("style");

    expect(styleTag?.innerHTML).not.toContain("--color-test");
  });

  it("should allow safe color values", () => {
    const safeColor = "hsl(var(--primary))";
    const safeHex = "#ef4444";
    const safeRgb = "rgb(239, 68, 68)";
    const safeVar = "var(--danger)";

    const config = {
      test1: { color: safeColor },
      test2: { color: safeHex },
      test3: { color: safeRgb },
      test4: { color: safeVar },
    };

    const { container } = render(
      <ChartStyle id="test-chart" config={config} />,
    );
    const styleTag = container.querySelector("style");

    expect(styleTag?.innerHTML).toContain(
      "--color-test1: hsl(var(--primary));",
    );
    expect(styleTag?.innerHTML).toContain("--color-test2: #ef4444;");
    expect(styleTag?.innerHTML).toContain("--color-test3: rgb(239, 68, 68);");
    expect(styleTag?.innerHTML).toContain("--color-test4: var(--danger);");
  });
});
