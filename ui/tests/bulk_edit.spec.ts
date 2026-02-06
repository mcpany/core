/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';

test.describe('Bulk Edit Services', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await cleanupServices(request);
        await cleanupUser(request, "e2e-admin");

        await seedServices(request);
        await seedUser(request, "e2e-admin");

        // Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'e2e-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
        await cleanupUser(request, "e2e-admin");
    });

    test('should bulk edit tags, timeout and env vars', async ({ page }) => {
        await page.goto('/upstream-services');
        await expect(page.getByRole('heading', { level: 1, name: 'Upstream Services' })).toBeVisible();

        // Select "Local CLI" and "Math" (checking checkboxes)
        // Find row by text and click checkbox
        const cliRow = page.getByRole('row').filter({ hasText: 'Local CLI' });
        await cliRow.getByRole('checkbox').click();

        const mathRow = page.getByRole('row').filter({ hasText: 'Math' });
        await mathRow.getByRole('checkbox').click();

        // Check if "Bulk Edit" button appears
        const bulkEditBtn = page.getByRole('button', { name: 'Bulk Edit' });
        await expect(bulkEditBtn).toBeVisible();
        await bulkEditBtn.click();

        // Dialog should open
        const dialog = page.getByRole('dialog');
        await expect(dialog).toBeVisible();
        await expect(dialog.getByText('Bulk Edit Services')).toBeVisible();

        // 1. Add Tag
        await dialog.getByLabel('Add Tags').fill('bulk-test-tag');

        // 2. Set Timeout (assuming input exists - will fail if not implemented yet)
        // We will implement this in next step, so this test expects the UI to exist.
        // If the UI is not there yet, this test fails, which is expected for TDD.
        // However, I should probably write the test assuming the UI will be there.
        await dialog.getByLabel('Timeout').fill('45s');

        // 3. Add Env Var (assuming UI structure)
        // Click "Add Variable"
        await dialog.getByRole('button', { name: 'Add Variable' }).click();
        // Fill Key/Value. Assuming inputs are present.
        // We might need to target by placeholder or test-id if label is ambiguous.
        await dialog.getByPlaceholder('Key').fill('NEW_BULK_ENV');
        await dialog.getByPlaceholder('Value').fill('BULK_VALUE');

        // Apply
        await dialog.getByRole('button', { name: 'Apply Changes' }).click();

        // Wait for dialog to close
        await expect(dialog).toBeHidden();

        // Verify "Services Updated" toast?
        // await expect(page.getByText(/services have been updated/)).toBeVisible();

        // Reload to verify persistence
        await page.reload();
        await expect(page.getByRole('heading', { level: 1, name: 'Upstream Services' })).toBeVisible();

        // Check Tags on UI (ServiceList usually shows tags)
        await expect(cliRow.getByText('bulk-test-tag')).toBeVisible();
        await expect(mathRow.getByText('bulk-test-tag')).toBeVisible();

        // Verify Timeout and Env via UI
        await cliRow.getByRole('button', { name: 'Open menu' }).click();
        await page.getByRole('menuitem', { name: 'Edit' }).click();

        // Verify Timeout in Advanced tab
        await page.getByRole('tab', { name: 'Advanced' }).click();
        await expect(page.getByLabel('Timeout (s)')).toHaveValue('45s');

        // Verify Env Vars in Connection tab
        await page.getByRole('tab', { name: 'Connection' }).click();
        await expect(page.getByDisplayValue('NEW_BULK_ENV')).toBeVisible();
        await expect(page.getByDisplayValue('BULK_VALUE')).toBeVisible();

        // Close sheet
        await page.getByRole('button', { name: 'Save Changes' }).click();

    });
});
