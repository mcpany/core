/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request, page }) => {
        await seedServices(request);
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');
        // seedServices adds "Payment Gateway" with "process_payment", "User Service" with "get_user", "Math" with "calculator"
        await expect(page.getByText('process_payment').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Process a payment').first()).toBeVisible({ timeout: 10000 });

        await expect(page.getByText('calculator').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no tools', async ({ request, page }) => {
        // Clear services to simulate empty state
        await cleanupServices(request);

        await page.goto('/tools');
        // The table shows one row with "No tools found." when empty
        await expect(page.locator('table tbody tr')).toHaveCount(1);
        await expect(page.locator('text=No tools found.')).toBeVisible();
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        const toolRow = page.locator('tr').filter({ hasText: 'calculator' });
        await toolRow.getByText('Inspect').click();

        // Check for drawer content. 'calculator' tool has description 'calc' in seedServices
        await expect(page.getByText('calc', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('calculator').first()).toBeVisible();
    });
});
