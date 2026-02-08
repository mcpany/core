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

        // Expect built-in tools (Weather Service) which is more reliable in CI
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 15000 });
    });

    test('should show empty state when no tools', async ({ page, request }) => {
        // Cleanup tools for this test
        await cleanupServices(request);
        // Also ensure weather-service is removed (it might be re-added by reload, but let's try)
        await request.delete('/api/v1/services/weather-service', { headers: { 'X-API-Key': process.env.MCPANY_API_KEY || 'test-token' } }).catch(() => {});

        await page.goto('/tools');
        await page.reload(); // Ensure fresh state

        // The table shows one row with "No tools found." when empty
        await expect(page.locator('text=No tools found.')).toBeVisible({ timeout: 15000 });
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        // Wait for tool to appear first
        await expect(page.getByText('get_weather').first()).toBeVisible({ timeout: 15000 });

        const toolRow = page.locator('tr').filter({ hasText: 'get_weather' });
        await toolRow.getByText('Inspect').click();

        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('get_weather').first()).toBeVisible();
    });
});
