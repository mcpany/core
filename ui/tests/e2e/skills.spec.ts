/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Agent Skills', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/skills');
    // Ensure we are on the list page
    await expect(page).toHaveURL(/\/skills\/?$/);
  });

  test.skip('should create and list a new skill', async ({ page }) => {
    // Generate unique name
    const testSkillName = `e2e-test-skill-${Date.now()}`;

    // 1. Navigate to Create Page
    // Use getByRole for better accessibility/reliability check
    try {
        await page.getByRole('link', { name: 'Create Skill' }).click();
    } catch {
        // Fallback for smaller screens or alternative layouts
        await page.goto('/skills/create');
    }
    await expect(page).toHaveURL(/\/skills\/create/);

    // 2. Fill Step 1: Metadata
    await page.fill('input#name', testSkillName);
    await page.fill('textarea#description', 'Created by E2E test');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 3. Fill Step 2: Instructions
    await expect(page.locator('text=Step 2: Instructions')).toBeVisible();
    await page.fill('textarea', '# E2E Instructions\n\nRun this.');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 4. Step 3: Assets
    await expect(page.locator('text=Step 3: Assets')).toBeVisible();

    // Wait for creation API response
    const createPromise = page.waitForResponse(response =>
        response.url().includes('/v1/skills') &&
        (response.status() === 200 || response.status() === 201)
    );
    await page.getByRole('button', { name: 'Create Skill' }).click();
    await createPromise;

    // 5. Verify Redirect to List
    await expect(page).toHaveURL(/\/skills\/?$/);

    // 6. Verify existence with retry
    // In K8s/Distributed systems, read-after-write might be eventually consistent.
    await expect(async () => {
        await page.reload(); // Reload to refresh list
        // Use a more specific locator for the skill card/item
        const skillLocator = page.locator(`text=${testSkillName}`);
        await expect(skillLocator).toBeVisible({ timeout: 5000 });
    }).toPass({
        timeout: 45000, // Increased timeout for K8s
        intervals: [2000, 5000, 10000] // Backoff retry
    });
  });

  test.skip('should view skill details', async ({ page }) => {
    const skillName = 'view-test-skill-' + Date.now();

    // Create skill first (reusing logic but simplified)
    await page.goto('/skills/create');
    await page.fill('input#name', skillName);
    await page.fill('textarea#description', 'Created by E2E test');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    await expect(page.locator('text=Step 2: Instructions')).toBeVisible();
    await page.fill('textarea', '# E2E Instructions\n\nRun this.');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    const createPromise = page.waitForResponse(response =>
        response.url().includes('/v1/skills') &&
        (response.status() === 200 || response.status() === 201)
    );
    await page.getByRole('button', { name: 'Create Skill' }).click();
    await createPromise;
    await expect(page).toHaveURL(/\/skills\/?$/);

    // Wait for list update
    await expect(async () => {
        await page.reload();
        await expect(page.locator(`text=${skillName}`)).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 45000, intervals: [2000, 5000, 10000] });

    // Navigate to detail page directly to verify routing
    await page.goto(`/skills/${skillName}`);
    await expect(page).toHaveURL(new RegExp(`/skills/${skillName}`));

    // Handle eventual consistency for the detail page
    await expect(async () => {
       const notFound = page.locator('text=Skill not found');
       const errorToast = page.locator('text=Failed to load skill');
       const loading = page.locator('text=Loading skill...');

       if (await errorToast.isVisible()) {
           await page.reload();
       } else if (await notFound.isVisible()) {
           await page.reload();
       }

       await expect(page.locator('h1')).toContainText(skillName);
    }).toPass({ timeout: 45000, intervals: [2000, 5000] });
  });
});
