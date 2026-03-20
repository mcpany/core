/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Credentials Management', () => {

  test.beforeEach(async ({ request }) => {
      // Clear out credentials using seed empty state to start clean
      await request.post('/api/v1/debug/seed', {
          data: {
              upstream_services: [],
              credentials: [],
              secrets: [],
              profiles: [],
              users: []
          },
          headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' }
      });
  });

  test('should list, create, update and delete credentials', async ({ page }) => {

    await page.goto('/credentials');

    // Wait for the table to load.
    // If it's empty, wait for the table body to have 0 rows or for the empty state message.

    // Sometimes it takes a moment for the data to arrive. Wait for it to become visible.
    await expect(page.getByText('No credentials found')).toBeVisible({ timeout: 15000 });

    // 2. Create Credential
    await page.getByRole('button', { name: 'New Credential' }).click();
    await expect(page.getByText('Create Credential')).toBeVisible();

    await page.getByPlaceholder('My Credential').fill('Test API Key');
    // Default format is API Key, so just fill details
    await page.getByPlaceholder('X-API-Key').fill('Authorization');
    await page.getByPlaceholder('...secret key...').fill('secret-key');

    await page.waitForTimeout(500);
    await page.getByRole('button', { name: 'Save' }).click({ force: true });

    // Verify it appears in list
    await expect(page.getByText('Test API Key')).toBeVisible({ timeout: 10000 });
    await expect(page.locator('tbody').getByText('API Key', { exact: true })).toBeVisible();

    // 3. Update Credential
    await page.getByRole('button', { name: 'Edit' }).click();
    await page.getByPlaceholder('My Credential').fill('Updated API Key');
    await page.getByRole('button', { name: 'Save' }).click();

    await expect(page.getByText('Updated API Key')).toBeVisible();

    // 4. Delete Credential
    // Accept delete confirmation
    page.on('dialog', dialog => dialog.accept());

    await page.getByRole('button', { name: 'Delete' }).click();
    // In our UI, we might use a custom dialog instead of window.confirm
    // If it's a Radix alert dialog:
    if (await page.getByText('Are you sure?').isVisible()) {
         await page.getByRole('button', { name: 'Delete' }).last().click();
    }

    await expect(page.getByText('No credentials found')).toBeVisible();
  });
});
