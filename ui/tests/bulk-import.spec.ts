/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Import Feature', () => {
  test('should import a service via wizard', async ({ page }) => {
    // 1. Navigate
    await page.goto('/upstream-services');

    // 2. Open Wizard
    await page.getByRole('button', { name: 'Bulk Import' }).click();
    await expect(page.getByRole('heading', { name: 'Bulk Service Import' })).toBeVisible();

    // 3. Input JSON
    // Default tab is "JSON Text"
    const jsonConfig = JSON.stringify([
      {
        name: "test-bulk-service",
        httpService: {
          address: "https://example.com"
        }
      }
    ]);

    await page.getByLabel('Paste JSON Configuration').fill(jsonConfig);

    // 4. Click Next
    await page.getByRole('button', { name: 'Next: Validate' }).click();

    // 5. Verify Selection Step
    // Wait for validation
    await expect(page.getByText('Review & Select')).toBeVisible();
    await expect(page.getByText('Found 1 services. 1 valid.')).toBeVisible();

    // Verify row
    await expect(page.getByRole('cell', { name: 'test-bulk-service', exact: true })).toBeVisible();
    await expect(page.getByRole('cell', { name: 'Valid' })).toBeVisible();

    // 6. Import
    // Checkbox should be auto-selected for valid items
    await page.getByRole('button', { name: 'Import 1 Services' }).click();

    // 7. Verify Result
    await expect(page.getByText('Import Complete')).toBeVisible();
    await expect(page.getByText('Successfully imported 1 services.')).toBeVisible();

    // 8. Close and Verify List
    await page.getByRole('button', { name: 'Close & View Services' }).click();
    await expect(page.getByRole('dialog')).toBeHidden();

    // Verify service appears in list
    await expect(page.getByRole('cell', { name: 'test-bulk-service', exact: true })).toBeVisible();

    // Cleanup
    // Locate the row with the service name, find the actions menu, delete it.
    // This assumes the list is updated.

    // Click actions menu for the row
    const row = page.getByRole('row').filter({ hasText: 'test-bulk-service' });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();

    // Confirm dialog
    page.on('dialog', dialog => dialog.accept());
    // The app uses window.confirm for delete
  });
});
