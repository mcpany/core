/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, cleanupUser } from './test-data';

test.describe('Credentials Management', () => {

  test.beforeEach(async ({ page, request }) => {
    await seedUser(request, "e2e-cred-admin");
    // Login
    await page.goto('/login');
    await page.fill('input[name="username"]', 'e2e-cred-admin');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/');
  });

  test.afterEach(async ({ request }) => {
    await cleanupUser(request, "e2e-cred-admin");
  });

  test('should list, create, update and delete credentials', async ({ page }) => {
    await page.goto('/credentials');

    // Ensure empty state or at least new credential is not there
    await expect(page.getByText('Test API Key')).toBeHidden();

    // 2. Create Credential
    await page.getByRole('button', { name: 'New Credential' }).click();
    await expect(page.getByText('Create Credential')).toBeVisible();

    await page.getByPlaceholder('My Credential').fill('Test API Key');
    await page.getByPlaceholder('X-API-Key').fill('Authorization');
    await page.getByPlaceholder('...secret key...').fill('secret-key');

    await page.getByRole('button', { name: 'Save' }).click();

    // Verify it appears in list
    await expect(page.getByText('Test API Key')).toBeVisible();
    await expect(page.locator('tbody').getByText('API Key', { exact: true })).toBeVisible();

    // 3. Update Credential
    // We target the row with "Test API Key" to ensure we click the right Edit button
    const row = page.locator('tr').filter({ hasText: 'Test API Key' });
    await row.getByRole('button', { name: 'Edit' }).click();

    await page.getByPlaceholder('My Credential').fill('Updated API Key');
    await page.getByRole('button', { name: 'Save' }).click();

    await expect(page.getByText('Updated API Key')).toBeVisible();

    // 4. Delete Credential
    const updatedRow = page.locator('tr').filter({ hasText: 'Updated API Key' });

    // Accept delete confirmation (native)
    page.on('dialog', dialog => dialog.accept());

    await updatedRow.getByRole('button', { name: 'Delete' }).click();

    // Handle Radix alert dialog if present (overrides native dialog often)
    // Wait for dialog to appear to avoid race condition
    try {
        await expect(page.getByText('Are you sure?')).toBeVisible({ timeout: 5000 });
        await page.getByRole('button', { name: 'Delete' }).last().click();
    } catch (e) {
        // Ignore if native dialog handled it
    }

    await expect(page.getByText('Updated API Key')).toBeHidden();
  });
});
