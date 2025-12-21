/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Exploration', () => {
    test.beforeEach(async ({ page }) => {
        // Mock services endpoint which aggregates resources
        await page.route('**/v1/services', async (route) => {
            await route.fulfill({
                json: {
                    services: [
                        {
                            name: 'log-service',
                            http_service: {
                                address: 'http://localhost',
                                resources: [
                                    {
                                        uri: 'file:///logs/app.log',
                                        name: 'Application Logs',
                                        mimeType: 'text/plain',
                                        description: 'Main application logs'
                                    }
                                ]
                            }
                        },
                        {
                            name: 'db-service',
                            grpc_service: {
                                address: 'localhost:9090',
                                resources: [
                                    {
                                        uri: 'postgres://db/users',
                                        name: 'User Database',
                                        mimeType: 'application/x-postgres',
                                        description: 'User records'
                                    }
                                ]
                            }
                        }
                    ]
                }
            });
        });
    });

    test('should list available resources', async ({ page }) => {
        await page.goto('/resources');

        // Use first() to avoid ambiguity if multiple elements match text (e.g. name vs description)
        await expect(page.getByText('Application Logs').first()).toBeVisible();
        await expect(page.getByText('Main application logs')).toBeVisible();
        await expect(page.getByText('text/plain')).toBeVisible();

        await expect(page.getByText('User Database')).toBeVisible();
    });

    test('should show empty state when no resources', async ({ page }) => {
        await page.route('**/v1/services', async (route) => {
            await route.fulfill({ json: { services: [] } });
        });

        await page.goto('/resources');
        await expect(page.getByText('No resources available')).toBeVisible();
    });
});
