/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ page, request }) => {
        await seedServices(request);
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');

        // Seeded services include 'Payment Gateway' with 'process_payment' and 'Math' with 'calculator'
        await expect(page.getByText('process_payment').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Process a payment').first()).toBeVisible({ timeout: 10000 });

        await expect(page.getByText('calculator').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no tools', async ({ page, request }) => {
        // Cleanup all services to ensure no tools
        await cleanupServices(request);

        await page.goto('/tools');
        // The table shows one row with "No tools found." when empty
        await expect(page.locator('table tbody tr')).toHaveCount(1);
        await expect(page.locator('text=No tools found.')).toBeVisible();
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        // Inspection relies on schema being present in the tool definition
        // We use process_payment which has a description
        const toolRow = page.locator('tr').filter({ hasText: 'process_payment' });
        await toolRow.getByText('Inspect').click();

        // Check for side sheet
        await expect(page.getByText('Tool Details', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('process_payment').first()).toBeVisible();
    });
});
