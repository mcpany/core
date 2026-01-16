/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Agent Skills', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/skills');
  });

  test('should create and list a new skill', async ({ page }) => {


    // 1. Navigate to Create Page as normal   // Generate unique name to avoid collision on persistent backend
    const testSkillName = `e2e-test-skill-${Date.now()}`;
    await page.click('text=Create Skill');
    await expect(page).toHaveURL(/\/skills\/create/);

    // 2. Fill Step 1: Metadata
    await page.fill('input#name', testSkillName);
    await page.fill('textarea#description', 'Created by E2E test');
    await page.click('button:has-text("Next")');

    // 3. Fill Step 2: Instructions
    await expect(page.locator('text=Step 2: Instructions')).toBeVisible();
    await page.fill('textarea', '# E2E Instructions\n\nRun this.');
    await page.click('button:has-text("Next")');

    // 4. Step 3: Assets (Skip upload for now)
    await expect(page.locator('text=Step 3: Assets')).toBeVisible();
    await page.click('button:has-text("Create Skill")');

    // 5. Verify Redirect to List
    await expect(page).toHaveURL(/\/skills/);

    // 6. Verify existence with retry/reload, as backend might be slightly eventual or UI SWR cache needs update
    await expect(async () => {
        // Reload page to bypass client cache if needed
        await page.reload();
        await expect(page.locator(`text=${testSkillName}`)).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 20000 });

    await expect(page.locator('text=Created by E2E test')).toBeVisible();
  });

  test('should view skill details', async ({ page }) => {
    // Ensure unique name to avoid collision with previous test
    const skillName = 'view-test-skill-' + Date.now();

    await page.click('text=Create Skill');
    await page.fill('input#name', skillName);
    await page.click('button:has-text("Next")');
    await page.click('button:has-text("Next")');
    await page.click('button:has-text("Create Skill")');

    // Wait for list with retry
    await expect(async () => {
        await page.reload();
        await expect(page.locator(`text=${skillName}`)).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 20000 });

    // Click View - finding the specific card link
    // The card title usually contains the name, and there is a "View Details" link
    // We can target the "View Details" link inside the card that has the skill name
    const card = page.locator('.card', { hasText: skillName });
    await card.locator('text=View Details').click();

    await expect(page).toHaveURL(new RegExp(`/skills/${skillName}`));
    await expect(page.locator('h1')).toContainText(skillName);
  });
});
