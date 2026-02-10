/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './e2e/test-data';

test.describe('Tool Management (Real Data)', () => {
    test.beforeEach(async ({ request }) => {
        await seedServices(request);
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
    });

    test('should list available tools from backend', async ({ page }) => {
        await page.goto('/tools');
        // Check for the builtin/config tool (weather-service -> get_weather)
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 15000 });
    });

    test('should execute a tool and show results', async ({ page }) => {
        await page.goto('/tools');

        const toolRow = page.locator('tr', { hasText: 'get_weather' });
        await toolRow.getByRole('button', { name: 'Inspect' }).click();

        await expect(page.getByRole('dialog')).toBeVisible();

        // Click Execute (default input is fine for this tool)
        await page.getByRole('button', { name: 'Execute' }).click();

        // Verify Result
        // The output is the result of `echo '{"weather": "sunny"}'`
        // It should contain "sunny" in the stdout field.
        await expect(page.getByText('sunny')).toBeVisible({ timeout: 10000 });
    });
});
