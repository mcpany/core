/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { test, expect } from '@playwright/test';

test.describe('Smart Connection Wizard', () => {
    test('should allow creating a new service via the wizard and verify DB state', async ({ page, request }) => {
        // Navigate to the Upstream Services page
        await page.goto('/upstream-services');

        // Click the "Add Service" button
        await page.getByRole('button', { name: 'Add Service' }).click();

        // Wait for the sheet to animate in
        await expect(page.getByRole('dialog')).toBeVisible();

        // Select the "Wizard Seed Template" template from the available templates list
        await page.locator('h3').filter({ hasText: 'Wizard Seed Template' }).click();

        // Ensure we moved to Step 1: Configuration Form
        await expect(page.getByText('Service Name')).toBeVisible();

        const testServiceName = `wizard-e2e-test-${Date.now()}`;

        // Fill in the Service Name input
        await page.getByLabel('Service Name').fill(testServiceName);

        // Click "Connect" to trigger validation (Step 2 -> 3)
        // Ensure exact match for "Connect" button text to avoid matching other similarly named buttons
        await page.getByRole('button', { name: /^Connect$/ }).click();

        // Wait for validation to succeed and display the success step
        await expect(page.getByText('Connection Successful')).toBeVisible({ timeout: 20000 });

        // Click "Save & Finish"
        await page.getByRole('button', { name: /Save & Finish/i }).click();

        // Assert that the new service is visible in the UI list
        await expect(page.getByText(testServiceName)).toBeVisible({ timeout: 10000 });

        // VERIFY BACKEND STATE (Database Seeding Verification)
        const res = await request.get('/api/v1/services');
        expect(res.ok()).toBeTruthy();

        const data = await res.json();
        const servicesList = Array.isArray(data) ? data : (data.services || []);

        // Assert that our newly created service actually exists in the backend API response
        const createdService = servicesList.find((s: { name: string }) => s.name === testServiceName);
        expect(createdService).toBeDefined();

        // Cleanup after verification
        await request.delete(`/api/v1/services/${testServiceName}`);
    });
});
