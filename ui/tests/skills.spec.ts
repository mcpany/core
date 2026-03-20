/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from "@playwright/test";
import { seedUser, cleanupUser, seedSkill } from "./e2e/test-data";

test.describe("Skills Management", () => {
  test.beforeEach(async ({ request, page }) => {
    // We only seed a single user and skill for this specific test
    await seedUser(request, "e2e-skills-user");

    // Login first
    await page.goto("/login");
    await page.fill('input[name="username"]', "e2e-skills-user");
    await page.fill('input[name="password"]', "password");
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL("/", { timeout: 15000 });
  });

  test.afterEach(async ({ request }) => {
    await cleanupUser(request, "e2e-skills-user");
  });

  test("should load skills page with empty state", async ({ page }) => {
    await page.goto("/skills");
    await expect(page.getByText("No skills found")).toBeVisible({
      timeout: 10000,
    });
    await expect(
      page.getByRole("button", { name: "Create Skill" }),
    ).toBeVisible();
  });

  test("should delete a skill via custom alert dialog", async ({
    page,
    request,
  }) => {
    const testSkillName = "test-e2e-skill";
    await seedSkill(testSkillName, request);

    await page.goto("/skills");

    // Verify the skill was seeded and is displayed
    await expect(page.getByText(testSkillName).first()).toBeVisible({
      timeout: 10000,
    });

    // Click the delete button on the card
    const trashButton = page
      .locator(".backdrop-blur-sm")
      .filter({ hasText: testSkillName })
      .locator(".text-destructive");
    await trashButton.click();

    // Verify custom alert dialog appears
    await expect(page.getByRole("alertdialog")).toBeVisible();
    await expect(
      page.getByRole("heading", { name: "Delete Skill" }),
    ).toBeVisible();
    await expect(
      page.getByText(
        `Are you sure you want to delete the skill "${testSkillName}"?`,
      ),
    ).toBeVisible();

    // Confirm deletion
    await page.getByRole("button", { name: "Delete" }).click();

    // Verify success toast and skill removal
    await expect(page.getByText("Skill deleted")).toBeVisible({
      timeout: 5000,
    });
    await expect(page.getByText(testSkillName)).not.toBeVisible({
      timeout: 5000,
    });
  });
});
