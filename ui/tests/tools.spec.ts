/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';

test.describe('Tools Management', () => {
    // Run sequentially to avoid DB collision
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await seedServices(request);
        await seedUser(request, "tools-admin");

        // Login
        await page.goto('/login');
        await page.fill('input[name="username"]', 'tools-admin');
        await page.fill('input[name="password"]', 'password');
        await page.click('button[type="submit"]');
        await expect(page).toHaveURL('/', { timeout: 15000 });
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
        await cleanupUser(request, "tools-admin");
    });

    test('should list available tools from backend', async ({ page }) => {
        await page.goto('/tools');
        // Check for real seeded tools
        await expect(page.getByText('process_payment').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('get_user').first()).toBeVisible();
        await expect(page.getByText('calculator').first()).toBeVisible();
    });

    test('should support bulk actions', async ({ page }) => {
        await page.goto('/tools');
        await expect(page.getByText('process_payment').first()).toBeVisible({ timeout: 10000 });

        // Select 'process_payment'
        const row1 = page.locator('tr', { hasText: 'process_payment' });
        // Radix Checkbox is usually a button with role="checkbox"
        await row1.getByRole('checkbox').click();

        // Select 'calculator'
        const row2 = page.locator('tr', { hasText: 'calculator' });
        await row2.getByRole('checkbox').click();

        // Verify Bulk Action Bar appears
        await expect(page.getByText('2 selected')).toBeVisible();

        // Click Disable (Use specific selector to avoid matching "Disabled" text in table)
        await page.locator('button', { hasText: 'Disable' }).click();

        // Verify Status in UI updates to "Disabled"
        // We wait for the "Disabled" text to appear in the row
        await expect(row1.getByText('Disabled')).toBeVisible();
        await expect(row2.getByText('Disabled')).toBeVisible();

        // Verify 'get_user' is still Enabled (it wasn't selected)
        const row3 = page.locator('tr', { hasText: 'get_user' });
        await expect(row3.getByText('Enabled')).toBeVisible();
    });

    test('should select all tools', async ({ page }) => {
        await page.goto('/tools');
        await expect(page.getByText('process_payment').first()).toBeVisible();

        // Click "Select All" in header
        // Header checkbox is the first checkbox on the page usually, or inside thead
        await page.locator('thead').getByRole('checkbox').click();

        // Verify Bulk Action Bar appears with correct count (3 seeded tools)
        await expect(page.getByText('3 selected')).toBeVisible();

        // Unselect all
        await page.locator('thead').getByRole('checkbox').click();
        await expect(page.getByText('3 selected')).not.toBeVisible();
    });
});
