/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

import { apiClient } from '../../src/lib/client';

test.describe('Credentials Management', () => {

  test.beforeEach(async () => {
    // Ensure clean state before test
    try {
        const credentials = await apiClient.listCredentials();
        for (const cred of credentials) {
            if (cred.name.includes('Test API Key') || cred.name.includes('Updated API Key')) {
                 await apiClient.deleteCredential(cred.id);
            }
        }
    } catch (e) {
        console.warn('Cleanup failed:', e);
    }
  });

  test('should list, create, update and delete credentials with polished UI', async ({ page }) => {
    // 1. Initial List (Empty)
    await page.goto('/credentials');

    // Verify Polished Empty State
    await expect(page.getByText('No credentials found')).toBeVisible();
    await expect(page.getByText('Add credentials to authenticate your MCP Any instance with external services.')).toBeVisible();

    // Use the Empty State CTA to open dialog
    await page.getByRole('button', { name: 'Create First Credential' }).click();
    await expect(page.getByText('Create Credential')).toBeVisible();

    // 2. Create Credential via UI
    await page.getByPlaceholder('My Credential').fill('Test API Key');
    // Default format is API Key, so just fill details
    await page.getByPlaceholder('X-API-Key').fill('Authorization');
    await page.getByPlaceholder('...secret key...').fill('secret-key');

    await page.getByRole('button', { name: 'Save' }).click();

    // Verify it appears in list with polished UI elements (Badges, etc)
    await expect(page.getByText('Test API Key')).toBeVisible({ timeout: 10000 });
    // Verify the Badge is rendered
    await expect(page.locator('tbody').getByText('API Key', { exact: true })).toBeVisible();

    // 3. Update Credential
    await page.getByRole('button', { name: 'Edit' }).first().click();
    await page.getByPlaceholder('My Credential').fill('Updated API Key');
    await page.getByRole('button', { name: 'Save' }).click();

    await expect(page.getByText('Updated API Key')).toBeVisible({ timeout: 10000 });

    // 4. Delete Credential
    // Accept delete confirmation
    page.once('dialog', dialog => dialog.accept());

    await page.getByRole('button', { name: 'Delete' }).first().click();

    // Should return to empty state
    await expect(page.getByText('No credentials found')).toBeVisible({ timeout: 10000 });
  });
});
