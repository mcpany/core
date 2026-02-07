/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Secrets Manager - Bulk Actions', () => {
  test('should allow bulk deletion of secrets', async ({ page }) => {
    const timestamp = Date.now();
    const secrets = [
        { name: `bulk-secret-1-${timestamp}`, key: `BULK_KEY_1_${timestamp}`, value: `val-1` },
        { name: `bulk-secret-2-${timestamp}`, key: `BULK_KEY_2_${timestamp}`, value: `val-2` },
        { name: `bulk-secret-3-${timestamp}`, key: `BULK_KEY_3_${timestamp}`, value: `val-3` },
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
        await expect(page.getByText(secret.name)).toBeVisible();
    }

    // 2. Select 2 secrets
    // We expect checkboxes to be present after our refactor.
    // Assuming checkboxes have aria-label `Select {name}` or similar row selection.
    // Or we can find the checkbox within the row.

    // Select secret 1
    const row1 = page.getByRole('row').filter({ hasText: secrets[0].name });
    await row1.getByRole('checkbox').click();

    // Select secret 2
    const row2 = page.getByRole('row').filter({ hasText: secrets[1].name });
    await row2.getByRole('checkbox').click();

    // 3. Verify Toolbar / Bulk Action Button appears
    const deleteButton = page.getByRole('button', { name: /Delete/ }).filter({ hasText: /Delete/ });
    // It might say "Delete (2)" or similar.
    await expect(deleteButton).toBeVisible();

    // 4. Click Delete and Confirm
    // Assuming there might be a confirmation dialog or it just deletes.
    // The ServiceList confirms with `confirm()`.
    // Playwright handles `confirm()` dialogs automatically by dismissing them unless a handler is set.
    // But we want to accept it.
    page.on('dialog', dialog => dialog.accept());
    await deleteButton.click();

    // Wait for success message to ensure operation completed
    await expect(page.getByText('secrets deleted successfully').first()).toBeVisible();

    // 5. Verify Deletion
    await expect(page.getByText(secrets[0].name)).not.toBeVisible();
    await expect(page.getByText(secrets[1].name)).not.toBeVisible();

    // Verify 3rd secret remains
    await expect(page.getByText(secrets[2].name)).toBeVisible();
  });
});
