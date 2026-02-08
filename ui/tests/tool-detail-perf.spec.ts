/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';
import { seedServices, cleanupServices, seedUser, cleanupUser } from './e2e/test-data';
import { login } from './e2e/auth-helper';

test.describe('Tool Detail Performance Optimization', () => {
    test.beforeEach(async ({ request, page }) => {
        await seedUser(request, "e2e-admin");
        await seedServices(request);
        await login(page);
    });

    test.afterEach(async ({ request }) => {
        await cleanupServices(request);
        await cleanupUser(request, "e2e-admin");
    });

    test('should load tool details and metrics correctly', async ({ page }) => {
        const serviceName = 'Math New';
        const toolName = 'calculator';

        await page.goto('/tools');
        await expect(page.getByText(toolName).first()).toBeVisible({ timeout: 30000 });

        // Navigate to tool detail page
        await page.goto(`/service/${encodeURIComponent(serviceName)}/tool/${toolName}`);

        // Verify Tool Name
        await expect(page.getByRole('heading', { name: toolName })).toBeVisible({ timeout: 30000 });

        // Verify Tool Description
        await expect(page.getByText('Perform basic math')).toBeVisible();

        // Metrics might not be seeded or zero, but we can verify the section exists
        await expect(page.getByRole('heading', { name: 'Usage Metrics' })).toBeVisible();
    });

    test('should handle missing service gracefully', async ({ page }) => {
        const serviceId = 'missing-service';
        const toolName = 'test-tool';

        await page.goto(`/service/${serviceId}/tool/${toolName}`);

        // Expect 404 or Not Found message
        await expect(page.getByText(/not found|404|error|failed/i)).toBeVisible({ timeout: 15000 });
    });
});
