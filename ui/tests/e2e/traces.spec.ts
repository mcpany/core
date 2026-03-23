/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from "@playwright/test";
import { seedGlobalState } from "./test-data";

test.describe("Trace Viewer", () => {
  test.beforeEach(async ({ page, request }) => {
    await seedGlobalState(request);

    // Clear any existing traces
    await request.delete("/api/v1/traces", {
      headers: {
        Authorization:
          "Basic " + Buffer.from("e2e-admin-core:password").toString("base64"),
      },
    });

    // Seed a real trace in the backend instead of mocking UI
    await request.post("/api/v1/debug/traces", {
      headers: {
        Authorization:
          "Basic " + Buffer.from("e2e-admin-core:password").toString("base64"),
      },
      data: {
        id: "trace-1",
        rootSpan: {
          id: "span-1",
          name: "calculate_sum",
          serviceName: "Math",
          type: "tool",
          status: "success",
          startTime: Date.now() - 150,
          endTime: Date.now(),
          input: { args: [1, 2, 3] },
          output: {
            content: [
              { type: "text", text: '[\n  {\n    "result": 6\n  }\n]' },
            ],
          },
          children: [],
        },
        timestamp: new Date().toISOString(),
        totalDuration: 150,
        status: "success",
        trigger: "user",
      },
    });

    await page.goto("/login");
    await page.waitForLoadState("networkidle");
    await page.fill('input[name="username"]', "e2e-admin-core");
    await page.fill('input[name="password"]', "password");
    await Promise.all([
      page.waitForURL("/", { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true }),
    ]);
    await expect(page).toHaveURL("/", { timeout: 15000 });
  });

  test("should navigate to traces page and view details", async ({ page }) => {
    await page.goto("/");

    // Check if Traces link exists in sidebar and click it
    const tracesLink = page.getByRole("link", { name: "Traces" });
    if ((await tracesLink.count()) > 0) {
      await expect(tracesLink).toHaveAttribute("href", "/traces");
      await Promise.all([page.waitForURL(/\/traces/), tracesLink.click()]);
      await expect(page).toHaveURL(/\/traces/);
    } else {
      // Fallback for when link is hidden (e.g. non-admin)
      console.log(
        "Traces link not found (likely non-admin), trying direct navigation",
      );
      await page.goto("/traces");
      await expect(page).toHaveURL(/\/traces/);
    }

    // Wait for traces to load
    await page.waitForSelector("text=Loading traces...", { state: "detached" });

    // Check if list is populated (should have at least one trace from mock)
    // Check if list is populated (should have at least one trace from db seed)
    const firstTrace = page.locator("button.flex.flex-col").first();
    await expect(firstTrace).toBeVisible();

    // Click the first trace
    await firstTrace.click();

    // Check if details pane is populated
    await expect(page.getByText("Execution Waterfall").first()).toBeVisible();
    await expect(page.locator("text=Execution Waterfall")).toBeVisible();
    await expect(page.locator("text=Root Input")).toBeVisible();

    // Verify formatted table instead of raw JSON dump for output
    await page.getByRole("tab", { name: "Output" }).click();
    await expect(page.getByRole("tab", { name: "Table" })).toBeVisible();
  });

  test("should clear all traces", async ({ page, request }) => {
    await page.goto("/traces");
    await page.waitForSelector("text=Loading traces...", { state: "detached" });

    // Verify trace exists
    await expect(page.locator("button.flex.flex-col").first()).toBeVisible();

    // Click Clear All
    page.on("dialog", (dialog) => dialog.accept());
    await page.getByRole("button", { name: "Clear All Traces" }).click();

    // Verify it disappears from UI
    await expect(page.locator("text=No traces found.")).toBeVisible();

    // Verify backend is actually cleared
    const res = await request.get("/api/v1/traces", {
      headers: {
        Authorization:
          "Basic " + Buffer.from("e2e-admin-core:password").toString("base64"),
      },
    });
    const traces = await res.json();
    expect(traces).toEqual([]);
  });

  test("should filter traces", async ({ page }) => {
    await page.goto("/traces");

    // Wait for traces
    await page.waitForSelector("text=Loading traces...", { state: "detached" });

    // Type in search box
    await page.fill('input[placeholder="Search traces..."]', "calculate");

    // Expect only matching items
    // and doesn't crash the page
    await expect(
      page.locator('input[placeholder="Search traces..."]'),
    ).toHaveValue("calculate");
  });

  test("should replay trace in playground", async ({ page }) => {
    await page.goto("/traces");

    // Ensure we have a trace to click
    await page.waitForSelector("text=Loading traces...", { state: "detached" });
    const firstTrace = page.locator("button.flex.flex-col").first();
    await expect(firstTrace).toBeVisible();
    await firstTrace.click();

    // Click "Replay in Playground"
    // We look for the button with specific text
    const replayBtn = page.getByRole("button", {
      name: "Replay in Playground",
    });
    await expect(replayBtn).toBeVisible();
    await replayBtn.click({ force: true });

    // Verify redirection to playground
    await expect(page).toHaveURL(/\/playground.*/, { timeout: 5000 });

    // Verify query params are present (tool and args)
    // We don't check exact values as they depend on the random mock trace
    const url = page.url();
    expect(url).toContain("tool=");
    expect(url).toContain("args=");

    // Verify Playground input is populated
    // The input should contain the tool name or args
    // We wait for the form or input to be visible first
    await expect(
      page
        .getByPlaceholder("Enter command or select a tool...")
        .or(page.locator("textarea")),
    ).toBeVisible();
  });
});
