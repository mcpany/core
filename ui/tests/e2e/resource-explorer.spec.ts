/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from "@playwright/test";
import { seedGlobalState } from "./test-data";

test.describe("Resource Explorer", () => {
  test.beforeEach(async ({ request }) => {
    // We will use seedGlobalState from test-data.ts, which seeds multiple services,
    // including Echo Service (svc_echo) which we just updated to expose our resources.
    await seedGlobalState(request);
  });

  test("should load resources and allow selection", async ({ page }) => {
    // Navigate to the resources page
    await page.goto("/resources");

    // Wait for the resource list to populate (using actual seeded API)
    await expect(page.getByText("config.json").first()).toBeVisible({
      timeout: 15000,
    });
    await expect(page.getByText("README.md").first()).toBeVisible();

    // Verify search functionality
    const searchInput = page.getByPlaceholder("Search resources...");
    await searchInput.fill("script");
    await expect(page.getByText("script.py").first()).toBeVisible();
    await expect(page.getByText("config.json")).not.toBeVisible();

    // Clear search
    await searchInput.fill("");
    await expect(page.getByText("config.json").first()).toBeVisible();

    // Select a resource
    await page.getByText("config.json").first().click();

    // Verify preview loads
    // Use first() because the list item also shows the URI
    await expect(page.getByText("file:///config.json").first()).toBeVisible(); // URI header

    // Check if content area is visible. For JSON we expect the key 'foo' and value 'bar'
    await expect(page.getByText("foo").first()).toBeVisible();
    await expect(page.getByText("bar").first()).toBeVisible();

    // Verify toolbar buttons
    await expect(page.getByTitle("List View")).toBeVisible();
    await expect(page.getByTitle("Grid View")).toBeVisible();

    // Switch to Grid view
    await page.getByTitle("Grid View").click();

    // Verify grid item exists
    await expect(page.getByText("config.json").first()).toBeVisible();
  });
});
