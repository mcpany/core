/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ page }) => {
       // Real Data Policy: No mocks.
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');

        // weather-service provides get_weather
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Get current weather').first()).toBeVisible({ timeout: 10000 });
    });

    /*
    test.skip('should show empty state when no tools', async ({ page }) => {
        // Cannot easily test empty state with real seeded data unless we delete everything.
    });
    */

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');

        const toolRow = page.locator('tr').filter({ hasText: 'get_weather' });
        await toolRow.getByText('Inspect').click();

        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('get_weather').first()).toBeVisible();
    });
});
