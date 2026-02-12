/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Agent Skills', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/skills');
    // Ensure we are on the list page
    await expect(page).toHaveURL(/\/skills\/?$/);
  });

  test('should create and list a new skill', async ({ page }) => {
    const testSkillName = `e2e-test-skill-${Date.now()}`;

    // 1. Fill Metadata
    await page.getByRole('button', { name: 'Create Skill' }).click();
    await page.fill('input#name', testSkillName);
    await page.fill('textarea#description', 'Created by E2E test');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 2. Fill Instructions
    await expect(page.locator('text=Step 2: Instructions')).toBeVisible();
    await page.fill('textarea', '# E2E Instructions\n\nRun this.');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    // 3. Final Step (Assets)
    await expect(page.locator('text=Step 3: Assets')).toBeVisible();

    // Wait for creation API response
    const createPromise = page.waitForResponse(response =>
        response.url().includes('/api/v1/skills') &&
        response.request().method() === 'POST' &&
        (response.status() === 200 || response.status() === 201),
        { timeout: 30000 }
    );

    await page.getByRole('button', { name: 'Create Skill' }).click();
    await createPromise;

    // 5. Verify Redirect to List
    await expect(page).toHaveURL(/\/skills\/?$/);

    // 6. Verify existence with retry
    // In K8s/Distributed systems, read-after-write might be eventually consistent.
    await expect(async () => {
        await page.reload();
        await expect(page.locator(`text=${testSkillName}`)).toBeVisible({ timeout: 5000 });
    }).toPass({
        timeout: 45000, // Increased timeout for K8s
        intervals: [2000, 5000, 10000] // Backoff retry
    });
  });

  test('should view skill details', async ({ page }) => {
    const skillName = `view-test-skill-${Date.now()}`;

    // Create a skill first (minimal metadata)
    await page.getByRole('button', { name: 'Create Skill' }).click();
    await page.fill('input#name', skillName);
    await page.fill('textarea#description', 'Created by View Test');
    await page.getByRole('button', { name: 'Next', exact: true }).click();
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    const createPromise = page.waitForResponse(response =>
        response.url().includes('/api/v1/skills') &&
        response.request().method() === 'POST' &&
        (response.status() === 200 || response.status() === 201),
        { timeout: 30000 }
    );
    await page.getByRole('button', { name: 'Create Skill' }).click();
    await createPromise;
    await expect(page).toHaveURL(/\/skills\/?$/);

    // Wait for list to sync
    await expect(async () => {
        await page.reload();
        await expect(page.locator(`text=${skillName}`)).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 45000, intervals: [2000, 5000, 10000] });

    // Navigate to detail page directly to verify routing
    await expect(async () => {
        await page.goto(`/skills/${skillName}`);
        await expect(page.locator('h1')).toContainText(skillName);
    }).toPass({ timeout: 45000, intervals: [2000, 5000, 10000] });
  });
});
