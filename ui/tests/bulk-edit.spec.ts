/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Bulk Edit Services', () => {
    const serviceNames = ['bulk-svc-1', 'bulk-svc-2', 'bulk-svc-3'];

    test.beforeEach(async ({ request }) => {
        // Cleanup first
        for (const name of serviceNames) {
            await request.delete(`/api/v1/services/${name}`, {
                failOnStatusCode: false
            });
        }

        // Seed
        for (const name of serviceNames) {
            await request.post('/api/v1/services', {
                data: {
                    name,
                    id: name,
                    version: '1.0.0',
                    command_line_service: {
                        command: 'echo',
                        working_directory: '/tmp'
                    }
                }
            });
        }
    });

    test.afterEach(async ({ request }) => {
        for (const name of serviceNames) {
            await request.delete(`/api/v1/services/${name}`, {
                 failOnStatusCode: false
            });
        }
    });

    test('should bulk edit timeout and tags', async ({ page, request }) => {
        await page.goto('/upstream-services');

        // Wait for services to appear
        await expect(page.getByText(serviceNames[0])).toBeVisible();

        // Click checkbox for each service
        for (const name of serviceNames) {
             const row = page.getByRole('row').filter({ hasText: name });
             await row.getByRole('checkbox', { name: `Select ${name}` }).click();
        }

        // Click Bulk Edit
        await page.getByRole('button', { name: 'Bulk Edit' }).click();

        // Dialog should be visible
        await expect(page.getByRole('dialog')).toBeVisible();

        // Add Tag
        await page.getByLabel('Add Tags').fill('bulk-updated');

        // Set Timeout
        await page.getByLabel('Timeout').fill('99');

        // Apply
        await page.getByRole('button', { name: 'Apply Changes' }).click();

        // Verify Toast
        await expect(page.getByText('Services Updated')).toBeVisible();

        // Wait a bit for backend to persist/propagate if needed
        await page.waitForTimeout(500);

        // Verify via API
        for (const name of serviceNames) {
            const res = await request.get(`/api/v1/services/${name}`);
            expect(res.ok()).toBeTruthy();
            const data = await res.json();
            const service = data.service || data;

            expect(service.tags).toContain('bulk-updated');
            expect(service.resilience?.timeout).toBe('99s');
        }
    });
});
