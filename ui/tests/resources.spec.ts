/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resource Exploration', () => {
    const serviceName = 'resource-test-service';

    test.beforeAll(async ({ request }) => {
        // Clean up
        await request.delete(`/api/v1/services/${serviceName}`).catch(() => { });

        // Seed service with actual resources to fetch
        const response = await request.post('/api/v1/services', {
            data: {
                name: serviceName,
                command_line_service: {
                    command: 'echo',
                    resources: [
                        {
                            name: 'Application Logs',
                            uri: `mcp://${serviceName}/logs.txt`,
                            mime_type: 'text/plain',
                            description: 'Logs',
                            static: {
                                text_content: 'some application logs here'
                            }
                        },
                        {
                            name: 'User Database',
                            uri: `mcp://${serviceName}/db.json`,
                            mime_type: 'application/json',
                            description: 'Database',
                            static: {
                                text_content: '{"users": [{"id": 1, "name": "Alice"}]}'
                            }
                        }
                    ]
                }
            }
        });
        if (!response.ok()) {
            console.error('Failed to seed resources service', await response.text());
        }
        expect(response.ok()).toBeTruthy();
    });

    test.afterAll(async ({ request }) => {
        await request.delete(`/api/v1/services/${serviceName}`).catch(() => { });
    });

    test('should list available resources from real database', async ({ page }) => {
        await page.goto('/resources');

        // Real data fetch verification
        await expect(page.getByText('Application Logs').first()).toBeVisible({ timeout: 10000 });
        await expect(page.getByText('User Database').first()).toBeVisible({ timeout: 10000 });

        // Verify JSON View renders when clicking User Database
        await page.getByText('User Database').first().click();

        // Let's verify that the JSON viewer is active for this JSON resource.
        // It renders "users", "1", "Alice" inside the UI
        await expect(page.getByText('Alice')).toBeVisible({ timeout: 10000 });
    });
});
