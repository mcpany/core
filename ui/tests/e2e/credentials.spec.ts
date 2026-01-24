/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Credentials Management', () => {



  test('should list, create, update and delete credentials', async ({ page }) => {
    // 1. Initial List (Should accept empty or existing)
    await page.goto('/credentials');
    // Ensure the page loads (wait for table or empty text)
    // Use .or() to handle either state
    await expect(page.locator('table').or(page.getByText('No credentials found'))).toBeVisible({ timeout: 10000 });


    // 2. Create Credential
    const timestamp = Date.now();
    const credName = `Test Cred ${timestamp}`;
    const updatedCredName = `Updated Cred ${timestamp}`;

    await page.getByRole('button', { name: 'New Credential' }).click();
    await expect(page.getByText('Create Credential')).toBeVisible();

    await page.getByPlaceholder('My Credential').fill(credName);
    // Default format is API Key, so just fill details
    await page.getByPlaceholder('X-API-Key').fill('Authorization');
    await page.getByPlaceholder('...secret key...').fill('secret-key');

    await page.getByRole('button', { name: 'Save' }).click({ force: true });

    // Verify it appears in list
    // Wait for list to update or reload
    await expect(page.getByText(credName)).toBeVisible({ timeout: 10000 });
    await expect(page.locator('tbody').getByText('API Key', { exact: true }).filter({ hasText: 'API Key' }).first()).toBeVisible();


    // 3. Update Credential
    // Find the row with our credential
    const row = page.locator('tr').filter({ hasText: credName });
    await row.getByRole('button', { name: 'Edit' }).click();

    await page.getByPlaceholder('My Credential').fill(updatedCredName);
    await page.getByRole('button', { name: 'Save' }).click({ force: true });

    await expect(page.getByText(updatedCredName)).toBeVisible();
    await expect(page.getByText(credName)).toBeHidden();

    // 4. Delete Credential
    const updatedRow = page.locator('tr').filter({ hasText: updatedCredName });

    // Accept delete confirmation
    page.on('dialog', dialog => dialog.accept());

    await updatedRow.getByRole('button', { name: 'Delete' }).click({ force: true });

    // Handle custom dialog if exists (Radix UI)
    // We check if the dialog content is visible
    try {
        const dialog = page.getByRole('alertdialog');
        if (await dialog.isVisible({ timeout: 2000 })) {
             await dialog.getByRole('button', { name: 'Delete' }).click({ force: true });
        }
    } catch (e) {
        // Ignore if no custom dialog
    }

    await expect(page.getByText(updatedCredName)).toBeHidden();
  });
});
