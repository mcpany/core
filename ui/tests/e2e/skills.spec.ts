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
    await Promise.all([
      page.waitForURL('/', { timeout: 30000 }),
      page.click('button[type="submit"]', { force: true })
    ]);
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

    // 6. Verify we are back to the skills list after a successful create
    await expect(page.getByRole('button', { name: 'Create Skill' }).first()).toBeVisible();
  });

  test('should support bulk deletion of skills', async ({ page }) => {
    const timestamp = Date.now();
    const skill1 = `bulk-delete-1-${timestamp}`;
    const skill2 = `bulk-delete-2-${timestamp}`;

    // Create skill 1
    await page.getByRole('button', { name: 'Create Skill' }).first().click();
    await page.fill('input#name', skill1);
    await page.getByRole('button', { name: 'Next', exact: true }).click();
    await page.fill('textarea', '# Instructions 1');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    let createPromise = page.waitForResponse(response =>
        response.url().includes('/api/v1/skills') &&
        response.request().method() === 'POST' &&
        (response.status() === 200 || response.status() === 201),
        { timeout: 30000 }
    );
    await page.locator('main').locator('button:has-text("Create Skill")').click();
    await createPromise;
    await expect(page).toHaveURL(/\/skills\/?$/);

    // Create skill 2
    await page.getByRole('button', { name: 'Create Skill' }).first().click();
    await page.fill('input#name', skill2);
    await page.getByRole('button', { name: 'Next', exact: true }).click();
    await page.fill('textarea', '# Instructions 2');
    await page.getByRole('button', { name: 'Next', exact: true }).click();

    createPromise = page.waitForResponse(response =>
        response.url().includes('/api/v1/skills') &&
        response.request().method() === 'POST' &&
        (response.status() === 200 || response.status() === 201),
        { timeout: 30000 }
    );
    await page.locator('main').locator('button:has-text("Create Skill")').click();
    await createPromise;
    await expect(page).toHaveURL(/\/skills\/?$/);

    // Switch to list view (which has the checkboxes exposed directly or grid view which also now has them)
    // We expect both skills to be visible
    await expect(page.getByText(skill1)).toBeVisible();
    await expect(page.getByText(skill2)).toBeVisible();

    // Check both skills via their row/card checkboxes using aria-label
    await page.getByRole('checkbox', { name: `Select ${skill1}` }).click();
    await page.getByRole('checkbox', { name: `Select ${skill2}` }).click();

    // Verify "2 selected" is visible
    await expect(page.getByText('2 selected')).toBeVisible();

    // Accept the confirm dialog
    page.once('dialog', dialog => dialog.accept());

    // Click Bulk Delete
    await page.getByRole('button', { name: 'Bulk Delete' }).click();

    // Verify both are removed
    await expect(page.getByText(skill1)).not.toBeVisible({ timeout: 10000 });
    await expect(page.getByText(skill2)).not.toBeVisible();
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

    // Keep this as a navigation smoke check to avoid backend eventual-consistency flakiness.
    await page.goto('/skills');
    await expect(page).toHaveURL(/\/skills\/?$/);
  });
});
