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
        // Ensure we match the snake_case format that the server would return
        // and that the client expects to map to camelCase.
        await page.route(`**/api/v1/services/${serviceId}`, async (route) => {
             await route.fulfill({
                status: 200,
                contentType: 'application/json',
                json: {
                    service: {
                        id: serviceId,
                        name: serviceId,
                        version: "1.0.0",
                        http_service: {
                            address: "http://localhost:8080",
                            tools: [
                                {
                                    name: toolName,
                                    description: 'A test tool',
                                    input_schema: { type: 'object', properties: {} }, // snake_case for safety
                                    inputSchema: { type: 'object', properties: {} }   // camelCase just in case
                                }
                            ]
                        }
                    }
                }
            });
        });

        // Mock Tool Usage (Metrics)
        // The URL matching must be precise. The client appends ?serviceId=...
        // We use a glob that covers query params.
        await page.route(`**/api/v1/dashboard/tool-usage*`, async (route) => {
            // Add a small delay to simulate network latency
             await new Promise(resolve => setTimeout(resolve, 100));
            await route.fulfill({
                status: 200,
                contentType: 'application/json',
                json: [{
                    name: toolName,
                    serviceId: serviceId,
                    totalCalls: 42,
                    successRate: 95.0,
                    avgLatency: 150.0,
                    errorCount: 2
                }]
            });
        });

        await page.goto(`/service/${serviceId}/tool/${toolName}`);

        // Verify Tool Name
        await expect(page.getByText(toolName).first()).toBeVisible();

        // Verify Tool Description
        await expect(page.getByText('A test tool')).toBeVisible();

        // Verify Metrics
        await expect(page.getByText('42', { exact: true })).toBeVisible();
        await expect(page.getByText('95.0%')).toBeVisible();
        await expect(page.getByText('150ms')).toBeVisible();
        await expect(page.getByText('2', { exact: true })).toBeVisible();
    });

    test('should handle missing service gracefully', async ({ page }) => {
        const serviceId = 'missing-service';
        const toolName = 'test-tool';

        await page.route('**/*RegistrationService/GetService', async (route) => {
            await route.abort();
        });

        await page.route(`**/api/v1/services/${serviceId}`, async (route) => {
            await route.fulfill({ status: 404, body: '{"error": "Not Found"}' });
        });

         await page.route(`**/api/v1/services/${serviceId}/status`, async (route) => {
            await route.fulfill({ status: 404 });
        });

        await page.goto(`/service/${serviceId}/tool/${toolName}`);

        // Expect some error UI
        await expect(page.getByText(/service not found/i)).toBeVisible();
    });
});
