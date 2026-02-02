/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ request }) => {
        await seedServices(request);
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');

        await expect(page.getByText('process_payment').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Process a payment').first()).toBeVisible({ timeout: 10000 });

        await expect(page.getByText('calculator').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no tools', async ({ page, request }) => {
        // Clear services to simulate no tools
        await cleanupServices(request);

        await page.goto('/tools');
        // The table shows one row with "No tools found." when empty
        // Use a more flexible locator
        await expect(page.locator('text=No tools found.').or(page.locator('text=No results found'))).toBeVisible();
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        const toolRow = page.locator('tr').filter({ hasText: 'process_payment' });
        await toolRow.getByText('Inspect').click();

        await expect(page.getByRole('dialog')).toBeVisible();
        await expect(page.getByRole('dialog').getByText('process_payment')).toBeVisible();
    });
});
