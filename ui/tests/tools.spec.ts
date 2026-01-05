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
                            serviceName: 'weather-service'
                        },
                        {
                            name: 'calculator',
                            description: 'Perform basic math',
                            source: 'discovered',
                            serviceName: 'math-service'
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
        await page.route((url) => url.pathname.includes('/api/v1/tools'), async (route) => {
            await route.fulfill({ json: [] });
        });

        await page.goto('/tools');
        await expect(page.locator('table tbody tr')).toHaveCount(0);
    });
});
