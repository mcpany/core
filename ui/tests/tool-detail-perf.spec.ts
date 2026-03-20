/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Tool Detail Performance Optimization', () => {

    const serviceId = `test-service-${Date.now()}`;
    const toolName = 'test-tool';

    test.beforeAll(async ({ request }) => {
        // Seed real service
        await request.post('/api/v1/services', {
            data: {
                name: serviceId,
                command_line_service: {
                    command: "echo test",
                    tools: [
                        {
                            name: toolName,
                            description: 'A test tool',
                            inputSchema: { type: 'object', properties: {} }
                        }
                    ]
                }
            }
        });

        // Execute the tool a few times to generate metrics/stats natively if possible
        for (let i = 0; i < 3; i++) {
            await request.post('/api/v1/execute', {
                data: {
                    service: serviceId,
                    tool: toolName,
                    args: {}
                }
            }).catch(() => {});
        }
    });

    test.afterAll(async ({ request }) => {
        await request.delete(`/api/v1/services/${serviceId}`).catch(() => {});
    });

    test('should load tool details and metrics correctly', async ({ page }) => {
        // Mock gRPC call (failure to force fallback if app attempts gRPC first)
        await page.route('**/*RegistrationService/GetService', async (route) => {
            await route.abort();
        });

        // DO NOT mock /api/v1/services or /status
        // Use the seeded DB

        await page.goto(`/service/${serviceId}/tool/${toolName}`);

        // Verify Tool Name
        await expect(page.getByText(toolName).first()).toBeVisible({ timeout: 15000 });

        // Verify Tool Description
        await expect(page.getByText('A test tool')).toBeVisible();

        // Verify Metrics - it should be 3 since we seeded 3 executions
        await expect(page.getByText('3')).toBeVisible();
    });

    test('should handle missing service gracefully', async ({ page }) => {
        const missingServiceId = 'missing-service-123';
        const missingToolName = 'test-tool';

        await page.route('**/*RegistrationService/GetService', async (route) => {
            await route.abort();
        });

        // Do not mock the REST endpoints, let it 404 naturally

        await page.goto(`/service/${missingServiceId}/tool/${missingToolName}`);

        await expect(page.getByRole('alert').filter({ hasText: /not found|404|error|failed/i })).toBeVisible({ timeout: 10000 });
    });
});
