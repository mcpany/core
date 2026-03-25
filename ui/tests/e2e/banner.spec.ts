import { seedGlobalState } from "./test-data";
/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from "@playwright/test";

test.describe("System Status Banner", () => {
  test.beforeEach(async ({ page, request }) => {
    await seedGlobalState(request);

    // Reset any previous mocks
    await page.unrouteAll({ behavior: "ignoreErrors" });
  });

  test("should not be visible when system is healthy", async ({ page }) => {
    // Mock healthy doctor response

    await page.goto("/");
    await expect(
      page.locator('div[role="alert"]').filter({ hasText: "System Status" }),
    ).not.toBeVisible();
    await expect(page.getByText("Connection Error")).not.toBeVisible();
    await expect(page.getByText("Configuration Error")).not.toBeVisible();
  });

  test("should show connection error when backend is unreachable", async ({
    page,
  }) => {
    // Mock network error

    await page.goto("/");
    await expect(page.getByText("Connection Error")).toBeVisible();
    await expect(
      page.getByText("Could not connect to the server health check"),
    ).toBeVisible();
  });

  test("should show configuration error when config check fails", async ({
    page,
  }) => {
    await page.goto("/");
    await expect(page.getByText("Configuration Error")).toBeVisible();
    await expect(
      page.getByText("Invalid YAML syntax in config.yaml"),
    ).toBeVisible();
  });

  test("should show configuration error with diff", async ({ page }) => {
    await page.goto("/");
    await expect(page.getByText("Configuration Error")).toBeVisible();
    await expect(page.getByText("Invalid YAML syntax")).toBeVisible();
    await expect(page.getByText("Configuration Diff:")).toBeVisible();
    // Use .locator with hasText to find specific lines if getByText is too strict about newlines
    await expect(
      page.locator("pre").filter({ hasText: "-valid" }),
    ).toBeVisible();
    await expect(
      page.locator("pre").filter({ hasText: "+invalid" }),
    ).toBeVisible();
  });

  test("should show degraded status for other check failures", async ({
    page,
  }) => {
    await page.goto("/");
    const banner = page
      .locator('div[role="alert"]')
      .filter({ hasText: "System Status: Degraded" });
    await expect(banner).toBeVisible();
    await expect(
      banner.getByText("Database: Connection timeout"),
    ).toBeVisible();
    await expect(banner.getByText("Cache: Redis unavailable")).toBeVisible();
  });
});
