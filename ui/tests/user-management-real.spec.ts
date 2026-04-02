/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedUser, seedGlobalState } from './e2e/test-data';

test.describe('User Management - Bulk Delete (Real Data)', () => {
    test.beforeAll(async ({ request }) => {
        try {
            await seedGlobalState(request);
            await seedUser(request, "e2e-user-1");
            await seedUser(request, "e2e-user-2");
            // e2e-admin-core is created by seedGlobalState
        } catch (e) {
            console.warn('Backend not available for seeding, proceeding without it:', e);
        }
    });

    test('should allow selecting multiple users and deleting them in bulk', async ({ page }) => {
        // Go to the users page
        await page.goto('/users');

        // Wait for users to render (real data from backend)
        await expect(page.locator('text=e2e-user-1').first()).toBeVisible();
        await expect(page.locator('text=e2e-user-2').first()).toBeVisible();
        await expect(page.locator('text=e2e-admin-core').first()).toBeVisible();

        // Check the checkboxes for user 1 and user 2 using aria-labels
        await page.getByRole('checkbox', { name: 'Select e2e-user-1' }).check();
        await page.getByRole('checkbox', { name: 'Select e2e-user-2' }).check();

        // Verify the bulk delete banner appears with correct count
        await expect(page.locator('text=2 selected').first()).toBeVisible();

        // Setup a listener for dialogs (window.confirm)
        page.on('dialog', dialog => dialog.accept());

        // Click Bulk Delete
        await page.locator('button:has-text("Bulk Delete")').click();

        // Verify success toast appears
        await expect(page.locator('text=Users Deleted').first()).toBeVisible();

        // Verify users are removed from the list
        await expect(page.locator('text=e2e-user-1')).toHaveCount(0);
        await expect(page.locator('text=e2e-user-2')).toHaveCount(0);

        // Verify admin is still there
        await expect(page.locator('text=e2e-admin-core').first()).toBeVisible();
    });
});
