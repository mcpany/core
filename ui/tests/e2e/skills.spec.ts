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
    // 1. Navigate to Create Page
    await page.click('text=Create Skill');
    await expect(page).toHaveURL(/\/skills\/create/);

    // 2. Fill Step 1: Metadata
    await page.fill('input#name', 'e2e-test-skill');
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
    await expect(page.locator('text=e2e-test-skill')).toBeVisible();
    await expect(page.locator('text=Created by E2E test')).toBeVisible();
  });

  test('should view skill details', async ({ page }) => {
    // Assuming 'e2e-test-skill' exists from previous test or we create one.
    // Tests might run in parallel or randomized order, so ideally create fresh or mock.
    // For E2E against real backend, we should be careful.
    // Let's create one first to be safe, or assume clean state.

    // Quick create for isolation if needed, but let's try to just check if we can click one.
    // If list is empty, this fails.
    // Let's create one first.
    await page.click('text=Create Skill');
    await page.fill('input#name', 'view-test-skill');
    await page.click('button:has-text("Next")');
    await page.click('button:has-text("Next")');
    await page.click('button:has-text("Create Skill")');

    // Click View
    await page.click('text=view-test-skill');
    // or find the card and click View/Edit?
    // The list likely links the name or has a button.
    // In SkillList.tsx: <Link href={`/skills/${skill.name}`}> ... <Button>View</Button> ... </Link>

    // Let's just click the name if it's a link, or the View button.
    // Finding the card for 'view-test-skill'
    const card = page.locator('.card', { hasText: 'view-test-skill' }).first();
    // Verify it exists (might need reload if list doesn't auto-refresh? It should)
    // Next.js client navigation should reflect the update if router.push was used.

    // Click "View Details" (assuming text)
    await page.click('text=view-test-skill'); // This might just be text in card.
    // Let's look for the link.
    // If the whole card is not clickable, we look for "View Details" button.
    // Assume "View Details" button exists.

    await expect(page).toHaveURL(/\/skills\/view-test-skill/);
    await expect(page.locator('h1')).toHaveText('view-test-skill');
  });
});
