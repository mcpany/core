/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from "@playwright/test";
import { seedGlobalState, seedUser, cleanupUser } from "./test-data";

test.describe("Prompts Workbench", () => {
  test.beforeEach(async ({ page, request }) => {
    await seedGlobalState(request);
    await seedUser(request, "e2e-admin-prompts");

    // Login
    await page.goto("/login");
    await page.fill('input[name="username"]', "e2e-admin-prompts");
    await page.fill('input[name="password"]', "password");
    await Promise.all([
      page.waitForURL("/", { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true }),
    ]);
  });

  test.afterEach(async ({ request }) => {
    await cleanupUser(request, "e2e-admin-prompts");
  });

  test("should load prompts list and allow selection", async ({ page }) => {
    // Navigate to prompts page
    await page.goto("/prompts");

    // Check if the page title exists
    await expect(
      page.locator("h3", { hasText: "Prompt Library" }),
    ).toBeVisible();

    // Check for search input to ensure basic layout
    await expect(
      page.locator('input[placeholder="Search prompts..."]'),
    ).toBeVisible();

    // Handle potential empty state or populated list
    const noPrompts = page.getByText("No prompts found");
    const firstPrompt = page
      .locator("div.flex.flex-col.p-2.gap-1 > button")
      .first();

    // Wait for either no prompts functionality or the list to populate
    await Promise.race([
      expect(noPrompts).toBeVisible(),
      expect(firstPrompt).toBeVisible(),
    ]);

    if (await firstPrompt.isVisible()) {
      await firstPrompt.click();
      // Check for details view
      await expect(page.getByTestId("prompt-details")).toContainText(
        "Configuration",
      );

      // Let's test the execution and RichResultViewer
      const aInput = page.locator("#a");
      const bInput = page.locator("#b");
      await expect(aInput).toBeVisible();
      await expect(bInput).toBeVisible();
      await aInput.fill("5");
      await bInput.fill("3");

      const generateBtn = page.getByRole("button", { name: "Generate" });
      await generateBtn.click();

      // Wait for the result to appear
      // With RichResultViewer, we verify the JSON or rendered view tabs exist
      await expect(page.getByRole("tab", { name: "JSON" })).toBeVisible({
        timeout: 10000,
      });

      // Check if text constructed using the template appears
      await expect(page.getByText("What is 5 + 3?")).toBeVisible();
    } else {
      await expect(noPrompts).toBeVisible();
    }
  });
});
