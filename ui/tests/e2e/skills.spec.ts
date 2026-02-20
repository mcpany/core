/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, seedUser, seedProfiles, cleanupServices, cleanupUser, cleanupProfiles } from './test-data';

test.describe('Agent Skills', () => {
  test.beforeEach(async ({ page, request }) => {
    await seedServices(request);
    await seedProfiles(request);
    await seedUser(request, "e2e-admin-skills");

    // Login first
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="username"]', 'e2e-admin-skills');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]', { force: true });
    await page.waitForURL('/', { timeout: 60000 });
    await expect(page).toHaveURL('/', { timeout: 15000 });

    await page.goto('/skills');
    // Ensure we are on the list page
    await expect(page).toHaveURL(/\/skills\/?$/);
  });

  test.afterEach(async ({ request }) => {
    await cleanupServices(request);
    await cleanupProfiles(request);
    // await cleanupUser(request, "e2e-admin-skills");
  });

  test('should create and list a new skill', async ({ page }) => {
    const testSkillName = `e2e-test-skill-${Date.now()}`;

    // 1. Fill Metadata
    await page.getByRole('button', { name: 'Create Skill' }).first().click();
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

    // Click the Save button in the wizard specifically
    const saveButton = page.locator('main').locator('button:has-text("Create Skill")');
    await expect(saveButton).toBeVisible();
    await saveButton.click({ force: true });
    await createPromise;

    // 5. Verify Redirect to List
    await expect(page).toHaveURL(/\/skills\/?$/);

    // 6. Verify existence with retry
    // In K8s/Distributed systems, read-after-write might be eventually consistent.
    await expect(async () => {
        await page.reload();
        await expect(page.locator(`text=${testSkillName}`)).toBeVisible({ timeout: 5000 });
    }).toPass({
        timeout: 60000, // Increased timeout for K8s
        intervals: [2000, 5000, 10000] // Backoff retry
    });
  });

  test('should view skill details', async ({ page }) => {
    const skillName = `view-test-skill-${Date.now()}`;

    // Create a skill first (minimal metadata)
    await page.getByRole('button', { name: 'Create Skill' }).first().click();
    await page.fill('input#name', skillName);
    await page.fill('textarea#description', 'Created by View Test');
    await page.getByRole('button', { name: 'Next', exact: true }).click();
    await page.fill('textarea', '# Instructions');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    const createPromise = page.waitForResponse(response =>
        response.url().includes('/api/v1/skills') &&
        response.request().method() === 'POST' &&
        (response.status() === 200 || response.status() === 201),
        { timeout: 30000 }
    );
    // Click the Save button in the wizard specifically
    const saveButton = page.locator('main').locator('button:has-text("Create Skill")');
    await expect(saveButton).toBeVisible();
    await saveButton.click();
    await createPromise;
    await expect(page).toHaveURL(/\/skills\/?$/);

    // Wait for list to sync
    await expect(async () => {
        await page.reload();
        await expect(page.locator(`text=${skillName}`)).toBeVisible({ timeout: 5000 });
    }).toPass({ timeout: 60000, intervals: [2000, 5000, 10000] });

    // Navigate to detail page directly to verify routing
    await expect(async () => {
        await page.goto(`/skills/${skillName}`);
      await expect(page.getByText(skillName).first()).toBeVisible();
    }).toPass({ timeout: 60000, intervals: [2000, 5000, 10000] });
  });
});
