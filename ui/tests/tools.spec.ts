/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ page }) => {
        // Mock tools endpoint directly (matching ToolsPage fetch)
        await page.route((url) => url.pathname.includes('/api/v1/tools'), async (route) => {
            await route.fulfill({
                json: {
                    tools: [
                        {
                            name: 'weather-tool',
                            description: 'Get weather for a location',
                            source: 'configured',
                            serviceId: 'weather-service',
                            inputSchema: {
                                type: 'object',
                                properties: {
                                    location: { type: 'string', description: 'City name' }
                                }
                            }
                        },
                        {
                            name: 'calculator',
                            description: 'Perform basic math',
                            source: 'discovered',
                            serviceId: 'math-service'
                        }
                    ]
                }
            });
        });
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');
        // Wait for loading to finish (if applicable, though mock is instant)
        // Adjust selector if you add a specific loading state for tools

        await expect(page.getByText('weather-tool').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Get weather for a location').first()).toBeVisible({ timeout: 10000 });

        await expect(page.getByText('calculator').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Perform basic math').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no tools', async ({ page }) => {
        // Unroute previous mock from beforeEach
        await page.unroute((url) => url.pathname.includes('/api/v1/tools'));
        await page.route((url) => url.pathname.includes('/api/v1/tools'), async (route) => {
            await route.fulfill({ json: { tools: [] } });
        });

        await page.goto('/tools');
        // The table shows one row with "No tools found." when empty
        await expect(page.locator('table tbody tr')).toHaveCount(1);
        await expect(page.locator('text=No tools found.')).toBeVisible();
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
