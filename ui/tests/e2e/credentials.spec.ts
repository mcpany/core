/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Credentials Management', () => {



  test('should list, create, update and delete credentials', async ({ page, request }) => {
    // 1. Initial List (Empty)
    // Clear out credentials completely via seed debug endpoint
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

    await page.goto('/credentials');
    await expect(page.getByText('No credentials found')).toBeVisible();

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
