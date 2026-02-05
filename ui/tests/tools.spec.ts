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

        // Wait for tools to be visible
        // seedServices adds 'process_payment' and 'calculator' and 'get_user'
        await expect(page.getByText('process_payment').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Process a payment').first()).toBeVisible({ timeout: 10000 });

        await expect(page.getByText('calculator').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no tools', async ({ page, request }) => {
        // Clear services
        await cleanupServices(request);

        await page.goto('/tools');
        // The table shows one row with "No tools found." when empty
        // Or "No tools found matching your criteria" if filter is active
        // Let's assume the empty state text
        await expect(page.locator('text=No tools found')).toBeVisible({ timeout: 10000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        const toolRow = page.locator('tr').filter({ hasText: 'process_payment' });
        await toolRow.getByText('Inspect').click();

        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('process_payment').first()).toBeVisible();
    });
});
