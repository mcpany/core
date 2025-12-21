/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Tool Exploration', () => {
    test.beforeEach(async ({ page }) => {
        // Mock services endpoint which aggregates tools
        await page.route('**/v1/services', async (route) => {
            await route.fulfill({
                json: {
                    services: [
                        {
                            name: 'weather-service',
                            http_service: {
                                address: 'http://localhost',
                                tools: [
                                    {
                                        name: 'weather-tool',
                                        description: 'Get weather for a location',
                                        input_schema: {
                                            type: 'object',
                                            properties: {
                                                location: { type: 'string' }
                                            }
                                        }
                                    }
                                ]
                            }
                        },
                        {
                            name: 'math-service',
                            command_line_service: {
                                command: 'calc',
                                tools: [
                                    {
                                        name: 'calculator',
                                        description: 'Perform basic math',
                                        input_schema: { type: 'object' }
                                    }
                                ]
                            }
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

        await expect(page.getByText('weather-tool')).toBeVisible();
        await expect(page.getByText('Get weather for a location')).toBeVisible();

        await expect(page.getByText('calculator')).toBeVisible();
        await expect(page.getByText('Perform basic math')).toBeVisible();
    });

    test('should show empty state when no tools', async ({ page }) => {
        await page.route('**/v1/services', async (route) => {
            await route.fulfill({ json: { services: [] } });
        });

        await page.goto('/tools');
        await expect(page.getByText('No tools available')).toBeVisible();
    });
});
