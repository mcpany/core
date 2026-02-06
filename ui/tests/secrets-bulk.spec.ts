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
        await expect(page.getByRole('dialog')).toBeHidden();
        // Wait for it to appear
        await expect(page.getByText(secret.name)).toBeVisible();
    }

    // 2. Select 2 secrets
    // Select Secret 1
    const row1 = page.getByRole('row', { name: secrets[0].name });
    await row1.getByRole('checkbox').check();

    // Select Secret 2
    const row2 = page.getByRole('row', { name: secrets[1].name });
    await row2.getByRole('checkbox').check();

    // Secret 3 remains unselected
    const row3 = page.getByRole('row', { name: secrets[2].name });
    await expect(row3.getByRole('checkbox')).not.toBeChecked();

    // 3. Verify Bulk Actions Toolbar appears
    const bulkDeleteBtn = page.getByRole('button', { name: 'Delete', exact: false }).filter({ hasText: 'Delete' }).last();
    await expect(bulkDeleteBtn).toBeVisible();

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
    await expect(page.getByText(secrets[0].name)).not.toBeVisible();
    await expect(page.getByText(secrets[1].name)).not.toBeVisible();
    await expect(page.getByText(secrets[2].name)).toBeVisible();
  });
});
