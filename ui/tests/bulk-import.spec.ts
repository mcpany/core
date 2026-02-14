/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Service Import Verification', () => {
  test('should verify the new bulk import wizard flow', async ({ page }) => {
    // 1. Navigate to Services Page
    await page.goto('/upstream-services');
    await expect(page.getByRole('heading', { level: 1, name: 'Upstream Services' })).toBeVisible();

    // 2. Click Bulk Import Button
    await page.getByRole('button', { name: 'Bulk Import' }).click();

    // 3. Wizard - Step 1: Source
    await expect(page.getByRole('dialog')).toBeVisible();
    await expect(page.getByRole('tab', { name: 'Paste JSON' })).toBeVisible();

    // Fill JSON with a valid dummy service
    // Note: Use a name that doesn't conflict or handle cleanup
    const serviceName = `test-import-${Date.now()}`;
    const jsonContent = JSON.stringify([{
        name: serviceName,
        httpService: { address: 'http://example.com' }
    }]);

    await page.getByLabel('Service Configuration (JSON)').fill(jsonContent);

    // Click Review
    await page.getByRole('button', { name: 'Review' }).click();

    // 4. Wizard - Step 2: Validation
    // Wait for validation table to appear
    await expect(page.getByText(`Found 1 services`)).toBeVisible();
    // Be specific about finding 'Valid' status in the table
    await expect(page.getByRole('cell').filter({ hasText: 'Valid' })).toBeVisible();
    await expect(page.getByText(serviceName)).toBeVisible();

    // 5. Wizard - Step 3: Import
    // Verify import button shows correct count
    await expect(page.getByRole('button', { name: 'Import (1)' })).toBeVisible();

    // Click Import
    await page.getByRole('button', { name: 'Import (1)' }).click();

    // 6. Wizard - Step 4: Result
    await expect(page.getByRole('heading', { name: 'Import Complete' })).toBeVisible();
    await expect(page.getByText('1 successful')).toBeVisible();

    // Close Wizard
    await page.getByRole('button', { name: 'Done' }).click();

    // 7. Verify Service in List
    // We might need to wait for list refresh or ensure it's there
    // The wizard calls onImportSuccess which triggers fetchServices
    // Force reload to be sure if hot update missed
    await page.reload();
    await expect(page.getByRole('link', { name: serviceName })).toBeVisible();

    // Cleanup: Delete the service
    // Setup dialog handler before clicking delete
    page.on('dialog', dialog => dialog.accept());

    // Find row with service name, click delete in dropdown
    const row = page.getByRole('row').filter({ hasText: serviceName });
    await row.getByRole('button', { name: 'Open menu' }).click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();

    // Wait for it to disappear
    await expect(page.getByRole('link', { name: serviceName })).not.toBeVisible();
  });
});
