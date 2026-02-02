/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ page, request }) => {
        // Seed tools via services
        // Use command_line_service to avoid SSRF blocking on localhost addresses in CI.
        const seedData = {
            services: [
                {
                    name: "weather-service",
                    command_line_service: {
                        command: "echo",
                        args: ["weather-service-ready"],
                        tools: [
                            {
                                name: "weather-tool",
                                description: "Get weather for a location",
                                input_schema: {
                                    type: "object",
                                    properties: {
                                        location: {type: "string", description: "City name"}
                                    }
                                }
                            }
                        ]
                    }
                },
                {
                    name: "math-service",
                    command_line_service: {
                        command: "echo",
                        args: ["math-service-ready"],
                        tools: [
                            {name: "calculator", description: "Perform basic math"}
                        ]
                    }
                }
            ]
        };

        const headers: any = {};
        if (process.env.MCPANY_API_KEY) {
            headers['X-API-Key'] = process.env.MCPANY_API_KEY;
        } else {
            headers['X-API-Key'] = 'test-token';
        }

        const res = await request.post('/api/v1/debug/seed_state', {
            data: seedData,
            headers: headers
        });
        expect(res.ok()).toBeTruthy();
    });

    test('should list available tools', async ({ page }) => {
        await page.goto('/tools');

        await expect(page.getByText('weather-tool').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Get weather for a location').first()).toBeVisible({ timeout: 10000 });

        await expect(page.getByText('calculator').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('Perform basic math').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no tools', async ({ page, request }) => {
        // Clear state
        const headers: any = {};
        if (process.env.MCPANY_API_KEY) {
            headers['X-API-Key'] = process.env.MCPANY_API_KEY;
        } else {
            headers['X-API-Key'] = 'test-token';
        }

        await request.post('/api/v1/debug/seed_state', {
            data: { services: [] }, // Empty services
            headers: headers
        });

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
