/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ request }) => {
        // Register services with tools using the API
        // weather-service
        await request.post('/api/v1/services', {
            data: {
                name: 'weather-service',
                httpService: { address: 'http://localhost:9999' }, // Dummy address
                tools: [
                    {
                        name: 'weather-tool',
                        description: 'Get weather for a location',
                        inputSchema: JSON.stringify({
                            type: 'object',
                            properties: {
                                location: { type: 'string', description: 'City name' }
                            }
                        })
                    }
                ]
            }
        });

        // math-service
        await request.post('/api/v1/services', {
            data: {
                name: 'math-service',
                httpService: { address: 'http://localhost:9998' },
                tools: [
                    {
                        name: 'calculator',
                        description: 'Perform basic math'
                    }
                ]
            }
        });
    });

    test.afterEach(async ({ request }) => {
        // Cleanup
        await request.delete('/api/v1/services/weather-service');
        await request.delete('/api/v1/services/math-service');
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');

        await expect(page.getByText('weather-tool').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Get weather for a location').first()).toBeVisible({ timeout: 10000 });

        await expect(page.getByText('calculator').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Perform basic math').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no tools', async ({ page, request }) => {
        // Unregister services to ensure empty state
        await request.delete('/api/v1/services/weather-service');
        await request.delete('/api/v1/services/math-service');

        await page.goto('/tools');
        // The table shows one row with "No tools found." when empty
        await expect(page.locator('table tbody tr')).toHaveCount(1);
        await expect(page.locator('text=No tools found.')).toBeVisible();
    });

    test('should allow inspecting a tool', async ({ page }) => {
        await page.goto('/tools');
        const toolRow = page.locator('tr').filter({ hasText: 'weather-tool' });
        await toolRow.getByText('Inspect').click();

        await expect(page.getByText('Schema', { exact: true }).first()).toBeVisible();
        await expect(page.getByText('weather-tool').first()).toBeVisible();
    });
});
