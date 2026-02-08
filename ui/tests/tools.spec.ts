/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.describe.configure({ mode: 'serial' });

    test.beforeEach(async ({ request }) => {
        await seedServices(request);
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');
        // seedServices now adds: "Weather Service" which provides "get_weather"

        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Get current weather').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no tools', async ({ page, request }) => {
        // Clear services to simulate empty state
        await cleanupServices(request);

        await page.goto('/tools');
        // The table shows one row with "No tools found." when empty
        // Ensure table body exists first
        await expect(page.locator('table tbody')).toBeVisible();
        await expect(page.locator('text=No tools found.')).toBeVisible();
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        // Inspection relies on schema being present in the tool definition
        // get_weather is a good candidate
        const toolRow = page.locator('tr').filter({ hasText: 'get_weather' });
        await toolRow.getByText('Inspect').click();

        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('get_weather').first()).toBeVisible();
    });
});
