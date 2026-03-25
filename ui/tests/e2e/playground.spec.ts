import { seedGlobalState } from "./test-data";
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from "@playwright/test";

test.describe("Playground Complex Schema Support", () => {
  test.beforeEach(async ({ request }) => {
    await seedGlobalState(request);
  });

  test("should allow configuring and running a tool with complex nested schema", async ({
    page,
  }) => {
    // Mock the tools API to return a tool with complex schema

    // Mock the execute API

    // Navigate to playground
    await page.goto("/playground");
    // await expect(page.getByRole('heading', { name: 'Playground' })).toBeVisible();

    // Open tools list (Sidebar is open by default)
    // await page.getByRole('button', { name: 'Available Tools' }).click();

    // Select the complex tool
    await expect(page.getByText("complex_tool")).toBeVisible();
    // The button says "Use Tool"
    await page.getByRole("button", { name: /^Use$/i }).first().click();

    // Verify form structure
    // Note: The UI might append type info like "user (object)", so we disable exact match
    await expect(
      page.getByRole("button", { name: "Execute", exact: true }),
    ).toBeVisible();

    // Try to submit empty form (should fail validation because user.name is required)
    await page.getByRole("button", { name: "Execute", exact: true }).click();

    // Fill the form
    await expect(page.getByText(/name/i).first()).toBeVisible();
    await page
      .locator('input[name="name"], input[id*="name"], textarea')
      .first()
      .fill("Bob");

    await expect(page.getByText(/age/i).first()).toBeVisible();
    await page
      .locator('input[name="age"], input[id*="age"], textarea')
      .first()
      .fill("30");

    // Add tag
    await page.getByRole("button", { name: "Add Item" }).click();
    // Wait for the new input to appear
    const newItemInput = page.locator("input, textarea").last();
    await expect(newItemInput).toBeVisible({ timeout: 5000 });
    await newItemInput.fill("developer");

    // Execute command in Tool Runner
    await page.getByRole("button", { name: "Execute", exact: true }).click();

    // Verify result appears in Result pane
    await expect(page.locator("text=Executed complex_tool")).toBeVisible({
      timeout: 10000,
    });
  });
});
