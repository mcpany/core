/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ page, request }) => {
        // Use real data seeding instead of mocks
        await seedServices(request);
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');
        // Wait for loading to finish

        // Expect seeded tools
        // "process_payment" from Payment Gateway
        await expect(page.getByText('process_payment').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Process a payment').first()).toBeVisible({ timeout: 10000 });

        // "calculator" from Math service
        await expect(page.getByText('calculator').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('calc').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no tools', async ({ page, request }) => {
        // Mock empty tools response to ensure reliable empty state testing
        await page.route('**/api/v1/tools*', async route => {
            await route.fulfill({ json: [] });
        });

        await page.goto('/tools');
        // The table shows one row with "No tools found." when empty
        await expect(page.locator('table tbody tr')).toHaveCount(1);
        await expect(page.locator('text=No tools found.')).toBeVisible();
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');

        // Inspect 'process_payment'
        const toolRow = page.locator('tr').filter({ hasText: 'process_payment' });
        await toolRow.getByText('Inspect').click();

        // Schema should be visible (if seeded, which it is implicitly via service config,
        // though seedServices in test-data.ts only provides name/desc in tool definition for HTTP service.
        // The backend might auto-discover or use provided info.
        // The seeded data: tools: [{ name: "process_payment", description: "Process a payment" }]
        // Does it have inputSchema? No.
        // So Schema might be empty or default.
        // But the test expects 'Schema' text.
        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('process_payment').first()).toBeVisible();
    });
});
