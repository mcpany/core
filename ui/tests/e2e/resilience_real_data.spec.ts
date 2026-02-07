/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Resilience Configuration', () => {
    test.describe.configure({ mode: 'serial' });

    const serviceName = 'resilience-test-' + Date.now();

    test.beforeAll(async ({ request }) => {
        // Seed the service
        const res = await request.post('/api/v1/services', {
            data: {
                name: serviceName,
                http_service: {
                    address: "http://example.com"
                },
                disable: false
            }
        });
        expect(res.ok()).toBeTruthy();
    });

    test.afterAll(async ({ request }) => {
        // Cleanup
        await request.delete(`/api/v1/services/${serviceName}`);
    });

    test('should allow configuring resilience settings', async ({ page, request }) => {
        await page.goto('/upstream-services');

        // Open Edit Dialog
        const row = page.locator('tr').filter({ hasText: serviceName });
        await row.getByRole('button', { name: 'Open menu' }).click();
        await page.getByRole('menuitem', { name: 'Edit' }).click();

        // Navigate to Advanced
        await page.getByRole('tab', { name: 'Advanced' }).click();

        // Verify Resilience Editor is visible
        await expect(page.getByText('Timeout')).toBeVisible();

        // Set Timeout
        // Using locator for input value directly inside the timeout card
        const timeoutInput = page.locator('input').filter({ hasText: '' }).first();
        // Better: placeholder
        await page.getByPlaceholder('30s').fill('45s');

        // Enable Retry Policy
        // Click the switch in the "Retry Policy" card header
        await page.locator('div.flex.items-center.justify-between').filter({ hasText: 'Retry Policy' }).getByRole('switch').click();

        // Verify inputs appear
        await expect(page.getByText('Number of Retries')).toBeVisible();
        await page.getByLabel('Number of Retries').fill('5');

        // Save
        await page.getByRole('button', { name: 'Save Changes' }).click();

        // Wait for dialog to close
        await expect(page.getByRole('dialog')).toBeHidden();

        // Verify Persistence via API
        const res = await request.get(`/api/v1/services/${serviceName}`);
        expect(res.ok()).toBeTruthy();
        const data = await res.json();
        const service = data.service || data;

        console.log('Service Data:', JSON.stringify(service, null, 2));

        // Check Resilience
        // Note: snake_case expected from backend
        expect(service.resilience).toBeDefined();
        expect(service.resilience.timeout).toBe('45s');
        expect(service.resilience.retry_policy).toBeDefined();
        expect(service.resilience.retry_policy.number_of_retries).toBe(5);
    });
});
