/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedCollection } from './e2e/test-data';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ page, request }) => {
        await request.post('/api/v1/debug/reset');
    });

    test('should list available tools', async ({ page, request }) => {
        // Seed services/tools
        // We can reuse seedServices from test-data or seedCollection
        await seedCollection('test-collection', request);

        await page.goto('/tools');
        // Wait for loading to finish
        // weather-tool comes from seedCollection
        await expect(page.getByText('weather').first()).toBeVisible({ timeout: 15000 });
    });

    test('should show empty state when no tools', async ({ page }) => {
        // Reset ensures no tools
        await page.goto('/tools');
        // The table shows one row with "No tools found." when empty
        await expect(page.locator('text=No tools found.')).toBeVisible({ timeout: 10000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        // Inspection relies on schema being present in the tool definition
        // The mock in beforeEach includes a basic definition
        const toolRow = page.locator('tr').filter({ hasText: 'weather-tool' });
        await toolRow.getByText('Inspect').click();

        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('weather-tool').first()).toBeVisible();
    });
});
