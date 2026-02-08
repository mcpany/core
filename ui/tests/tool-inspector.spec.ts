/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';
import { login } from './e2e/auth-helper';

test.describe('Tool Inspector', () => {
    test.beforeEach(async ({ request, page }) => {
        await seedUser(request, "e2e-admin");
        await seedServices(request);
        await login(page);
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
        await cleanupUser(request, "e2e-admin");
    });

    test('Tools page loads and inspector opens', async ({ page }) => {
        await page.goto('/tools');

        // Check if calculator tool is listed
        await expect(page.getByText('calculator').first()).toBeVisible({ timeout: 30000 });

        // Open inspector for calculator
        const row = page.locator('tr').filter({ hasText: 'calculator' });
        await row.getByText('Inspect').click();

        // Check if inspector sheet is open
        await expect(page.getByRole('heading', { name: 'calculator' })).toBeVisible();

        // Check if schema is displayed
        await expect(page.getByText('Schema', { exact: true })).toBeVisible();

        // Switch to JSON tab to verify raw schema
        await page.getByRole('tab', { name: 'JSON' }).click();

        // Check for standard schema properties
        await expect(page.locator('pre').filter({ hasText: /"type": "object"/ })).toBeVisible();

        // Verify service name is shown in details
        await expect(page.locator('div[role="dialog"]').getByText('Math New')).toBeVisible();
    });
});
