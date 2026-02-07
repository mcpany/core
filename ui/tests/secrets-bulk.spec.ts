/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Secrets Manager - Bulk Actions', () => {
  test('should allow bulk deletion of secrets', async ({ page }) => {
    const timestamp = Date.now();
    const secrets = [
        { name: `bulk-test-1-${timestamp}`, key: `BULK_KEY_1_${timestamp}`, value: 'val-1' },
        { name: `bulk-test-2-${timestamp}`, key: `BULK_KEY_2_${timestamp}`, value: 'val-2' },
        { name: `bulk-test-3-${timestamp}`, key: `BULK_KEY_3_${timestamp}`, value: 'val-3' },
    ];

    await page.goto('/secrets');

    // Ensure we are on the page
    await expect(page.getByRole('heading', { name: 'API Keys & Secrets' })).toBeVisible();

    // 1. Seed Data: Create 3 secrets
    for (const secret of secrets) {
        await page.getByRole('button', { name: 'Add Secret' }).click();
        await expect(page.getByRole('dialog')).toBeVisible();
        await page.fill('#name', secret.name);
        await page.fill('#key', secret.key);
        await page.fill('#value', secret.value);
        await page.getByRole('button', { name: 'Save Secret' }).click();

        // Wait for dialog to close OR error toast
        try {
            await expect(page.getByRole('dialog')).toBeHidden({ timeout: 5000 });
        } catch (e) {
            // Check for error toast if dialog is still open
            const errorToast = page.getByText(/Failed to save secret/i);
            if (await errorToast.isVisible()) {
                const text = await errorToast.textContent();
                throw new Error(`Failed to save secret: ${text}`);
            }
            throw e;
        }

        // Wait for it to appear in the list
        await expect(page.getByText(secret.name)).toBeVisible();
    }

    // 2. Select 2 secrets
    // Use more specific selector to avoid ambiguity if multiple rows have same text (unlikely here but good practice)
    // Select Secret 1
    await page.locator('tr').filter({ hasText: secrets[0].name }).getByRole('checkbox').check();

    // Select Secret 2
    await page.locator('tr').filter({ hasText: secrets[1].name }).getByRole('checkbox').check();

    // Secret 3 remains unselected
    await expect(page.locator('tr').filter({ hasText: secrets[2].name }).getByRole('checkbox')).not.toBeChecked();

    // 3. Verify Bulk Actions Toolbar appears
    // The toolbar button text contains "Delete (2)"
    const bulkDeleteBtn = page.getByRole('button', { name: /Delete \(\d+\)/ });
    await expect(bulkDeleteBtn).toBeVisible();
    await expect(bulkDeleteBtn).toHaveText(/Delete \(2\)/);

    // 4. Click Bulk Delete and Handle Confirmation Dialog
    await bulkDeleteBtn.click();

    // Verify AlertDialog appears
    const alertDialog = page.getByRole('alertdialog');
    await expect(alertDialog).toBeVisible();
    await expect(alertDialog.getByText(/Are you absolutely sure/)).toBeVisible();

    // Click Delete in AlertDialog
    await alertDialog.getByRole('button', { name: 'Delete' }).click();
    await expect(alertDialog).toBeHidden();

    // 5. Verify Deletion
    // Wait for list to update. The deleted items should disappear.
    await expect(page.getByText(secrets[0].name)).not.toBeVisible();
    await expect(page.getByText(secrets[1].name)).not.toBeVisible();
    await expect(page.getByText(secrets[2].name)).toBeVisible();
  });
});
