/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Exploration', () => {
    test.beforeEach(async ({ page }) => {
        // Mock resources endpoint directly
        await page.route('/api/resources', async (route) => {
            await route.fulfill({
                json: {
                  resources: [
                    {
                        name: 'Application Logs',
                        mimeType: 'text/plain',
                        serviceName: 'log-service',
                        uri: 'file:///var/log/app.log',
                        enabled: true
                    },
                    {
                        name: 'User Database',
                        mimeType: 'application/x-postgres',
                        serviceName: 'db-service',
                        uri: 'postgres://localhost:5432/users',
                        enabled: true
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
        // Description is not currently shown in the table
        // await expect(page.getByText('Main application logs')).toBeVisible();
        await expect(page.getByText('text/plain')).toBeVisible();

        await expect(page.getByText('User Database').first()).toBeVisible({ timeout: 10000 });
    });

    test('should show empty state when no resources', async ({ page }) => {
        await page.route('/api/resources', async (route) => {
            await route.fulfill({ json: { resources: [] } });
        });

        await page.goto('/resources');
        await expect(page.locator('table tbody tr')).toHaveCount(0);
    });
});
