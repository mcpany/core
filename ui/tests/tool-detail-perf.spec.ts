/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Tool Detail Performance Optimization', () => {
    test('should load tool details and metrics correctly', async ({ page }) => {
        const serviceId = 'test-service';
        const toolName = 'test-tool';

        // Mock gRPC call (failure to force fallback)
        await page.route('**/*RegistrationService/GetService', async (route) => {
            await route.abort();
        });

        // Mock Service Details (REST)
        await page.route(`**/api/v1/services/${serviceId}`, async (route) => {
             await route.fulfill({
                json: {
                    service: {
                        name: serviceId,
                        http_service: {
                            tools: [
                                {
                                    name: toolName,
                                    description: 'A test tool',
                                    inputSchema: { type: 'object', properties: {} }
                                }
                            ]
                        }
                    }
                }
            });
        });

        // Mock Service Status (Metrics)
        await page.route(`**/api/v1/services/${serviceId}/status`, async (route) => {
            // Add a small delay to simulate network latency
             await new Promise(resolve => setTimeout(resolve, 100));
            await route.fulfill({
                json: {
                    metrics: {
                        [`tool_usage:${toolName}`]: 42
                    }
                }
            });
        });

        await page.goto(`/service/${serviceId}/tool/${toolName}`);

        // Verify Tool Name
        await expect(page.getByText(toolName).first()).toBeVisible();

        // Verify Tool Description
        await expect(page.getByText('A test tool')).toBeVisible();

        // Verify Metrics
        await expect(page.getByText('42')).toBeVisible();
    });

    test('should handle missing service gracefully', async ({ page }) => {
        const serviceId = 'missing-service';
        const toolName = 'test-tool';

        await page.route('**/*RegistrationService/GetService', async (route) => {
            await route.abort();
        });

        await page.route(`**/api/v1/services/${serviceId}`, async (route) => {
            await route.fulfill({ status: 404, body: 'Not Found' });
        });

         await page.route(`**/api/v1/services/${serviceId}/status`, async (route) => {
            await route.fulfill({ status: 404 });
        });

        await page.goto(`/service/${serviceId}/tool/${toolName}`);

        await expect(page.getByRole('alert').filter({ hasText: /not found|404|error|failed/i })).toBeVisible();
    });
});
